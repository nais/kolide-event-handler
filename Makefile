PROTOC = $(shell which protoc)
PROTOC_GEN_GO = $(shell which protoc-gen-go)

.PHONY: proto build test all

all: test build

install-protobuf-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

proto:
	$(PROTOC) --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=. --go-grpc_out=. pkg/pb/kolide-event-handler.proto

test:
	go test ./...

build:
	go build -o ./bin/kolide-event-handler ./cmd/kolide-event-handler/

fmt:
	go run mvdan.cc/gofumpt -w ./