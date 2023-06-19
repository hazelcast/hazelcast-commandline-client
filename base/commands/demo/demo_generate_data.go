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

	"github.com/hazelcast/hazelcast-commandline-client/base/commands/sql"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
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
	cc.AddBoolFlag(flagPreview, "", false, false, "print the generated data")
	return nil
}

func (cm GenerateDataCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.Args()[0]
	generator, ok := supportedEventStreams[name]
	if !ok {
		return fmt.Errorf("Stream generator '%s' is not supported, run --help to see supported ones", name)
	}
	mv, err := ec.Props().GetBlocking(demoMapPropertyName)
	if err != nil {
		return err
	}
	m := mv.(*hazelcast.Map)
	ctx, newStop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer newStop()
	chv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Creating mapping: %s", name))
		q, err := generator.MappingQuery(m.Name())
		if err != nil {
			return nil, err
		}
		_, cancel, err := sql.ExecSQL(ctx, ec, q)
		if err != nil {
			return nil, err
		}
		defer cancel()
		sp.SetText(fmt.Sprintf("Generating wikipedia stream events: %s", name))
		ch := generator.Stream(ctx)
		return ch, nil

	})
	if err != nil {
		return err
	}
	defer stop()
	ec.PrintlnUnnecessary(fmt.Sprintf(`Run the following SQL query to see the generated data
	
	SELECT
	id, meta_id, user_
	FROM %s;
	
Generating event stream...
`, m.Name()))
	return generateResult(ctx, ec, m, chv.(chan demo.StreamItem))
}

func generateResult(ctx context.Context, ec plug.ExecContext, m *hazelcast.Map, itemCh <-chan demo.StreamItem) error {
	outCh := make(chan output.Row)
	defer close(outCh)
	preview := ec.Props().GetBool(flagPreview)
	if preview {
		go ec.AddOutputStream(ctx, outCh)
	}
	for {
		var ev demo.StreamItem
		select {
		case ev = <-itemCh:
		case <-ctx.Done():
			return ctx.Err()
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
		if preview {
			select {
			case outCh <- ev.Row():
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func init() {
	Must(plug.Registry.RegisterCommand("demo:generate-data", &GenerateDataCmd{}))
}
