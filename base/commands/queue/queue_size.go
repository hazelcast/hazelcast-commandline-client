//go:build std || queue

package queue

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func init() {
	c := commands.NewSizeCommand("Queue", "queue", getQueue)
	check.Must(plug.Registry.RegisterCommand("queue:size", c))
}
