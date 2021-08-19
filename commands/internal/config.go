package internal

import (
	"strings"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
	"github.com/spf13/cobra"
)

func DefaultConfig() *hazelcast.Config {
	config := hazelcast.NewConfig()
	config.Logger.Level = logger.ErrorLevel
	return &config
}

func MakeConfig(cmd *cobra.Command) (*hazelcast.Config, error) {
	flags := cmd.InheritedFlags()
	config := DefaultConfig()
	token, err := flags.GetString("cloud-token")
	if err != nil {
		return nil, err
	}
	if token != "" {
		config.Cluster.Cloud.Token = strings.TrimSpace(token)
		config.Cluster.Cloud.Enabled = true
	}
	addrRaw, err := flags.GetString("address")
	if err != nil {
		return nil, err
	}
	if addrRaw != "" {
		addresses := strings.Split(strings.TrimSpace(addrRaw), ",")
		config.Cluster.Network.Addresses = addresses
	}
	cluster, err := flags.GetString("cluster-name")
	if err != nil {
		return nil, err
	}
	config.Cluster.Name = strings.TrimSpace(cluster)
	return config, nil
}
