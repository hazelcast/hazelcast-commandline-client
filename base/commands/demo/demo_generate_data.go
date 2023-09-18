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
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	metric "github.com/hazelcast/hazelcast-commandline-client/clc/metrics"
	"github.com/hazelcast/hazelcast-commandline-client/clc/sql"
	hzerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
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

type GenerateDataCommand struct{}

func (GenerateDataCommand) Init(cc plug.InitContext) error {
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

func (GenerateDataCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
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

func generatePreviewResult(ctx context.Context, ec plug.ExecContext, generator dataStreamGenerator, keyVals map[string]string) error {
	ec.Metrics().Increment(metric.NewSimpleKey(), "total.demo."+cmd.RunningMode(ec))
	maxCount := ec.Props().GetInt(flagMaxValues)
	if maxCount < 1 {
		maxCount = 10
	}
	mapName := keyVals[pairMapName]
	if mapName == "" {
		mapName = "<map-name>"
	}
	mq, err := generator.GenerateMappingQuery(mapName)
	if err != nil {
		return err
	}
	itemCh, stopStream := generator.Stream(ctx)
	defer stopStream()
	ec.PrintlnUnnecessary(fmt.Sprintf("Following mapping will be created when run without preview:\n\n%s", mq))
	_, stop, err := cmd.ExecuteBlocking(ctx, ec, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText("")
		outCh := make(chan output.Row)
		go feedPreviewItems(ctx, maxCount, outCh, itemCh)
		return nil, ec.AddOutputStream(ctx, outCh)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func generateResult(ctx context.Context, ec plug.ExecContext, generator dataStreamGenerator, keyVals map[string]string) error {
	mapName, ok := keyVals[pairMapName]
	if !ok {
		return fmt.Errorf("either %s key-value pair must be given or --preview must be used", pairMapName)
	}
	maxCount := ec.Props().GetInt(flagMaxValues)
	query, err := generator.GenerateMappingQuery(mapName)
	if err != nil {
		return err
	}
	query, stop, err := cmd.ExecuteBlocking(ctx, ec, func(ctx context.Context, sp clc.Spinner) (string, error) {
		sp.SetText("Creating the mapping")
		if _, err := sql.ExecSQL(ctx, ec, query); err != nil {
			return "", err
		}
		cid, vid := cmd.FindClusterIDs(ctx, ec)
		ec.Metrics().Increment(metric.NewKey(cid, vid), "total.demo."+cmd.RunningMode(ec))
		return query, nil
	})
	if err != nil {
		return err
	}
	stop()
	ec.PrintlnUnnecessary(fmt.Sprintf("OK Following mapping is created:\n\n%s", query))
	ec.PrintlnUnnecessary(fmt.Sprintf(`Run the following SQL query to see the generated data
	
	SELECT
	__key, meta_dt as "timestamp", user_name, comment
	FROM "%s"
	LIMIT 10;
	
`, mapName))
	var count int64
	_, stop, err = cmd.ExecuteBlocking(ctx, ec, func(ctx context.Context, sp clc.Spinner) (any, error) {
		errCh := make(chan error)
		itemCh, stopStream := generator.Stream(ctx)
		defer stopStream()
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return 0, err
		}
		m, err := ci.Client().GetMap(ctx, mapName)
		if err != nil {
			return 0, err
		}
		go feedResultItems(ctx, ec, m, maxCount, itemCh, errCh, &count)
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
	if err != nil {
		if !hzerrors.IsUserCancelled(err) && !hzerrors.IsTimeout(err) {
			return err
		}
	}
	stop()
	msg := fmt.Sprintf("OK Generated %d events.", atomic.LoadInt64(&count))
	ec.PrintlnUnnecessary(msg)
	return nil
}

func feedResultItems(ctx context.Context, ec plug.ExecContext, m *hazelcast.Map, maxCount int64, itemCh <-chan demo.StreamItem, errCh chan<- error, outCount *int64) {
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

func feedPreviewItems(ctx context.Context, maxCount int64, outCh chan<- output.Row, itemCh <-chan demo.StreamItem) {
	var count int64
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
}

type dataStreamGenerator interface {
	Stream(ctx context.Context) (chan demo.StreamItem, context.CancelFunc)
	GenerateMappingQuery(mapName string) (string, error)
}

var supportedEventStreams = map[string]dataStreamGenerator{
	"wikipedia-event-stream": wikimedia.StreamGenerator{},
}

func init() {
	check.Must(plug.Registry.RegisterCommand("demo:generate-data", &GenerateDataCommand{}))
}
