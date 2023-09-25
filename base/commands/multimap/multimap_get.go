//go:build std || multimap

package multimap

import (
	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

func init() {
	d := makeDecodeResponseRowsFunc(codec.DecodeMultiMapGetResponse)
	c := commands.NewMapGetCommand("MultiMap", codec.EncodeMultiMapGetRequest, d, getMultiMap)
	check.Must(plug.Registry.RegisterCommand("multi-map:get", c))
}
