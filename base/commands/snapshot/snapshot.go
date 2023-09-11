//go:build std || snapshot

package snapshot

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Command struct{}

func (cm Command) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("snapshot")
	cc.AddCommandGroup(clc.GroupJetID, clc.GroupJetTitle)
	cc.SetCommandGroup(clc.GroupJetID)
	cc.SetTopLevel(true)
	help := "Jet snapshot operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm Command) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("snapshot", &Command{}))
}
