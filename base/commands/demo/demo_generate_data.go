//go:build std || demo

package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/sql"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/demo"
	"github.com/hazelcast/hazelcast-commandline-client/internal/demo/wikimedia"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type DataStreamGenerator interface {
	Stream(ctx context.Context) chan demo.StreamItem
	MappingQuery(mapName string) (string, error)
}

var supportedEventStreams = map[string]DataStreamGenerator{
	"wikipedia-event-stream": wikimedia.StreamGenerator{},
}

type GenerateDataCmd struct{}

func (cm GenerateDataCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("generate-data [name] [key=value, ...] [--preview]")
	long := `Generates stream events
	
Generate data for given name, supported names are:

- wikipedia-event-stream: Real-time Wikipedia event stream. Following key-value pairs can be set
	- map-name=<MAP-NAME>: generated stream items are written into the map

`
	short := "Generates stream events"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, math.MaxInt)
	cc.AddIntFlag(flagMaxValues, "", 0, false, "number of events to create")
	cc.AddBoolFlag(flagPreview, "", false, false, "print the generated data without interacting with the cluster")
	return nil
}

func (cm GenerateDataCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Args()[0]
	generator, ok := supportedEventStreams[name]
	if !ok {
		return fmt.Errorf("Stream generator '%s' is not supported, run --help to see supported ones", name)
	}
	keyVals, err := keyValMap(ec)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()
	chv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Generating wikipedia stream events: %s", name))
		ch := generator.Stream(ctx)
		return ch, nil

	})
	if err != nil {
		return err
	}
	defer stop()
	ch := chv.(chan demo.StreamItem)
	preview := ec.Props().GetBool(flagPreview)
	if preview {
		return generatePreviewResult(ctx, ec, generator, ch, keyVals)
	}
	return generateResult(ctx, ec, generator, ch, keyVals)
}

func generatePreviewResult(ctx context.Context, ec plug.ExecContext, generator DataStreamGenerator, itemCh <-chan demo.StreamItem, keyVals map[string]string) error {
	outCh := make(chan output.Row)
	maxCount := ec.Props().GetInt(flagMaxValues)
	if maxCount < 1 {
		maxCount = 10
	}
	createdCount := 0
	mn := keyVals[pairMapName]
	mq, err := generator.MappingQuery(mn)
	if err != nil {
		return err
	}
	ec.PrintlnUnnecessary(fmt.Sprintf("Following mapping will be created when run without preview:\n%s", mq))
	ec.PrintlnUnnecessary("Generating preview items...")
	go func() {
	loop:
		for createdCount < int(maxCount) {
			var ev demo.StreamItem
			select {
			case ev = <-itemCh:
			case <-ctx.Done():
				break loop
			}
			select {
			case outCh <- ev.Row():
			case <-ctx.Done():
				break loop
			}
			createdCount++
		}
		close(outCh)
	}()
	return ec.AddOutputStream(ctx, outCh)
}

func generateResult(ctx context.Context, ec plug.ExecContext, generator DataStreamGenerator, itemCh <-chan demo.StreamItem, keyVals map[string]string) error {
	mv, err := ec.Props().GetBlocking(demoMapPropertyName)
	if err != nil {
		return err
	}
	m := mv.(*hazelcast.Map)
	err = runMappingQuery(ctx, ec, m, wikimedia.StreamGenerator{})
	outCh := make(chan output.Row)
	errCh := make(chan error, 1)
	maxCount := ec.Props().GetInt(flagMaxValues)
	createdCount := 0
	ec.PrintlnUnnecessary(fmt.Sprintf(`Run the following SQL query to see the generated data
	
	SELECT
	__key, meta_dt as "timestamp", user_, comment
	FROM "%s"
	LIMIT 10;
	
Generating event stream...
`, m.Name()))
	go func() {
	loop:
		for {
			var ev demo.StreamItem
			select {
			case ev = <-itemCh:
			case <-ctx.Done():
				break loop
			}
			fm := ev.FlatMap()
			b, err := json.Marshal(fm)
			if err != nil {
				ec.Logger().Warn("Could not marshall stream item: %s", err.Error())
				continue
			}
			_, err = m.Put(ctx, ev.ID(), serialization.JSON(b))
			if err != nil {
				ec.Logger().Warn("Could not put stream item into map %s: %s", m.Name(), err.Error())
				continue
			}
			createdCount++
			if maxCount > 0 && createdCount == int(maxCount) {
				errCh <- nil
				break
			}
		}
		close(outCh)
	}()
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func runMappingQuery(ctx context.Context, ec plug.ExecContext, m *hazelcast.Map, generator DataStreamGenerator) error {
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Creating mapping for map: %s", m.Name()))
		q, err := generator.MappingQuery(m.Name())
		if err != nil {
			return nil, err
		}
		_, cancel, err := sql.ExecSQL(ctx, ec, q)
		if err != nil {
			return nil, err
		}
		cancel()
		return nil, nil
	})
	stop()
	return err
}

func init() {
	Must(plug.Registry.RegisterCommand("demo:generate-data", &GenerateDataCmd{}))
}
