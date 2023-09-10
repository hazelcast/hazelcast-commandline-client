//go:build std || queue

package queue

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/internal"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const flagCount = "count"

type QueuePollCommand struct{}

func (QueuePollCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("poll")
	help := "Remove the given number of elements from the given Queue"
	cc.SetCommandHelp(help, help)
	commands.AddValueTypeFlag(cc)
	cc.AddIntFlag(flagCount, "", 1, false, "number of element to be removed from the given queue")
	return nil
}

func (QueuePollCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	queueName := ec.Props().GetString(base.FlagName)
	count := int(ec.Props().GetInt(flagCount))
	if count < 0 {
		return fmt.Errorf("%s cannot be negative", flagCount)
	}
	rows, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		ci, err := cmd.ClientInternal(ctx, ec, sp)
		if err != nil {
			return nil, err
		}
		sp.SetText(fmt.Sprintf("Polling from Queue '%s'", queueName))
		req := codec.EncodeQueuePollRequest(queueName, 0)
		pID, err := internal.StringToPartitionID(ci, queueName)
		if err != nil {
			return nil, err
		}
		var rows []output.Row
		for i := 0; i < count; i++ {
			rv, err := ci.InvokeOnPartition(ctx, req, pID, nil)
			if err != nil {
				return nil, err
			}
			data := codec.DecodeQueuePollResponse(rv)
			vt := data.Type()
			value, err := ci.DecodeData(data)
			if err != nil {
				ec.Logger().Info("The value was not decoded, due to error: %s", err.Error())
				value = serialization.NondecodedType(serialization.TypeToLabel(vt))
			}
			row := output.Row{
				output.Column{
					Name:  output.NameValue,
					Type:  vt,
					Value: value,
				},
			}
			if ec.Props().GetBool(base.FlagShowType) {
				row = append(row, output.Column{
					Name:  output.NameValueType,
					Type:  serialization.TypeString,
					Value: serialization.TypeToLabel(vt),
				})
			}
			rows = append(rows, row)
		}
		return rows, nil
	})
	if err != nil {
		return err
	}
	stop()
	return ec.AddOutputRows(ctx, rows.([]output.Row)...)
}

func init() {
	Must(plug.Registry.RegisterCommand("queue:poll", &QueuePollCommand{}))
}
