//go:build std || multimap

package multimap

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func init() {
	cmd := commands.NewDestroyCommand("MultiMap", "multimap", getMultiMap)
	check.Must(plug.Registry.RegisterCommand("multi-map:destroy", cmd))
}
