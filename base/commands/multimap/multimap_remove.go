//go:build std || multimap

package multimap

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

func init() {
	c := commands.NewMapRemoveCommand("MultiMap", codec.EncodeMultiMapRemoveRequest, makeDecodeResponseRowsFunc(codec.DecodeMultiMapRemoveResponse))
	check.Must(plug.Registry.RegisterCommand("multi-map:remove", c))
}
