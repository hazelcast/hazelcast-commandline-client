package internal

import (
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
)

func DefaultConfig() *hazelcast.Config {
	config := hazelcast.NewConfig()
	config.Logger.Level = logger.ErrorLevel
	return &config
}
