//go:build std || map

package _map

import (
	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/proto/codec"
)

func init() {
	c := commands.NewMapGetCommand[*hazelcast.Map]("Map", "map", codec.EncodeMapGetRequest, makeDecodeResponseRowsFunc(codec.DecodeMapGetResponse), getMap)
	check.Must(plug.Registry.RegisterCommand("map:get", c))
}
