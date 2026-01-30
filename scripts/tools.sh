#!/usr/bin/env bash
set -euo pipefail

GOLANGCI_LINT_VERSION="v2.4.0"

if command -v golangci-lint >/dev/null 2>&1; then
  golangci-lint --version
  exit 0
fi

go install "github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"

golangci-lint --version
