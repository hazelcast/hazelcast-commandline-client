//go:build std || demo

package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"
	"time"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/sql"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/demo"
	"github.com/hazelcast/hazelcast-commandline-client/internal/demo/wikimedia"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/str"
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
	long := `Generates a stream of events
	
Generate data for given name, supported names are:

* wikipedia-event-stream: Real-time Wikipedia event stream.
   Following key-value pairs can be set:
	* map=<MAP-NAME>: the target map to update with the generated stream entries.

`
	short := "Generates a stream of events"
	cc.SetCommandHelp(long, short)
	cc.SetPositionalArgCount(1, math.MaxInt)
	cc.AddIntFlag(flagMaxValues, "", 0, false, "number of events to create (default: 0, no limits)")
	cc.AddBoolFlag(flagPreview, "", false, false, "print the generated data without interacting with the cluster")
	return nil
}

func (cm GenerateDataCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Args()[0]
	generator, ok := supportedEventStreams[name]
	if !ok {
		return fmt.Errorf("stream generator '%s' is not supported, run --help to see supported ones", name)
	}
	keyVals := keyValMap(ec)
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
	count := 0
	mapName := keyVals[pairMapName]
	if mapName == "" {
		mapName = "<map-name>"
	}
	mq, err := generator.MappingQuery(mapName)
	if err != nil {
		return err
	}
	ec.PrintlnUnnecessary(fmt.Sprintf("Following mapping will be created when run without preview:\n\n%s", mq))
	ec.PrintlnUnnecessary("Generating preview items...")
	go func() {
	loop:
		for count < int(maxCount) {
			var ev demo.StreamItem
			select {
			case event, ok := <-itemCh:
				if !ok {
					break loop
				}
				ev = event
			}
			select {
			case outCh <- ev.Row():
			case <-ctx.Done():
				break loop
			}
			count++
		}
		close(outCh)
	}()
	return ec.AddOutputStream(ctx, outCh)
}

func generateResult(ctx context.Context, ec plug.ExecContext, generator DataStreamGenerator, itemCh <-chan demo.StreamItem, keyVals map[string]string) error {
	mapName, ok := keyVals[pairMapName]
	if !ok {
		return fmt.Errorf("%s key-value pair must be given", pairMapName)
	}
	m, err := getMap(ctx, ec, mapName)
	if err != nil {
		return err
	}
	err = runMappingQuery(ctx, ec, m, generator)
	maxCount := ec.Props().GetInt(flagMaxValues)
	count := 0
	ec.PrintlnUnnecessary(fmt.Sprintf(`Run the following SQL query to see the generated data
	
	SELECT
	__key, meta_dt as "timestamp", user_, comment
	FROM "%s"
	LIMIT 10;
	
Generating event stream...
`, m.Name()))
	errCh := make(chan error)
	go func() {
	loop:
		for {
			var ev demo.StreamItem
			select {
			case event, ok := <-itemCh:
				if !ok {
					errCh <- nil
					break loop
				}
				ev = event
			}
			fm := ev.KeyValues()
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
			count++
			if maxCount > 0 && count == int(maxCount) {
				errCh <- nil
				break
			}
		}
	}()
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case err := <-errCh:
				return nil, err
			case <-ticker.C:
				sp.SetText(fmt.Sprintf("Generated %d events", count))
			}
		}
	})
	stop()
	ec.PrintlnUnnecessary(fmt.Sprintf("Generated %d events", count))
	return err
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

func getMap(ctx context.Context, ec plug.ExecContext, mapName string) (*hazelcast.Map, error) {
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return nil, err
	}
	mv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Getting map %s", mapName))
		m, err := ci.Client().GetMap(ctx, mapName)
		if err != nil {
			return nil, err
		}
		return m, nil
	})
	if err != nil {
		return nil, err
	}
	stop()
	return mv.(*hazelcast.Map), nil
}

func keyValMap(ec plug.ExecContext) map[string]string {
	keyVals := map[string]string{}
	for _, keyval := range ec.Args()[1:] {
		k, v := str.ParseKeyValue(keyval)
		if k == "" {
			continue
		}
		keyVals[k] = v
	}
	return keyVals
}
func init() {
	Must(plug.Registry.RegisterCommand("demo:generate-data", &GenerateDataCmd{}))
}
