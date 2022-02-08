.PHONY: build

build:
	go build -o hzc github.com/hazelcast/hazelcast-commandline-client
generate-completion: build
	mkdir -p extras
	MODE="dev" ./hzc completion bash --no-descriptions > extras/bash_completion.sh
	MODE="dev" ./hzc completion zsh --no-descriptions > extras/zsh_completion.zsh
test:
	go test -v ./...