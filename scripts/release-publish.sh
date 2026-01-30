#!/usr/bin/env bash
set -euo pipefail

die() {
  echo "error: $*" >&2
  exit 1
}

usage() {
  cat >&2 <<'EOF'
Usage:
  scripts/release-publish.sh vX.Y.Z

Notes:
  - Refuses to run if the tag does not exist locally.
  - Publishes a prepared tag by pushing it to origin.
EOF
  exit 2
}

require_git() {
  command -v git >/dev/null 2>&1 || die "git is required"
}

validate_version() {
  local v="$1"
  [[ "$v" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "invalid version format: $v"
}

tag_exists() {
  local v="$1"
  git rev-parse -q --verify "refs/tags/$v" >/dev/null 2>&1
}

main() {
  require_git

  if [ $# -ne 1 ]; then
    usage
  fi

  local version="$1"
  validate_version "$version"

  if ! tag_exists "$version"; then
    die "tag does not exist locally: $version"
  fi

  git push origin "$version"

  echo "Release $version published. GitHub Actions will build assets."
}

main "$@"
