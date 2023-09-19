//go:build std || project

package project

import (
	"context"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Command struct{}

func (Command) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("project")
	cc.AddCommandGroup(groupProject, "Project")
	cc.SetCommandGroup(groupProject)
	cc.SetTopLevel(true)
	help := "Project commands"
	cc.SetCommandHelp(help, help)
	return nil
}

func (gc Command) Exec(ctx context.Context, ec plug.ExecContext) error {
	return nil
}

func init() {
	cmd := &Command{}
	Must(plug.Registry.RegisterCommand("project", cmd))
}
