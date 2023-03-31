//go:build base || map

package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type MapClearCommand struct{}

func (mc *MapClearCommand) Init(cc plug.InitContext) error {
	help := "Delete all entries of a Map"
	cc.SetCommandHelp(help, help)
	cc.SetCommandUsage("clear [-n MAP] [flags]")
	return nil
}

func (mc *MapClearCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mv, err := ec.Props().GetBlocking(mapPropertyName)
	if err != nil {
		return err
	}
	m := mv.(*hazelcast.Map)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Clearing map %s", m.Name()))
		if err := m.Clear(ctx); err != nil {
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
	Must(plug.Registry.RegisterCommand("map:clear", &MapClearCommand{}))
}
