#!/usr/bin/env bash
#MISE description="Generate protobuf"
set -euo pipefail

protoc \
  --go-grpc_opt="paths=source_relative" \
  --go_opt="paths=source_relative" \
  --go_out="." \
  --go-grpc_out="." \
  pkg/pb/kolide-event-handler.proto