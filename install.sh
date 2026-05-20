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

build_config_tool() {
    hdr "config tool"
    if ! command -v go >/dev/null; then
        warn "go not found — skipping ai-rofi-config (TUI unavailable)"
        warn "install Go (https://go.dev/dl/) then re-run this installer to enable launch --config"
        return 0
    fi
    info "building ai-rofi-config (Go)"
    (
        cd "$SRC_DIR/config" || exit 1
        go build -o "$BIN_DIR/ai-rofi-config" . 2>&1 | sed 's/^/    /'
    ) || { err "build failed"; return 1; }
    chmod +x "$BIN_DIR/ai-rofi-config"
    ok "$BIN_DIR/ai-rofi-config"
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

write_config_template() {
    local cfg="$1" provider="$2" key_line="$3"
    {
        echo "LAUNCH_PROVIDER=$provider"
        echo
        echo "# === Anthropic — https://console.anthropic.com/settings/keys ==="
        [ "$provider" = anthropic ] && [ -n "$key_line" ] && echo "$key_line" || echo "# export ANTHROPIC_API_KEY=sk-ant-..."
        echo "LAUNCH_ANTHROPIC_FAST=claude-haiku-4-5-20251001"
        echo "LAUNCH_ANTHROPIC_DEEP=claude-sonnet-4-6"
        echo
        echo "# === OpenAI — https://platform.openai.com/api-keys ==="
        [ "$provider" = openai ] && [ -n "$key_line" ] && echo "$key_line" || echo "# export OPENAI_API_KEY=sk-..."
        echo "LAUNCH_OPENAI_FAST=gpt-4o-mini"
        echo "LAUNCH_OPENAI_DEEP=gpt-4o"
        echo
        echo "# === Groq — https://console.groq.com/keys ==="
        [ "$provider" = groq ] && [ -n "$key_line" ] && echo "$key_line" || echo "# export GROQ_API_KEY=gsk_..."
        echo "LAUNCH_GROQ_FAST=llama-3.1-8b-instant"
        echo "LAUNCH_GROQ_DEEP=llama-3.3-70b-versatile"
        echo
        echo "# === Ollama (local — no key needed) ==="
        echo "LAUNCH_OLLAMA_URL=http://localhost:11434"
        echo "LAUNCH_OLLAMA_FAST=llama3.2:3b"
        echo "LAUNCH_OLLAMA_DEEP=llama3.1:8b"
        echo
        echo "# === Ollama Cloud — https://ollama.com/settings/keys ==="
        [ "$provider" = ollama-cloud ] && [ -n "$key_line" ] && echo "$key_line" || echo "# export OLLAMA_API_KEY=..."
        echo "LAUNCH_OLLAMA_CLOUD_URL=https://ollama.com"
        echo "LAUNCH_OLLAMA_CLOUD_FAST=gpt-oss:20b"
        echo "LAUNCH_OLLAMA_CLOUD_DEEP=gpt-oss:120b"
    } > "$cfg"
    chmod 600 "$cfg"
}

env_key_for() {
    case "$1" in
        anthropic)    echo ANTHROPIC_API_KEY ;;
        openai)       echo OPENAI_API_KEY ;;
        groq)         echo GROQ_API_KEY ;;
        ollama-cloud) echo OLLAMA_API_KEY ;;
        *)            echo "" ;;
    esac
}

configure_key() {
    hdr "ai providers"
    local cfg="$CONF_DIR/config"

    if [ -f "$cfg" ] && grep -q '^LAUNCH_PROVIDER=' "$cfg"; then
        ok "existing config detected at $cfg"
        info "edit it manually to add provider keys or change models"
        return 0
    fi

    if [ -f "$cfg" ] && grep -q '^export ANTHROPIC_API_KEY=' "$cfg"; then
        local existing; existing=$(grep '^export ANTHROPIC_API_KEY=' "$cfg" | head -1)
        write_config_template "$cfg" anthropic "$existing"
        ok "migrated legacy config to multi-provider format"
        return 0
    fi

    info "supported providers:"
    printf "    1) anthropic     %sClaude — recommended%s\n" "$c_c" "$c_reset"
    printf "    2) openai        GPT family\n"
    printf "    3) groq          fastest, generous free tier\n"
    printf "    4) ollama        local, no key required\n"
    printf "    5) ollama-cloud  hosted Ollama\n"

    local provider="anthropic"
    if [ -t 0 ]; then
        printf "  default provider [1]: "
        local choice=""; read -r choice || true
        case "$choice" in
            2) provider=openai ;;
            3) provider=groq ;;
            4) provider=ollama ;;
            5) provider=ollama-cloud ;;
            *) provider=anthropic ;;
        esac
    fi
    ok "default: $provider"

    local key_line=""
    if [ "$provider" != ollama ]; then
        local var; var=$(env_key_for "$provider")
        if [ -n "${!var:-}" ]; then
            key_line=$(printf 'export %s=%q' "$var" "${!var}")
            ok "using $var from env"
        elif [ -t 0 ]; then
            info "paste your $var (input hidden, leave empty to skip):"
            local key=""; IFS= read -r -s key || true; echo
            if [ -n "$key" ]; then
                key_line=$(printf 'export %s=%q' "$var" "$key")
                ok "key captured"
            else
                warn "no key — edit $cfg later"
            fi
        else
            warn "non-interactive — edit $cfg to add $var"
        fi
    fi

    write_config_template "$cfg" "$provider" "$key_line"
    ok "wrote $cfg (mode 600)"
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
    if [ -x "$BIN_DIR/ai-rofi-config" ]; then
        info "configure TUI:     launch --config"
    fi
    info "edit YAML at:      $CONF_DIR/config.yaml"
    info "uninstall:         bash $SRC_DIR/uninstall.sh"
}

migrate_to_yaml() {
    [ -x "$BIN_DIR/ai-rofi-config" ] || return 0
    [ -f "$CONF_DIR/config.yaml" ] && return 0
    [ -f "$CONF_DIR/config" ] || return 0
    hdr "migrating shell config → YAML"
    if "$BIN_DIR/ai-rofi-config" migrate 2>/dev/null; then
        ok "config.yaml created"
    fi
}

main() {
    printf "\n%s%s%s\n" "$c_b" "ai-rofi-launcher · installer" "$c_reset"
    ensure_deps
    install_files
    build_config_tool
    ensure_path
    configure_key
    migrate_to_yaml
    suggest_hotkey
    post_check
}

main "$@"
