.PHONY: all
all: test check build fmt

.PHONY: proto
proto:
	protoc pkg/pb/kolide-event-handler.proto \
		--go-grpc_opt="paths=source_relative" \
		--go_opt="paths=source_relative" \
		--go_out="." \
		--go-grpc_out="."

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -o ./bin/kolide-event-handler ./cmd/kolide-event-handler/

.PHONY: fmt
fmt:
	go tool mvdan.cc/gofumpt -w ./

.PHONY: check
check:
	go tool honnef.co/go/tools/cmd/staticcheck ./...
	go tool golang.org/x/vuln/cmd/govulncheck ./...
	go tool golang.org/x/tools/cmd/deadcode -test ./...
	go tool github.com/securego/gosec/v2/cmd/gosec --exclude-generated -terse ./...