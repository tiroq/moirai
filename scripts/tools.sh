#!/usr/bin/env bash
set -euo pipefail

GOLANGCI_LINT_VERSION="v2.8.0"

want_version="$GOLANGCI_LINT_VERSION"

if command -v golangci-lint >/dev/null 2>&1; then
  have_version="$(golangci-lint --version | grep -Eo 'v[0-9]+\\.[0-9]+\\.[0-9]+' | head -n 1 || true)"
  if [ "$have_version" = "$want_version" ]; then
    golangci-lint --version
    exit 0
  fi
fi

go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}"

golangci-lint --version
