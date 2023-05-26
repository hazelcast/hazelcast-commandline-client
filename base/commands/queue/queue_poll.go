//go:build base || queue

package _queue

import (
	"context"
	"fmt"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-go-client"
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
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	var rows []output.Row
	for i := 0; i < count; i++ {
		valueType, value, err := qc.poll(ctx, ec, ci, queueName, err)
		if err != nil {
			return err
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
	return ec.AddOutputRows(ctx, rows...)
}

func (qc *QueuePollCommand) poll(ctx context.Context, ec plug.ExecContext, ci *hazelcast.ClientInternal, queueName string, err error) (int32, interface{}, error) {
	req := codec.EncodeQueuePollRequest(ci, queueName, 0)
	pID, err := stringToPartitionID(ci, queueName)
	if err != nil {
		return 0, nil, err
	}
	rv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Polling from queue %s", queueName))
		return ci.InvokeOnPartition(ctx, req, pID, nil)
	})
	if err != nil {
		return 0, nil, err
	}
	stop()
	raw := codec.DecodeMapRemoveResponse(rv.(*hazelcast.ClientMessage))
	vt := raw.Type()
	value, err := ci.DecodeData(raw)
	if err != nil {
		ec.Logger().Info("The value was not decoded, due to error: %s", err.Error())
		value = serialization.NondecodedType(serialization.TypeToLabel(vt))
	}
	return vt, value, nil
}

func init() {
	Must(plug.Registry.RegisterCommand("queue:poll", &QueuePollCommand{}))
}

func stringToPartitionID(ci *hazelcast.ClientInternal, name string) (int32, error) {
	idx := strings.Index(name, "@")
	if keyData, err := ci.EncodeData(name[idx+1:]); err != nil {
		return 0, err
	} else if partitionID, err := ci.GetPartitionID(keyData); err != nil {
		return 0, err
	} else {
		return partitionID, nil
	}
}
