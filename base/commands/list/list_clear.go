//go:build std || list

package list

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func init() {
	cmd := commands.NewClearCommand("List", getList)
	check.Must(plug.Registry.RegisterCommand("list:clear", cmd))
}
