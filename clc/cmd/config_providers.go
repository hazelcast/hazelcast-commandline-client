package cmd

import (
	"github.com/hazelcast/hazelcast-go-client"

	"github.com/hazelcast/hazelcast-commandline-client/clc/config"
	"github.com/hazelcast/hazelcast-commandline-client/clc/logger"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type SimpleConfigProvider struct {
	Props  plug.ReadOnlyProperties
	Logger *logger.Logger
}

func (s SimpleConfigProvider) ClientConfig() (hazelcast.Config, error) {
	return config.MakeHzConfig(s.Props, s.Logger)
}
