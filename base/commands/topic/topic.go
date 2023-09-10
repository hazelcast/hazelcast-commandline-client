//go:build std || topic

package topic

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type TopicCommand struct{}

func (mc *TopicCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("topic")
	cc.AddCommandGroup(clc.GroupDDSID, clc.GroupDDSTitle)
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.SetTopLevel(true)
	help := "Topic operations"
	cc.SetCommandHelp(help, help)
	cc.AddStringFlag(base.FlagName, "n", base.DefaultName, false, "topic name")
	cc.AddBoolFlag(base.FlagShowType, "", false, false, "add the type names to the output")
	return nil
}

func (tc *TopicCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	cmd := &TopicCommand{}
	Must(plug.Registry.RegisterCommand("topic", cmd))
}
