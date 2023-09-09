//go:build std || queue

package queue

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	queueFlagName     = "name"
	queueFlagShowType = "show-type"
	queuePropertyName = "queue"
)

type QueueCommand struct {
}

func (qc *QueueCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("queue")
	cc.AddCommandGroup(clc.GroupDDSID, clc.GroupDDSTitle)
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.SetTopLevel(true)
	help := "Queue operations"
	cc.SetCommandHelp(help, help)
	cc.AddStringFlag(queueFlagName, "n", defaultQueueName, false, "queue name")
	cc.AddBoolFlag(queueFlagShowType, "", false, false, "add the type names to the output")
	return nil
}

func (qc *QueueCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func (qc *QueueCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	ctx := context.TODO()
	props.SetBlocking(queuePropertyName, func() (any, error) {
		queueName := ec.Props().GetString(queueFlagName)
		// empty queue name is allowed
		ci, err := ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		val, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
			sp.SetText(fmt.Sprintf("Getting queue %s", queueName))
			q, err := ci.Client().GetQueue(ctx, queueName)
			if err != nil {
				return nil, err
			}
			return q, nil
		})
		if err != nil {
			return nil, err
		}
		stop()
		return val.(*hazelcast.Queue), nil
	})
	return nil
}

func init() {
	cmd := &QueueCommand{}
	check.Must(plug.Registry.RegisterCommand("queue", cmd))
	plug.Registry.RegisterAugmentor("20-queue", cmd)
}
