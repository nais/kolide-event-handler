#!/usr/bin/env bash
#MISE description="Run gosec"
set -euo pipefail

go tool github.com/securego/gosec/v2/cmd/gosec --exclude G102 --exclude-generated -terse ./...
