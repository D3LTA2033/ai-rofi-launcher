# ai-rofi-launcher

A universal launcher for Linux that routes whatever you type to the right tool — Claude, calculator, web search, shell, app launcher — all behind one hotkey.

```
┌──────────────────────────────────────────────────┐
│  λ  ask anything…                                │
│    ai    deep   = calc   g/gh/yt/w/d   > shell   │
└──────────────────────────────────────────────────┘
```

## Features

- **Pure AI chat from a rofi popup.** Hit your hotkey, pick which AI you want for this session (your configured default is pre-highlighted — Enter accepts), then talk.
- **Real multi-turn context** — every reply is computed with the full prior conversation, not as isolated questions.
- **5 providers supported** — Anthropic Claude, OpenAI GPT, Groq Llama, local Ollama, Ollama Cloud. Per-session choice, no per-message prefixes to remember.
- **Latest reply rendered in-line** above the input on the next turn, so you read the answer the same place you type the follow-up.
- **Answers auto-copied to clipboard.**
- **Bubbletea TUI for settings** — `launch --config` to edit provider, models, and keys with arrow keys.
- **Catppuccin-styled rofi themes** — separate prompt + chat-view themes.

## Install

```bash
git clone https://github.com/D3LTA2033/ai-rofi-launcher.git
cd ai-rofi-launcher
bash install.sh
```

The installer:
1. Detects your distro (pacman / apt / dnf / zypper / brew) and installs missing deps
2. Copies `launch` to `~/.local/bin/` and themes to `~/.config/ai-rofi-launcher/themes/`
3. Builds `ai-rofi-config` (Go binary) for the interactive config TUI — if Go is installed
4. Adds `~/.local/bin` to your `PATH` if missing
5. Prompts you to pick a default provider and enter its API key
6. Migrates any legacy shell config to `config.yaml`
7. Detects your window manager and prints the exact hotkey snippet to paste

## Configure

```bash
launch --config        # opens the TUI
```

Arrow-key navigation, `Enter` to edit a field, `c` to clear a key, `s` to save, `q` to quit. API keys are entered with hidden input and displayed as `●●●●●●●● set` once configured.

Or edit `~/.config/ai-rofi-launcher/config.yaml` directly:

```yaml
default_provider: anthropic

providers:
  anthropic:
    api_key: sk-ant-...
    fast: claude-haiku-4-5-20251001
    deep: claude-sonnet-4-6
  openai:
    api_key: ""
    fast: gpt-4o-mini
    deep: gpt-4o
  groq:
    api_key: gsk_...
    fast: llama-3.1-8b-instant
    deep: llama-3.3-70b-versatile
  ollama:
    url: http://localhost:11434
    fast: llama3.2:3b
    deep: llama3.1:8b
  ollama_cloud:
    api_key: ""
    url: https://ollama.com
    fast: gpt-oss:20b
    deep: gpt-oss:120b
```

## AI providers

| Code | Provider     | Where to get a key                          |
|------|--------------|---------------------------------------------|
| `?a` | Anthropic    | https://console.anthropic.com/settings/keys |
| `?o` | OpenAI       | https://platform.openai.com/api-keys        |
| `?g` | Groq         | https://console.groq.com/keys  *(free tier)*|
| `?l` | Ollama local | `ollama serve` — no key                     |
| `?c` | Ollama Cloud | https://ollama.com/settings/keys            |

Bare `?` uses whatever you set as `LAUNCH_PROVIDER` in the config — pick one at install time, switch any time by editing `~/.config/ai-rofi-launcher/config`.

You can mix and match: type `?g` for an answer in 300 ms via Groq, `??a` when you want Claude Sonnet to think harder, `?l` for fully offline via local Ollama.

## Hotkey suggestions

The installer prints the right snippet for your WM. Common ones:

| WM        | Snippet                                                         |
|-----------|-----------------------------------------------------------------|
| Hyprland  | `bind = SUPER, SPACE, exec, ~/.local/bin/launch`                |
| i3 / Sway | `bindsym $mod+space exec ~/.local/bin/launch`                   |
| Openbox   | `<keybind key="W-space">…</keybind>` in `rc.xml`                |
| bspwm     | `super + space` → `~/.local/bin/launch` in `sxhkdrc`            |
| KDE       | Settings → Shortcuts → Custom → Global → `~/.local/bin/launch`  |
| GNOME     | Settings → Keyboard → Custom Shortcut → `~/.local/bin/launch`   |

## Themes

`~/.config/ai-rofi-launcher/themes/` — edit `prompt.rasi` (the input bar) and `view.rasi` (the result viewer) to taste. Default palette is Catppuccin Mocha.

## Advanced — `ai-rofi-config` subcommands

```bash
ai-rofi-config              # opens the TUI
ai-rofi-config export       # prints shell-evaluable env (used internally by launch)
ai-rofi-config migrate      # converts legacy shell config to config.yaml
ai-rofi-config path         # prints the config file path
ai-rofi-config show         # prints current config as YAML
```

The bash launcher invokes `ai-rofi-config export` on every run if `config.yaml` exists, falling back to the legacy shell config otherwise — so this remains fully backward compatible.

## Flow

1. Press your hotkey (e.g. Super+A).
2. **Picker** appears with 5 providers; your configured default has a ★ next to it and the cursor is on it. Providers with no API key configured show ⚠. `↵` to start.
3. **Chat** opens — a rofi popup with an input field and the latest AI reply rendered above it. Type, `↵` to send. Reply appears above the input on the next turn (and is auto-copied to clipboard). Conversation continues until you hit `Esc` or submit an empty message.

The conversation is kept entirely in memory for the session; nothing is saved to disk between sessions.

## Uninstall

```bash
bash uninstall.sh
```

Removes the script, themes, config, and history. WM hotkey bindings you set yourself need to be removed manually.

## Dependencies

Required: `rofi`, `libqalculate` (qalc), `xdotool`, `jq`, `curl`
Optional: `libnotify` (notify-send), `xclip` or `wl-clipboard`, `xdg-utils`, `rofimoji`
For the config TUI: `go` (used only at install time to build a single static binary)

## License

MIT — see [LICENSE](LICENSE).
