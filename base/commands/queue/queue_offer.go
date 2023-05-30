//go:build base || queue

package _queue

import (
	"context"
	"fmt"
	"math"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
)

type QueueOfferCommand struct {
}

func (qc *QueueOfferCommand) Init(cc plug.InitContext) error {
	addValueTypeFlag(cc)
	cc.SetPositionalArgCount(1, math.MaxInt)
	help := "Add values to the given Queue"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("[values] [flags]")
	return nil
}

func (qc *QueueOfferCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	queueName := ec.Props().GetString(queueFlagName)
	qv, err := ec.Props().GetBlocking(queuePropertyName)
	if err != nil {
		return err
	}
	for _, arg := range ec.Args() {
		ci, err := ec.ClientInternal(ctx)
		if err != nil {
			return err
		}
		vd, err := makeValueData(ec, ci, arg)
		if err != nil {
			return err
		}
		q := qv.(*hazelcast.Queue)
		_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
			sp.SetText(fmt.Sprintf("Adding values into queue %s", queueName))
			return q.Add(ctx, vd)
		})
		if err != nil {
			return nil
		}
		stop()
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("queue:offer", &QueueOfferCommand{}))
}
