package topic

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	"github.com/hazelcast/hazelcast-commandline-client/internal/mk"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-commandline-client/internal/topic"
)

func addValueTypeFlag(cc plug.InitContext) {
	help := fmt.Sprintf("value type (one of: %s)", strings.Join(internal.SupportedTypeNames, ", "))
	cc.AddStringFlag(topicFlagValueType, "v", "string", false, help)
}

func makeValueData(ec plug.ExecContext, ci *hazelcast.ClientInternal, valueStr string) (hazelcast.Data, error) {
	vt := ec.Props().GetString(topicFlagValueType)
	if vt == "" {
		vt = "string"
	}
	value, err := mk.ValueFromString(valueStr, vt)
	if err != nil {
		return nil, err
	}
	return ci.EncodeData(value)
}

func updateOutput(ctx context.Context, ec plug.ExecContext, events <-chan topic.TopicEvent) error {
	wantedCount := ec.Props().GetInt(topicFlagCount)
	printedCount := 0
	rowCh := make(chan output.Row)
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()
	ec.PrintlnUnnecessary(fmt.Sprintf("Listening for messages of topic %s", ec.Props().GetString(topicFlagName)))
	go func() {
	loop:
		for {
			var e topic.TopicEvent
			select {
			case e = <-events:
			case <-ctx.Done():
				break loop
			}
			row := eventRow(e, ec)
			select {
			case rowCh <- row:
			case <-ctx.Done():
				break loop
			}
			printedCount++
			if wantedCount > 0 && printedCount == int(wantedCount) {
				break loop
			}
		}
		close(rowCh)
	}()
	return ec.AddOutputStream(ctx, rowCh)
}

func eventRow(e topic.TopicEvent, ec plug.ExecContext) output.Row {
	row := output.Row{
		output.Column{
			Name:  "Value",
			Type:  e.ValueType,
			Value: e.Value,
		},
	}
	if ec.Props().GetBool(topicFlagShowType) {
		row = append(row,
			output.Column{
				Name:  "Type",
				Type:  serialization.TypeString,
				Value: serialization.TypeToLabel(e.ValueType),
			})
	}
	if ec.Props().GetBool(clc.PropertyVerbose) {
		row = append(row,
			output.Column{
				Name:  "Time",
				Type:  serialization.TypeJavaLocalDateTime,
				Value: e.PublishTime,
			},
			output.Column{
				Name:  "Topic",
				Type:  serialization.TypeString,
				Value: e.TopicName,
			},
			output.Column{
				Name:  "Member UUID",
				Type:  serialization.TypeUUID,
				Value: e.Member.UUID,
			},
			output.Column{
				Name:  "Member Address",
				Type:  serialization.TypeString,
				Value: string(e.Member.Address),
			})
	}
	return row
}
