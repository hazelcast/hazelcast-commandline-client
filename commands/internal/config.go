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
	cloudToken, err := flags.GetString("cloud-token")
	if err != nil {
		return nil, err
	}
	if cloudToken != "" {
		config.Cluster.Cloud.Token = cloudToken
		config.Cluster.Cloud.Enabled = true
	} else {
		addrRaw, err := flags.GetString("address")
		if err != nil {
			return nil, err
		}
		if addrRaw != "" {
			addresses := strings.Split(addrRaw, ",")
			config.Cluster.Network.Addresses = addresses
		}
	}
	cluster, err := flags.GetString("cluster-name")
	if err != nil {
		return nil, err
	}
	config.Cluster.Name = cluster
	return config, nil
}
