//go:build std || demo

package demo

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Command struct{}

func (Command) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("demo")
	cc.AddCommandGroup(GroupDemoID, "Demonstrations")
	cc.SetCommandGroup(GroupDemoID)
	cc.SetTopLevel(true)
	help := "Demonstration commands"
	cc.SetCommandHelp(help, help)
	return nil
}

func (Command) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("demo", &Command{}))
}
