//go:build base || cluster

package cluster

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ClusterCommand struct {
}

func (mc *ClusterCommand) Init(cc plug.InitContext) error {
	if !cc.Interactive() {
		cc.AddStringFlag(clc.PropertySchemaDir, "", paths.Schemas(), false, "set the schema directory")
	}
	cc.SetTopLevel(true)
	cc.SetCommandUsage("cluster [command] [flags]")
	help := "Cluster operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *ClusterCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("cluster", &ClusterCommand{}))
}
