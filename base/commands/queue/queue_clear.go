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

type QueueClearCommand struct{}

func (qc *QueueClearCommand) Init(cc plug.InitContext) error {
	help := "Delete all entries of a Queue"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("clear")
	return nil
}

func (qc *QueueClearCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	qv, err := ec.Props().GetBlocking(queuePropertyName)
	if err != nil {
		return err
	}
	q := qv.(*hazelcast.Queue)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Clearing queue %s", q.Name()))
		if err = q.Clear(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("queue:clear", &QueueClearCommand{}))
}
