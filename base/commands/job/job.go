//go:build std || job

package job

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Cmd struct{}

func (cm Cmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("job")
	cc.AddCommandGroup(clc.GroupJetID, clc.GroupJetTitle)
	cc.SetCommandGroup(clc.GroupJetID)
	cc.SetTopLevel(true)
	help := "Jet job operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (cm Cmd) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("job", &Cmd{}))
}
