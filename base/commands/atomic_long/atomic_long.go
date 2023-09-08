//go:build std || atomiclong

package atomiclong

import (
	"context"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	flagName = "name"
)

type AtomicLongCommand struct{}

func (AtomicLongCommand) Init(cc plug.InitContext) error {
	cc.AddCommandGroup(clc.GroupDDSID, clc.GroupDDSTitle)
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.AddStringFlag(flagName, "n", defaultAtomicLongName, false, "atomic long name")
	cc.SetTopLevel(true)
	cc.SetCommandUsage("atomic-long")
	help := "Atomic long operations"
	cc.SetCommandHelp(help, help)
	return nil
}

func (mc *AtomicLongCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("atomic-long", &AtomicLongCommand{}))
}
