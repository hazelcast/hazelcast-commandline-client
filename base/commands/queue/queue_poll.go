//go:build std || queue

package queue

import (
	"context"
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const flagCount = "count"

type QueuePollCommand struct {
}

func (qc *QueuePollCommand) Init(cc plug.InitContext) error {
	addValueTypeFlag(cc)
	help := "Remove the given number of elements from the given Queue"
	cc.AddIntFlag(flagCount, "", 1, false, "number of element to be removed from the given queue")
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("poll [flags]")
	cc.SetPositionalArgCount(0, 0)
	return nil
}

func (qc *QueuePollCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	queueName := ec.Props().GetString(queueFlagName)
	count := int(ec.Props().GetInt(flagCount))
	if count < 0 {
		return fmt.Errorf("%s cannot be negative", flagCount)
	}
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	rows, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Polling from queue %s", queueName))
		var rows []output.Row
		for i := 0; i < count; i++ {
			req := codec.EncodeQueuePollRequest(queueName, 0)
			pID, err := stringToPartitionID(ci, queueName)
			if err != nil {
				return nil, err
			}
			rv, err := ci.InvokeOnPartition(ctx, req, pID, nil)
			if err != nil {
				return nil, err
			}
			raw := codec.DecodeQueuePollResponse(rv)
			valueType := raw.Type()
			value, err := ci.DecodeData(raw)
			if err != nil {
				ec.Logger().Info("The value was not decoded, due to error: %s", err.Error())
				value = serialization.NondecodedType(serialization.TypeToLabel(valueType))
			}
			row := output.Row{
				output.Column{
					Name:  output.NameValue,
					Type:  valueType,
					Value: value,
				},
			}
			if ec.Props().GetBool(queueFlagShowType) {
				row = append(row, output.Column{
					Name:  output.NameValueType,
					Type:  serialization.TypeString,
					Value: serialization.TypeToLabel(valueType),
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

func stringToPartitionID(ci *hazelcast.ClientInternal, name string) (int32, error) {
	var partitionID int32
	var keyData hazelcast.Data
	var err error
	idx := strings.Index(name, "@")
	if keyData, err = ci.EncodeData(name[idx+1:]); err != nil {
		return 0, err
	}
	if partitionID, err = ci.GetPartitionID(keyData); err != nil {
		return 0, err
	}
	return partitionID, nil
}
