//go:build std || map

package _map

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func init() {
	c := commands.NewLockCommand("Map", "map", getMap)
	check.Must(plug.Registry.RegisterCommand("map:lock", c, plug.OnlyInteractive{}))
}
