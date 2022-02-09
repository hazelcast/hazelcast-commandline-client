module github.com/hazelcast/hazelcast-commandline-client

go 1.15

require (
	github.com/alecthomas/chroma v0.9.2
	github.com/c-bata/go-prompt v0.2.5
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/hazelcast/hazelcast-go-client v1.2.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/hazelcast/hazelcast-go-client v1.2.0 => github.com/hazelcast/hazelcast-go-client v1.1.2-0.20220124142245-1906eb58ac78
