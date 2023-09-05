//go:build std || demo

package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
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
)

const (
	flagPreview           = "preview"
	flagMaxValues         = "max-values"
	pairMapName           = "map"
	argGeneratorName      = "name"
	argTitleGeneratorName = "generator name"
	argKeyValues          = "keyValue"
	argTitleKeyValues     = "key=value"
)

type DataStreamGenerator interface {
	Stream(ctx context.Context) (chan demo.StreamItem, context.CancelFunc)
	MappingQuery(mapName string) (string, error)
}

var supportedEventStreams = map[string]DataStreamGenerator{
	"wikipedia-event-stream": wikimedia.StreamGenerator{},
}

type GenerateDataCmd struct{}

func (cm GenerateDataCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("generate-data")
	long := `Generates a stream of events
	
Generate data for given name, supported names are:

* wikipedia-event-stream: Real-time Wikipedia event stream.
   Following key-value pairs can be set:
	* map=<MAP-NAME>: the target map to update with the generated stream entries.

`
	short := "Generates a stream of events"
	cc.SetCommandHelp(long, short)
	cc.AddIntFlag(flagMaxValues, "", 0, false, "number of events to create (default: 0, no limits)")
	cc.AddBoolFlag(flagPreview, "", false, false, "print the generated data without interacting with the cluster")
	cc.AddStringArg(argGeneratorName, argTitleGeneratorName)
	cc.AddKeyValueSliceArg(argKeyValues, argTitleKeyValues, 0, clc.MaxArgs)
	return nil
}

func (cm GenerateDataCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	name := ec.GetStringArg(argGeneratorName)
	generator, ok := supportedEventStreams[name]
	if !ok {
		return fmt.Errorf("stream generator '%s' is not supported, run --help to see supported ones", name)
	}
	kvs := ec.GetKeyValuesArg(argKeyValues)
	ch, stopStream := generator.Stream(ctx)
	defer stopStream()
	preview := ec.Props().GetBool(flagPreview)
	if preview {
		return generatePreviewResult(ctx, ec, generator, ch, kvs.Map(), stopStream)
	}
	return generateResult(ctx, ec, generator, ch, kvs.Map(), stopStream)
}

func generatePreviewResult(ctx context.Context, ec plug.ExecContext, generator DataStreamGenerator, itemCh <-chan demo.StreamItem, keyVals map[string]string, stopStream context.CancelFunc) error {
	maxCount := ec.Props().GetInt(flagMaxValues)
	if maxCount < 1 {
		maxCount = 10
	}
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
	outCh := make(chan output.Row)
	count := int64(0)
	go func() {
	loop:
		for count < maxCount {
			var ev demo.StreamItem
			select {
			case event, ok := <-itemCh:
				if !ok {
					break loop
				}
				ev = event
			case <-ctx.Done():
				break loop
			}
			select {
			case outCh <- ev.Row():
			case <-ctx.Done():
				break loop
			}
			count++
		}
		close(outCh)
		stopStream()
	}()
	return ec.AddOutputStream(ctx, outCh)
}

func generateResult(ctx context.Context, ec plug.ExecContext, generator DataStreamGenerator, itemCh <-chan demo.StreamItem, keyVals map[string]string, stopStream context.CancelFunc) error {
	mapName, ok := keyVals[pairMapName]
	if !ok {
		return fmt.Errorf("%s key-value pair must be given", pairMapName)
	}
	m, err := getMap(ctx, ec, mapName)
	if err != nil {
		return err
	}
	query, err := generator.MappingQuery(mapName)
	if err != nil {
		return err
	}
	err = runMappingQuery(ctx, ec, mapName, query)
	if err != nil {
		return err
	}
	ec.PrintlnUnnecessary(fmt.Sprintf("Following mapping is created:\n\n%s", query))
	ec.PrintlnUnnecessary(fmt.Sprintf(`Run the following SQL query to see the generated data
	
	SELECT
	__key, meta_dt as "timestamp", user_name, comment
	FROM "%s"
	LIMIT 10;
	
`, m.Name()))
	maxCount := ec.Props().GetInt(flagMaxValues)
	count := int64(0)
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
			case <-ctx.Done():
				errCh <- ctx.Err()
				break loop
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
			atomic.AddInt64(&count, 1)
			if maxCount > 0 && atomic.LoadInt64(&count) == maxCount {
				errCh <- nil
				break
			}
		}
		close(errCh)
	}()
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case err := <-errCh:
				return nil, err
			case <-ticker.C:
				sp.SetText(fmt.Sprintf("Generated %d events", atomic.LoadInt64(&count)))
			}
		}
	})
	stop()
	stopStream()
	ec.PrintlnUnnecessary(fmt.Sprintf("Generated %d events", atomic.LoadInt64(&count)))
	return err
}

func runMappingQuery(ctx context.Context, ec plug.ExecContext, mapName, query string) error {
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Creating mapping for map: %s", mapName))
		_, cancel, err := sql.ExecSQL(ctx, ec, query)
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

func init() {
	Must(plug.Registry.RegisterCommand("demo:generate-data", &GenerateDataCmd{}))
}
