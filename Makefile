.PHONY: build

build:
	go build -o hzc github.com/hazelcast/hazelcast-commandline-client
generate-completion:
	mkdir -p extras
	MODE="dev" go run github.com/hazelcast/hazelcast-commandline-client completion bash --no-descriptions > extras/bash_completion.sh
	MODE="dev" go run github.com/hazelcast/hazelcast-commandline-client completion zsh --no-descriptions > extras/zsh_completion.zsh