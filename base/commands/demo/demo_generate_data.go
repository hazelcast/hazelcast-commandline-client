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
	preview := ec.Props().GetBool(flagPreview)
	if preview {
		return generatePreviewResult(ctx, ec, generator, kvs.Map())
	}
	return generateResult(ctx, ec, generator, kvs.Map())
}

func generatePreviewResult(ctx context.Context, ec plug.ExecContext, generator DataStreamGenerator, keyVals map[string]string) error {
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
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("")
		itemCh, stopStream := generator.Stream(ctx)
		defer stopStream()
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
		return nil, ec.AddOutputStream(ctx, outCh)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func generateResult(ctx context.Context, ec plug.ExecContext, generator DataStreamGenerator, keyVals map[string]string) error {
	mapName, ok := keyVals[pairMapName]
	if !ok {
		return fmt.Errorf("either %s key-value pair must be given or --preview must be used", pairMapName)
	}
	m, err := getMap(ctx, ec, mapName)
	if err != nil {
		return err
	}
	maxCount := ec.Props().GetInt(flagMaxValues)
	errCh := make(chan error)
	itemCh, stopStream := generator.Stream(ctx)
	defer stopStream()
	var count int64
	go feedItems(ctx, ec, m, maxCount, itemCh, errCh, &count)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		query, err := generator.MappingQuery(mapName)
		if err != nil {
			return nil, err
		}
		err = runMappingQuery(ctx, ec, sp, mapName, query)
		if err != nil {
			return nil, err
		}
		ec.PrintlnUnnecessary(fmt.Sprintf("Following mapping is created:\n\n%s", query))
		ec.PrintlnUnnecessary(fmt.Sprintf(`Run the following SQL query to see the generated data
	
	SELECT
	__key, meta_dt as "timestamp", user_name, comment
	FROM "%s"
	LIMIT 10;
	
`, m.Name()))
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
	ec.PrintlnUnnecessary(fmt.Sprintf("OK Generated %d events", atomic.LoadInt64(&count)))
	return err
}

func feedItems(ctx context.Context, ec plug.ExecContext, m *hazelcast.Map, maxCount int64, itemCh <-chan demo.StreamItem, errCh chan<- error, outCount *int64) {
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
		atomic.AddInt64(outCount, 1)
		if maxCount > 0 && atomic.LoadInt64(outCount) == maxCount {
			errCh <- nil
			break
		}
	}
	close(errCh)
}

func runMappingQuery(ctx context.Context, ec plug.ExecContext, sp clc.Spinner, mapName, query string) error {
	sp.SetText(fmt.Sprintf("Creating mapping for map: %s", mapName))
	if _, err := sql.ExecSQL(ctx, ec, sp, query); err != nil {
		return err
	}
	return nil
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
