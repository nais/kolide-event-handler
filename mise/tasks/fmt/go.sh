#!/usr/bin/env bash
#MISE description="Format go code"
set -euo pipefail

go tool mvdan.cc/gofumpt -w ./
