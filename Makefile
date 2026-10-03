BIN := "./bin/sys-mon-deamon"

CLIENT_BIN := "./bin/sys-mon-deamon-clt"

DOCKER_IMG="go-sys-monitor-deamon:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)



#Protoc
PROTOC_VERSION=36.1

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/sys-mon-deamon

build_client:
	go build -v -o $(CLIENT_BIN) -ldflags "$(LDFLAGS)" ./cmd/client

install-fmt-deps:
	(which gofumpt > /dev/null) || go install mvdan.cc/gofumpt@latest

fmt: install-fmt-deps
	gofmt -w ./
	gofumpt -l -w ./

run: run_cmd

run_cmd: build
	$(BIN) --config ./configs/go-sys-monitor-daemon.yaml

run_client: build_client
	$(CLIENT_BIN)  -host=localhost -port=8001 -log_level=DEBUG  -interval=5 -averaging-period=10 \
	-load_average=true  -cpu_stats=true -disks_load=true -disks_starts=true -network_top_talkers=true -network_conn_stats=true

run_client2: build_client
	$(CLIENT_BIN)  -host=localhost -port=8001 -log_level=DEBUG  -interval=10 -averaging-period=100 \
	-load_average=true  -cpu_stats=true -disks_load=true -disks_starts=true -network_top_talkers=true -network_conn_stats=true

run_client_error: build_client
	$(CLIENT_BIN)  -host=localhost -port=8001 -log_level=DEBUG  -interval=10 -averaging-period=101 \
	-load_average=true  -cpu_stats=true -disks_load=true -disks_starts=true -network_top_talkers=true -network_conn_stats=true



build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

unit-tests:
	go test -race ./internal/... ./cmd/... -v

integration-tests:
	go test -race ./integration-test -v

test: unit-tests integration-tests
	

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.11.4

lint: install-lint-deps
	golangci-lint run ./...


protoc-deps:
	(which protoc > /dev/null) || (curl -fsSL "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip" -o /tmp/protoc.zip && unzip -oq -d $(shell go env GOPATH) /tmp/protoc.zip)
	(which protoc-gen-go > /dev/null) || (go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12)
	(which protoc-gen-go-grpc > /dev/null) || (go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2)

protoc: protoc-deps
	go generate ./...

generate: protoc

.PHONY: build run build-img run-img version test lint
