//go:build std || job

package job

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Command struct{}

func (Command) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("job")
	cc.AddCommandGroup(clc.GroupJetID, clc.GroupJetTitle)
	cc.SetCommandGroup(clc.GroupJetID)
	cc.SetTopLevel(true)
	help := "Jet job operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (Command) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("job", &Command{}))
}
