//go:build std || serializer

package serializer

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type SerializerCommand struct{}

func (sc *SerializerCommand) Init(cc plug.InitContext) error {
	cc.SetTopLevel(true)
	cc.SetCommandUsage("serializer [command] [flags]")
	help := "Serialization commands"
	cc.SetCommandHelp(help, help)
	return nil
}

func (sc *SerializerCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("serializer", &SerializerCommand{}))
}
