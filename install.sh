#!/usr/bin/env bash
set -euo pipefail

APP_NAME="ai-rofi-launcher"
BIN_DIR="${HOME}/.local/bin"
CONF_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/$APP_NAME"
SRC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

c_reset=$'\033[0m'; c_b=$'\033[1m'; c_g=$'\033[32m'; c_y=$'\033[33m'; c_r=$'\033[31m'; c_c=$'\033[36m'
ok()   { printf "  %s✓%s %s\n" "$c_g" "$c_reset" "$1"; }
info() { printf "  %s→%s %s\n" "$c_c" "$c_reset" "$1"; }
warn() { printf "  %s!%s %s\n" "$c_y" "$c_reset" "$1"; }
err()  { printf "  %sx%s %s\n" "$c_r" "$c_reset" "$1" >&2; }
hdr()  { printf "\n%s%s%s\n" "$c_b" "$1" "$c_reset"; }

need_cmds=(rofi qalc xdotool jq curl)
opt_cmds=(notify-send xclip wl-copy xdg-open rofimoji)

detect_pm() {
    if   command -v pacman  >/dev/null; then echo "pacman"
    elif command -v apt-get >/dev/null; then echo "apt"
    elif command -v dnf     >/dev/null; then echo "dnf"
    elif command -v zypper  >/dev/null; then echo "zypper"
    elif command -v brew    >/dev/null; then echo "brew"
    else echo "unknown"; fi
}

pkg_for() {
    local pm="$1" cmd="$2"
    case "$pm:$cmd" in
        pacman:rofi)        echo "rofi" ;;
        pacman:qalc)        echo "libqalculate" ;;
        pacman:xdotool)     echo "xdotool" ;;
        pacman:jq)          echo "jq" ;;
        pacman:curl)        echo "curl" ;;
        pacman:notify-send) echo "libnotify" ;;
        pacman:xclip)       echo "xclip" ;;
        pacman:wl-copy)     echo "wl-clipboard" ;;
        pacman:xdg-open)    echo "xdg-utils" ;;
        pacman:rofimoji)    echo "rofimoji" ;;
        apt:rofi)           echo "rofi" ;;
        apt:qalc)           echo "qalc" ;;
        apt:xdotool)        echo "xdotool" ;;
        apt:jq)             echo "jq" ;;
        apt:curl)           echo "curl" ;;
        apt:notify-send)    echo "libnotify-bin" ;;
        apt:xclip)          echo "xclip" ;;
        apt:wl-copy)        echo "wl-clipboard" ;;
        apt:xdg-open)       echo "xdg-utils" ;;
        apt:rofimoji)       echo "rofimoji" ;;
        dnf:rofi)           echo "rofi" ;;
        dnf:qalc)           echo "libqalculate" ;;
        dnf:xdotool)        echo "xdotool" ;;
        dnf:jq)             echo "jq" ;;
        dnf:curl)           echo "curl" ;;
        dnf:notify-send)    echo "libnotify" ;;
        dnf:xclip)          echo "xclip" ;;
        dnf:wl-copy)        echo "wl-clipboard" ;;
        dnf:xdg-open)       echo "xdg-utils" ;;
        dnf:rofimoji)       echo "rofimoji" ;;
        *) echo "$cmd" ;;
    esac
}

install_pkgs() {
    local pm="$1"; shift
    [ $# -eq 0 ] && return 0
    case "$pm" in
        pacman) sudo pacman -S --needed --noconfirm "$@" ;;
        apt)    sudo apt-get update && sudo apt-get install -y "$@" ;;
        dnf)    sudo dnf install -y "$@" ;;
        zypper) sudo zypper install -y "$@" ;;
        brew)   brew install "$@" ;;
        *)      err "no known package manager — install manually: $*" ; return 1 ;;
    esac
}

ensure_deps() {
    hdr "checking dependencies"
    local pm; pm=$(detect_pm)
    info "package manager: $pm"
    local missing=()
    for c in "${need_cmds[@]}"; do
        if command -v "$c" >/dev/null; then ok "$c"; else warn "$c missing"; missing+=("$(pkg_for "$pm" "$c")"); fi
    done
    for c in "${opt_cmds[@]}"; do
        command -v "$c" >/dev/null && ok "$c (optional)" || warn "$c missing (optional)"
    done
    if [ ${#missing[@]} -gt 0 ]; then
        info "installing: ${missing[*]}"
        install_pkgs "$pm" "${missing[@]}" || { err "dependency install failed"; exit 1; }
    fi
}

install_files() {
    hdr "installing files"
    mkdir -p "$BIN_DIR" "$CONF_DIR/themes"
    install -m 755 "$SRC_DIR/bin/launch" "$BIN_DIR/launch"
    ok "$BIN_DIR/launch"
    install -m 644 "$SRC_DIR/themes/prompt.rasi" "$CONF_DIR/themes/prompt.rasi"
    install -m 644 "$SRC_DIR/themes/view.rasi"   "$CONF_DIR/themes/view.rasi"
    ok "$CONF_DIR/themes/"
}

ensure_path() {
    case ":$PATH:" in
        *":$BIN_DIR:"*) ok "PATH includes $BIN_DIR" ;;
        *)
            warn "$BIN_DIR not in PATH"
            local rc="$HOME/.bashrc"
            [ -n "${ZSH_VERSION:-}" ] || [[ "${SHELL:-}" == */zsh ]] && rc="$HOME/.zshrc"
            if ! grep -qs '\.local/bin' "$rc" 2>/dev/null; then
                printf '\nexport PATH="$HOME/.local/bin:$PATH"\n' >> "$rc"
                ok "added to $rc (open a new shell)"
            fi
            ;;
    esac
}

configure_key() {
    hdr "api key"
    local cfg="$CONF_DIR/config"
    if [ -f "$cfg" ] && grep -q '^export ANTHROPIC_API_KEY=' "$cfg"; then
        ok "key already configured in $cfg"
        return 0
    fi
    if [ -n "${ANTHROPIC_API_KEY:-}" ]; then
        printf 'export ANTHROPIC_API_KEY=%q\n' "$ANTHROPIC_API_KEY" > "$cfg"
        chmod 600 "$cfg"
        ok "saved key from env to $cfg"
        return 0
    fi
    if [ ! -t 0 ]; then
        warn "no key found and not running interactively"
        warn "edit $cfg later and add: export ANTHROPIC_API_KEY=sk-ant-..."
        : > "$cfg"; chmod 600 "$cfg"
        return 0
    fi
    info "paste your Anthropic API key (sk-ant-…), or leave empty to skip:"
    local key=""
    IFS= read -r -s key || true
    echo
    if [ -z "$key" ]; then
        warn "no key entered — edit $cfg later"
        : > "$cfg"; chmod 600 "$cfg"
    else
        printf 'export ANTHROPIC_API_KEY=%q\n' "$key" > "$cfg"
        chmod 600 "$cfg"
        ok "saved to $cfg (mode 600)"
    fi
}

detect_wm() {
    local wm="${XDG_CURRENT_DESKTOP:-}"
    if [ -z "$wm" ]; then wm="${DESKTOP_SESSION:-}"; fi
    if pgrep -x hyprland >/dev/null 2>&1; then echo "hyprland"; return; fi
    if pgrep -x sway     >/dev/null 2>&1; then echo "sway";     return; fi
    if pgrep -x i3       >/dev/null 2>&1; then echo "i3";       return; fi
    if pgrep -x openbox  >/dev/null 2>&1; then echo "openbox";  return; fi
    if pgrep -x bspwm    >/dev/null 2>&1; then echo "bspwm";    return; fi
    if pgrep -x awesome  >/dev/null 2>&1; then echo "awesome";  return; fi
    if pgrep -x kwin_x11 >/dev/null 2>&1 || pgrep -x kwin_wayland >/dev/null 2>&1; then echo "kde"; return; fi
    if pgrep -x gnome-shell >/dev/null 2>&1; then echo "gnome"; return; fi
    echo "${wm,,}"
}

suggest_hotkey() {
    hdr "hotkey binding"
    local wm; wm=$(detect_wm)
    info "detected: ${wm:-unknown}"
    local cmd="$BIN_DIR/launch"
    case "$wm" in
        openbox)
            cat <<EOF

  add this inside <keyboard> in ~/.config/openbox/rc.xml, then run:
      openbox --reconfigure

  <keybind key="W-space">
    <action name="Execute"><command>$cmd</command></action>
  </keybind>
EOF
            ;;
        i3|sway)
            cat <<EOF

  add to ~/.config/$wm/config and reload ($wm reload):

      bindsym \$mod+space exec $cmd
EOF
            ;;
        hyprland)
            cat <<EOF

  add to ~/.config/hypr/hyprland.conf and reload:

      bind = SUPER, SPACE, exec, $cmd
EOF
            ;;
        bspwm)
            cat <<EOF

  add to ~/.config/sxhkd/sxhkdrc and run: pkill -USR1 -x sxhkd

      super + space
          $cmd
EOF
            ;;
        kde|*kde*|plasma|*plasma*)
            cat <<EOF

  System Settings → Shortcuts → Custom Shortcuts → New → Global Shortcut → Command:
      $cmd
  Bind it to Meta+Space.
EOF
            ;;
        gnome|*gnome*)
            cat <<EOF

  Settings → Keyboard → Keyboard Shortcuts → Custom → +
      Name:    ai-rofi-launcher
      Command: $cmd
      Shortcut: Super+Space
EOF
            ;;
        *)
            cat <<EOF

  bind your WM's keybinding to:
      $cmd
EOF
            ;;
    esac
    echo
}

post_check() {
    hdr "all set"
    info "run it now:        launch"
    info "edit your key at:  $CONF_DIR/config"
    info "uninstall:         bash $SRC_DIR/uninstall.sh"
}

main() {
    printf "\n%s%s%s\n" "$c_b" "ai-rofi-launcher · installer" "$c_reset"
    ensure_deps
    install_files
    ensure_path
    configure_key
    suggest_hotkey
    post_check
}

main "$@"
