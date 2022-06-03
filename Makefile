.PHONY: build
TAG=$(shell git describe --tags 2> /dev/null || echo unknown)
CLIENT_TYPE="CLC"
LDFLAGS="-X 'github.com/hazelcast/hazelcast-go-client/internal.ClientType=$(CLIENT_TYPE)' -X 'github.com/hazelcast/hazelcast-go-client/internal.ClientVersion=$(TAG)'"
build:
	go build -ldflags $(LDFLAGS) -o hzc github.com/hazelcast/hazelcast-commandline-client
generate-completion: build
	mkdir -p extras
	MODE="dev" ./hzc completion bash --no-descriptions > extras/bash_completion.sh
	MODE="dev" ./hzc completion zsh --no-descriptions > extras/zsh_completion.zsh
test:
	go test -v ./...