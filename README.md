# ai-rofi-launcher

A universal launcher for Linux that routes whatever you type to the right tool — Claude, calculator, web search, shell, app launcher — all behind one hotkey.

```
┌──────────────────────────────────────────────────┐
│  λ  ask anything…                                │
│    ai    deep   = calc   g/gh/yt/w/d   > shell   │
└──────────────────────────────────────────────────┘
```

## Features

- **Talkroom web UI by default** — your hotkey opens a polished local web app with multi-chat sidebar, typewriter rendering, voice input, text-to-speech, and a per-session provider picker.
- **5 providers** — Anthropic Claude, OpenAI GPT, Groq Llama, local Ollama, Ollama Cloud. Keys live in your YAML config; provider switch is a dropdown in the topbar.
- **Multi-turn context** — full conversation history sent on every turn for real continuity.
- **Local proxy** — your API keys never leave your machine; the Go binary serves the HTML and proxies requests to whichever provider you pick. CORS-free and key-safe.
- **Voice in, voice out** — Web Speech API for mic input, SpeechSynthesis for reading replies aloud (toggleable).
- **Rofi fallback** — `launch --rofi` keeps the minimal popup mode if you prefer keyboard-only.
- **Bubbletea TUI for settings** — `launch --config` to edit provider, models, and keys with arrow keys.

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

1. Press your hotkey (e.g. Super+A) → `launch`.
2. **Talkroom opens in your browser** at `http://127.0.0.1:8765`. The first invocation starts a local Go server; subsequent invocations reuse it (just open a new tab).
3. **Pick a provider** from the topbar dropdown — defaults to whatever you set in `launch --config`. Only providers with a key configured (plus local Ollama) appear.
4. **Type or speak.** Hit the mic to dictate (Web Speech API). Toggle "speak" to have replies read aloud. Tabbed sidebar for multiple conversations.
5. **Multi-turn context** — every send includes the full conversation history so far.

To go keyboard-only: `launch --rofi`. To switch the default permanently: `LAUNCH_MODE=rofi launch` or export that var in your shell rc.

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
