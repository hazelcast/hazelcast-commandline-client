package internal

import "github.com/hazelcast/hazelcast-go-client/v4/hazelcast"

func DefaultConfig() *hazelcast.Config {
	config := hazelcast.NewConfig()
	config.SetProperty("hazelcast.client.logging.level", "error")
	return config
}
