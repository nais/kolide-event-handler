PROTOC = $(shell which protoc)

.PHONY: proto build test all

all: test check build

install-protobuf-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

proto: install-protobuf-go
	PATH="${PATH}:$(shell go env GOPATH)/bin" $(PROTOC) --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative --go_out=. --go-grpc_out=. pkg/pb/kolide-event-handler.proto

test:
	go test ./...

build:
	go build -o ./bin/kolide-event-handler ./cmd/kolide-event-handler/

fmt:
	go tool mvdan.cc/gofumpt -w ./

check:
	go tool honnef.co/go/tools/cmd/staticcheck ./...
	go tool golang.org/x/vuln/cmd/govulncheck ./...
	go tool golang.org/x/tools/cmd/deadcode -test ./...