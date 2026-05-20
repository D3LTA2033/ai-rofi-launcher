#!/usr/bin/env bash
set -euo pipefail

APP_NAME="ai-rofi-launcher"
BIN="${HOME}/.local/bin/launch"
CONF_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/$APP_NAME"
CACHE_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/$APP_NAME"

c_reset=$'\033[0m'; c_g=$'\033[32m'; c_y=$'\033[33m'
ok()   { printf "  %s✓%s %s\n" "$c_g" "$c_reset" "$1"; }
warn() { printf "  %s!%s %s\n" "$c_y" "$c_reset" "$1"; }

[ -f "$BIN" ]       && rm -f "$BIN"       && ok "removed $BIN"       || warn "$BIN not found"
[ -d "$CONF_DIR" ]  && rm -rf "$CONF_DIR" && ok "removed $CONF_DIR"  || warn "$CONF_DIR not found"
[ -d "$CACHE_DIR" ] && rm -rf "$CACHE_DIR"&& ok "removed $CACHE_DIR" || warn "$CACHE_DIR not found"

echo
echo "remove your WM hotkey binding manually if you set one."
