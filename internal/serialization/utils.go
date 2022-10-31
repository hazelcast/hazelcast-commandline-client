package serialization

import (
	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
)

func UpdateSerializationConfigWithRecursivePaths(cfg *hazelcast.Config, lg log.Logger, paths ...string) error {
	// TODO: in the next PR
	return nil
}
