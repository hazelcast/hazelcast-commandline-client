//go:build std || atomiclong

package atomiclong

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func init() {
	c := commands.NewDestroyCommand("AtomicLong", "atomiclong", getAtomicLong)
	check.Must(plug.Registry.RegisterCommand("atomic-long:destroy", c))
}
