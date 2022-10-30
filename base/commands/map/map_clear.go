package _map

import (
	"context"
	"fmt"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type MapClearCommand struct{}

func (mc *MapClearCommand) Init(cc plug.InitContext) error {
	help := "Delete all entries of an IMap"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *MapClearCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	mv, err := ec.Props().GetBlocking(mapPropertyName)
	if err != nil {
		return err
	}
	m := mv.(*hazelcast.Map)
	hint := fmt.Sprintf("Clearing map %s", m.Name())
	_, err = ec.ExecuteBlocking(ctx, hint, func(ctx context.Context) (any, error) {
		if err := m.Clear(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("map:clear", &MapClearCommand{}))
}
