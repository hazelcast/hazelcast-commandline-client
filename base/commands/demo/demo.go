//go:build std || demo

package demo

import (
	"context"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Cmd struct{}

func (cm *Cmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("demo")
	cc.AddCommandGroup(GroupDemoID, "Demonstrations")
	cc.SetCommandGroup(GroupDemoID)
	cc.SetTopLevel(true)
	help := "Demonstration commands"
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm *Cmd) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("demo", &Cmd{}))
}
