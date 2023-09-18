//go:build std || multimap

package multimap

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func init() {
	c := commands.NewTryLockCommand("MultiMap", getMultiMap)
	check.Must(plug.Registry.RegisterCommand("multi-map:try-lock", c, plug.OnlyInteractive{}))
}
