#!/usr/bin/env bash
set -euo pipefail

fail() {
  echo "smoke: $*" >&2
  exit 1
}

root_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
bin_path="${MOIRAI_BIN:-$root_dir/bin/moirai}"

if [ ! -x "$bin_path" ]; then
  fail "moirai binary not found or not executable: $bin_path"
fi

tmp_dir=$(mktemp -d 2>/dev/null || mktemp -d -t moirai-smoke)
trap 'rm -rf "$tmp_dir"' EXIT

export XDG_CONFIG_HOME="$tmp_dir/xdg"
export HOME="$tmp_dir/home"
config_dir="$XDG_CONFIG_HOME/opencode"
mkdir -p "$config_dir" "$HOME"

alpha_path="$config_dir/oh-my-opencode.json.alpha"
beta_path="$config_dir/oh-my-opencode.json.beta"

alpha_content='{"agents":{"sisyphus":{"model":"gpt-4","extra_alpha":"keep"}},"unknown_root":"alpha","nested":{"keep":true}}'
beta_content='{"agents":{"sisyphus":{"model":"gpt-4","extra_beta":"keep"}},"unknown_root":"beta","nested":{"keep":true}}'

printf '%s' "$alpha_content" > "$alpha_path"
printf '%s' "$beta_content" > "$beta_path"

list_output=$("$bin_path" list)
printf '%s\n' "$list_output" | grep -qF "ConfigDir: $config_dir" || fail "list used unexpected config dir"
printf '%s\n' "$list_output" | grep -qF " - alpha" || fail "list missing alpha"
printf '%s\n' "$list_output" | grep -qF " - beta" || fail "list missing beta"

"$bin_path" apply alpha
active_path="$config_dir/oh-my-opencode.json"
if [ ! -L "$active_path" ]; then
  fail "active config is not a symlink after apply alpha"
fi
active_target=$(readlink "$active_path")
if [ "$active_target" != "oh-my-opencode.json.alpha" ]; then
  fail "unexpected symlink target after apply alpha: $active_target"
fi

"$bin_path" apply beta
if [ ! -L "$active_path" ]; then
  fail "active config is not a symlink after apply beta"
fi
active_target=$(readlink "$active_path")
if [ "$active_target" != "oh-my-opencode.json.beta" ]; then
  fail "unexpected symlink target after apply beta: $active_target"
fi
if [ ! -f "$alpha_path" ]; then
  fail "alpha profile missing after apply beta"
fi
if [ ! -f "$beta_path" ]; then
  fail "beta profile missing after apply beta"
fi

backup_output=$("$bin_path" backup beta)
backup_path=$(printf '%s\n' "$backup_output" | awk -F'Backup: ' '/^Backup: / {print $2}')
if [ -z "$backup_path" ]; then
  fail "backup path not reported"
fi
if [ ! -f "$backup_path" ]; then
  fail "backup file missing: $backup_path"
fi
backup_base=$(basename "$backup_path")
case "$backup_base" in
  oh-my-opencode.json.beta.bak.*) ;;
  *) fail "backup name unexpected: $backup_base" ;;
esac

set +e
"$bin_path" diff beta --against last-backup >/dev/null
diff_status=$?
set -e
if [ "$diff_status" -ne 0 ] && [ "$diff_status" -ne 2 ]; then
  fail "diff returned unexpected status: $diff_status"
fi

beta_modified='{"agents":{"sisyphus":{"model":"gpt-4","extra_beta":"changed"}},"unknown_root":"beta-changed","nested":{"keep":false}}'
printf '%s' "$beta_modified" > "$beta_path"

restore_output=$("$bin_path" restore beta --from "$backup_path")
pre_backup_path=$(printf '%s\n' "$restore_output" | awk -F'PreBackup: ' '/^PreBackup: / {print $2}')
if [ -z "$pre_backup_path" ]; then
  fail "pre-restore backup path not reported"
fi
if [ ! -f "$pre_backup_path" ]; then
  fail "pre-restore backup missing: $pre_backup_path"
fi
pre_backup_base=$(basename "$pre_backup_path")
case "$pre_backup_base" in
  oh-my-opencode.json.beta.bak.*) ;;
  *) fail "pre-restore backup name unexpected: $pre_backup_base" ;;
esac

restored_content=$(cat "$beta_path")
if [ "$restored_content" != "$beta_content" ]; then
  fail "restored beta profile content mismatch"
fi
printf '%s\n' "$restored_content" | grep -qF '"unknown_root":"beta"' || fail "unknown root field missing"
printf '%s\n' "$restored_content" | grep -qF '"extra_beta":"keep"' || fail "unknown agent field missing"

set +e
"$bin_path" doctor beta >/dev/null
doctor_status=$?
set -e
if [ "$doctor_status" -ne 0 ] && [ "$doctor_status" -ne 2 ]; then
  fail "doctor returned unexpected status: $doctor_status"
fi
