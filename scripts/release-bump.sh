#!/usr/bin/env bash
set -euo pipefail

die() {
  echo "error: $*" >&2
  exit 1
}

usage() {
  cat >&2 <<'EOF'
Usage:
  scripts/release-bump.sh patch|minor|major
  scripts/release-bump.sh vX.Y.Z

Notes:
  - Refuses to run if the working tree is dirty.
  - Creates an annotated tag: git tag -a <version> -m "release <version>"
  - Does not push anything.
EOF
  exit 2
}

require_git() {
  command -v git >/dev/null 2>&1 || die "git is required"
}

is_clean_worktree() {
  git diff --quiet || return 1
  git diff --cached --quiet || return 1
}

latest_version_tag() {
  local t
  t=$(
    git tag --list --sort=-v:refname |
      grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' |
      head -n 1 || true
  )
  if [ -z "$t" ]; then
    echo "v0.0.0"
    return 0
  fi
  echo "$t"
}

validate_version() {
  local v="$1"
  [[ "$v" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "invalid version format: $v"
}

tag_exists() {
  local v="$1"
  git rev-parse -q --verify "refs/tags/$v" >/dev/null 2>&1
}

bump_from() {
  local base="$1"
  local kind="$2"
  local major minor patch

  base="${base#v}"
  IFS='.' read -r major minor patch <<<"$base"

  case "$kind" in
    patch)
      patch=$((patch + 1))
      ;;
    minor)
      minor=$((minor + 1))
      patch=0
      ;;
    major)
      major=$((major + 1))
      minor=0
      patch=0
      ;;
    *)
      die "unknown bump kind: $kind"
      ;;
  esac

  echo "v${major}.${minor}.${patch}"
}

main() {
  require_git

  if [ $# -ne 1 ]; then
    usage
  fi

  if ! is_clean_worktree; then
    die "working tree is dirty"
  fi

  local arg="$1"
  local next

  case "$arg" in
    patch|minor|major)
      next="$(bump_from "$(latest_version_tag)" "$arg")"
      ;;
    *)
      next="$arg"
      ;;
  esac

  validate_version "$next"

  if tag_exists "$next"; then
    die "tag already exists: $next"
  fi

  git tag -a "$next" -m "release $next"

  echo "created tag $next"
  echo "next: git push origin $next"
}

main "$@"
