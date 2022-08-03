.PHONY: build generate-completion test test-cover view-cover

GIT_COMMIT=$(shell git rev-parse HEAD 2> /dev/null || echo unknown)
CLC_VERSION=$(shell git describe --tags `git rev-list --tags --max-count=1` || echo UNKNOWN)
LDFLAGS="-X 'github.com/hazelcast/hazelcast-go-client/internal.ClientType=CLC' -X 'github.com/hazelcast/hazelcast-commandline-client/internal.GitCommit=$(GIT_COMMIT)' -X 'github.com/hazelcast/hazelcast-commandline-client/internal.ClientVersion=$(CLC_VERSION)' -X 'github.com/hazelcast/hazelcast-go-client/internal.ClientVersion=$(CLC_VERSION)'"
TEST_FLAGS ?= -v -count 1
COVERAGE_OUT = coverage.out

build:
	go build -ldflags $(LDFLAGS) -o hzc github.com/hazelcast/hazelcast-commandline-client

generate-completion: build
	mkdir -p extras
	MODE="dev" ./hzc completion bash --no-descriptions > extras/bash_completion.sh
	MODE="dev" ./hzc completion zsh --no-descriptions > extras/zsh_completion.zsh

test:
	go test $(TESTFLAGS) ./...

test-cover:
	go test $(TESTFLAGS) -coverprofile=$(COVERAGE_OUT) ./...

view-cover:
	go tool cover -func $(COVERAGE_OUT) | grep total:
	go tool cover -html $(COVERAGE_OUT) -o coverage.html
