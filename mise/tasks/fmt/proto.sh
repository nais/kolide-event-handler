#!/usr/bin/env bash
#MISE description="Format proto"
set -euo pipefail

buf format -w pkg/pb/kolide-event-handler.proto
