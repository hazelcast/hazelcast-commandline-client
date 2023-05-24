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

type QueuePollCommand struct {
}

func (qc *QueuePollCommand) Init(cc plug.InitContext) error {
	addValueTypeFlag(cc)
	help := "Remove the given number of elements from the given Queue"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("poll [-n queue-name] [flags]")
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (qc *QueuePollCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	queueName := ec.Props().GetString(queueFlagName)
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	req := codec.EncodeQueuePollRequest(ci, queueName, 0)
	pID, err := stringToPartitionID(ci, queueName)
	if err != nil {
		return err
	}
	rv, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Polling from queue %s", queueName))
		return ci.InvokeOnPartition(ctx, req, pID, nil)
	})
	if err != nil {
		return err
	}
	stop()
	raw := codec.DecodeMapRemoveResponse(rv.(*hazelcast.ClientMessage))
	vt := raw.Type()
	value, err := ci.DecodeData(raw)
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
	if ec.Props().GetBool(queueFlagShowType) {
		row = append(row, output.Column{
			Name:  output.NameValueType,
			Type:  serialization.TypeString,
			Value: serialization.TypeToLabel(vt),
		})
	}
	return ec.AddOutputRows(ctx, row)
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
