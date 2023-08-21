//go:build std || project

package project

import (
	"context"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ProjectCommand struct{}

func (gc *ProjectCommand) Init(cc plug.InitContext) error {
	cc.AddCommandGroup("project", "Project")
	cc.SetCommandGroup("project")
	cc.SetTopLevel(true)
	cc.SetCommandUsage("project [command] [flags]")
	help := "Project commands"
	cc.SetCommandHelp(help, help)
	return nil
}

func (gc ProjectCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	return nil
}

func (qc ProjectCommand) Augment(ec plug.ExecContext, props *plug.Properties) error {
	return nil
}

func init() {
	cmd := &ProjectCommand{}
	Must(plug.Registry.RegisterCommand("project", cmd))
	plug.Registry.RegisterAugmentor("20-project", cmd)
}
