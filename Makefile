.PHONY: build test test-cover view-cover

GIT_COMMIT = $(shell git rev-parse HEAD 2> /dev/null || echo unknown)
CLC_VERSION ?= v0.0.0
LDFLAGS = "-s -w -X 'github.com/hazelcast/hazelcast-commandline-client/internal.GitCommit=$(GIT_COMMIT)' -X 'github.com/hazelcast/hazelcast-commandline-client/internal.Version=$(CLC_VERSION)' -X 'github.com/hazelcast/hazelcast-go-client/internal.ClientType=CLC' -X 'github.com/hazelcast/hazelcast-go-client/internal.ClientVersion=$(CLC_VERSION)'"
TEST_FLAGS ?= -v -count 1
COVERAGE_OUT = coverage.out
PACKAGES = $(shell go list ./... | grep -v internal/it | tr '\n' ',')
BINARY_NAME ?= clc
GOOS ?= linux
GOARCH ?= amd64
RELEASE_BASE ?= hazelcast-clc_$(CLC_VERSION)_$(GOOS)_$(GOARCH)
RELEASE_FILE ?= release.txt
TARGZ ?= true

build:
	CGO_ENABLED=0 go build -tags base,hazelcastinternal,hazelcastinternaltest -ldflags $(LDFLAGS)  -o build/$(BINARY_NAME) ./cmd/clc

test:
	go test -tags base,hazelcastinternal,hazelcastinternaltest $(TEST_FLAGS) ./...

test-cover:
	go test -tags base,hazelcastinternal,hazelcastinternaltest $(TEST_FLAGS) -coverprofile=coverage.out -coverpkg $(PACKAGES) -coverprofile=$(COVERAGE_OUT) ./...

view-cover:
	go tool cover -func $(COVERAGE_OUT) | grep total:
	go tool cover -html $(COVERAGE_OUT) -o coverage.html

release: build
	mkdir -p build/$(RELEASE_BASE)
	cp LICENSE build/$(RELEASE_BASE)/LICENSE.txt
	cp README.md build/$(RELEASE_BASE)/README.txt
	cp build/$(BINARY_NAME) build/$(RELEASE_BASE)
ifeq ($(TARGZ), false)
	cd build && zip -r $(RELEASE_BASE).zip $(RELEASE_BASE)
	echo $(RELEASE_BASE).zip >> build/$(RELEASE_FILE)
else
	tar cfz build/$(RELEASE_BASE).tar.gz -C build $(RELEASE_BASE)
	echo $(RELEASE_BASE).tar.gz >> build/$(RELEASE_FILE)
endif
