//go:build std || list

package list

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func init() {
	cmd := commands.NewDestroyCommand("List", getList)
	check.Must(plug.Registry.RegisterCommand("list:destroy", cmd))
}
