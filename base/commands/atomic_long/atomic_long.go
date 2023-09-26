//go:build std || atomiclong

package atomiclong

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/base"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Command struct{}

func (Command) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("atomic-long")
	cc.AddCommandGroup(clc.GroupDDSID, clc.GroupDDSTitle)
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.SetTopLevel(true)
	help := "AtomicLong operations"
	cc.SetCommandHelp(help, help)
	cc.AddStringFlag(base.FlagName, "n", base.DefaultName, false, "AtomicLong name")
	return nil
}

func (Command) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	check.Must(plug.Registry.RegisterCommand("atomic-long", &Command{}))
}
