#!/usr/bin/env bash
#MISE description="Build the project"
set -euo pipefail

go build -o ./bin/kolide-event-handler .