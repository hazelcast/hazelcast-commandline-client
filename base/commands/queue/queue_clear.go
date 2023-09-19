//go:build std || queue

package queue

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func init() {
	c := commands.NewClearCommand("Queue", getQueue)
	check.Must(plug.Registry.RegisterCommand("queue:clear", c))
}
