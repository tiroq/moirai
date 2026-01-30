#!/usr/bin/env bash
set -euo pipefail

# golangci-lint v2 uses a v2 Go module path.
GOLANGCI_LINT_VERSION="v2.8.0"

if command -v golangci-lint >/dev/null 2>&1; then
  golangci-lint --version
  exit 0
fi

go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"

golangci-lint --version
