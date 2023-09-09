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
	cc.SetCommandUsage("atomic-long")
	cc.AddCommandGroup(clc.GroupDDSID, clc.GroupDDSTitle)
	cc.SetCommandGroup(clc.GroupDDSID)
	cc.SetTopLevel(true)
	help := "Atomic long operations"
	cc.SetCommandHelp(help, help)
	cc.AddStringFlag(flagName, "n", defaultAtomicLongName, false, "atomic long name")
	return nil
}

func (mc *AtomicLongCommand) Exec(context.Context, plug.ExecContext) error {
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("atomic-long", &AtomicLongCommand{}))
}
