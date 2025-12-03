#!/usr/bin/env bash
#MISE description="Verify all github actions are pinned"
set -euo pipefail

go tool github.com/sethvargo/ratchet lint .github/workflows/*.yaml
