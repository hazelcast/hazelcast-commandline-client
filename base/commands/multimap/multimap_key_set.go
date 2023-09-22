//go:build std || multimap

package multimap

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

func init() {
	c := commands.NewMapKeySetCommand("MultiMap", "multimap", codec.EncodeMultiMapKeySetRequest, codec.DecodeMultiMapKeySetResponse)
	check.Must(plug.Registry.RegisterCommand("multi-map:key-set", c))
}
