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

type QueueOfferCommand struct {
}

func (qc *QueueOfferCommand) Init(cc plug.InitContext) error {
	addValueTypeFlag(cc)
	cc.SetPositionalArgCount(1, 1)
	help := "Add a value to the given Queue"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("offer [value] [flags]")
	return nil
}

func (qc *QueueOfferCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	queueName := ec.Props().GetString(queueFlagName)
	qv, err := ec.Props().GetBlocking(queuePropertyName)
	if err != nil {
		return err
	}
	ci, err := ec.ClientInternal(ctx)
	if err != nil {
		return err
	}
	valStr := ec.Args()[0]
	vd, err := makeValueData(ec, ci, valStr)
	if err != nil {
		return err
	}
	q := qv.(*hazelcast.Queue)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Adding value into queue %s", queueName))
		return q.Add(ctx, vd)
	})
	if err != nil {
		return nil
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("queue:offer", &QueueOfferCommand{}))
}
