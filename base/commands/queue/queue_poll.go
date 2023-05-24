//go:build base || queue

package _queue

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
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
	qv, err := ec.Props().GetBlocking(queuePropertyName)
	if err != nil {
		return err
	}
	q := qv.(*hazelcast.Queue)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Polling from queue %s", queueName))
		return q.Poll(ctx)
	})
	if err != nil {
		return err
	}
	stop()
	//TODO: Decode Response? Yes, return the value polled.
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("queue:poll", &QueuePollCommand{}))
}
