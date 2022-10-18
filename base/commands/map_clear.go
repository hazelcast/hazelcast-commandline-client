package commands

import (
	"context"

	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
)

type MapClearCommand struct{}

func (mc *MapClearCommand) Init(cc plug.InitContext) error {
	help := "Delete all entries of an IMap"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *MapClearCommand) Exec(ec plug.ExecContext) error {
	ctx := context.TODO()
	m := check.MustValue(ec.Props().GetBlocking(mapPropertyName)).(*hazelcast.Map)
	return m.Clear(ctx)
}

func init() {
	Must(plug.Registry.RegisterCommand("map:clear", &MapClearCommand{}))
}
