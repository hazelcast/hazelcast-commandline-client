module github.com/hazelcast/hzc

go 1.15

require (
	github.com/hazelcast/hazelcast-go-client/v4 v4.0.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
)

replace github.com/hazelcast/hazelcast-go-client/v4 => github.com/yuce/hazelcast-go-client/v4 v4.0.0-dev.2
