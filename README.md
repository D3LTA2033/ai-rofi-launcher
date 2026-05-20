# ai-rofi-launcher

A universal launcher for Linux that routes whatever you type to the right tool — Claude, calculator, web search, shell, app launcher — all behind one hotkey.

```
┌──────────────────────────────────────────────────┐
│  λ  ask anything…                                │
│    ai    deep   = calc   g/gh/yt/w/d   > shell   │
└──────────────────────────────────────────────────┘
```

## Features

- **`?` ask AI** — uses your default provider, answer copied to clipboard
- **`??` deep model** — same provider, stronger model
- **`?a` / `?o` / `?g` / `?l` / `?c`** — force a specific provider for one query
  (Anthropic / OpenAI / Groq / Ollama local / Ollama Cloud) · prepend `??` for the deep variant
- **5 AI providers supported** — Anthropic Claude, OpenAI, Groq, local Ollama, Ollama Cloud
- **`=` calculator** — qalc with units, currency, dates
- **`g` / `gh` / `yt` / `w` / `d`** — Google / GitHub / YouTube / Wikipedia / DuckDuckGo
- **`>` shell** — run any command, see the output in a popup
- **`:emoji`** — emoji picker (via rofimoji)
- **URL** — opens in your browser
- **anything else** — falls through to the rofi app launcher
- **history** — past queries appear as fuzzy-searchable suggestions
- **catppuccin-styled rofi themes** — separate prompt + result-view themes

## Install

```bash
git clone https://github.com/D3LTA2033/ai-rofi-launcher.git
cd ai-rofi-launcher
bash install.sh
```

The installer:
1. Detects your distro (pacman / apt / dnf / zypper / brew) and installs missing deps
2. Copies `launch` to `~/.local/bin/` and themes to `~/.config/ai-rofi-launcher/themes/`
3. Adds `~/.local/bin` to your `PATH` if missing
4. Prompts for your Anthropic API key (saved to `~/.config/ai-rofi-launcher/config`, mode 600)
5. Detects your window manager and prints the exact hotkey snippet to paste

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

## Configuration

`~/.config/ai-rofi-launcher/config`:

```bash
LAUNCH_PROVIDER=anthropic       # anthropic | openai | groq | ollama | ollama-cloud

export ANTHROPIC_API_KEY=sk-ant-...
LAUNCH_ANTHROPIC_FAST=claude-haiku-4-5-20251001
LAUNCH_ANTHROPIC_DEEP=claude-sonnet-4-6

# export OPENAI_API_KEY=sk-...
LAUNCH_OPENAI_FAST=gpt-4o-mini
LAUNCH_OPENAI_DEEP=gpt-4o

# export GROQ_API_KEY=gsk_...
LAUNCH_GROQ_FAST=llama-3.1-8b-instant
LAUNCH_GROQ_DEEP=llama-3.3-70b-versatile

LAUNCH_OLLAMA_URL=http://localhost:11434
LAUNCH_OLLAMA_FAST=llama3.2:3b
LAUNCH_OLLAMA_DEEP=llama3.1:8b

# export OLLAMA_API_KEY=...
LAUNCH_OLLAMA_CLOUD_URL=https://ollama.com
LAUNCH_OLLAMA_CLOUD_FAST=gpt-oss:20b
LAUNCH_OLLAMA_CLOUD_DEEP=gpt-oss:120b
```

The installer writes this template for you; you only fill in the keys you actually want to use. Themes live in `~/.config/ai-rofi-launcher/themes/` — edit `prompt.rasi` and `view.rasi` to taste.

## Examples

| Type                                  | What happens                                       |
|---------------------------------------|----------------------------------------------------|
| `? best way to debug a memleak`       | default provider, fast model                       |
| `?? design a rate limiter at 10k rps` | default provider, deep model                       |
| `?g what's the capital of mongolia`   | Groq (sub-second answer)                           |
| `??a write a haiku about ohms law`    | Anthropic Sonnet                                   |
| `?o explain gradient descent`         | OpenAI GPT-4o-mini                                 |
| `?l summarize this for me…`           | Local Ollama — no API call leaves your machine     |
| `?c help me refactor this`            | Ollama Cloud                                       |
| `= 1.5 BTC to USD`                    | qalc with currency                                 |
| `42 * sqrt(7)`                        | auto-detected math                                 |
| `g rust async traits`                 | Google in your browser                             |
| `gh tokio runtime`                    | GitHub code search                                 |
| `> df -h /`                           | runs it, popup with output                         |
| `https://news.ycombinator.com`        | opens URL                                          |
| `firefox`                             | drun fallback launches the app                     |

## Uninstall

```bash
bash uninstall.sh
```

Removes the script, themes, config, and history. WM hotkey bindings you set yourself need to be removed manually.

## Dependencies

Required: `rofi`, `libqalculate` (qalc), `xdotool`, `jq`, `curl`
Optional: `libnotify` (notify-send), `xclip` or `wl-clipboard`, `xdg-utils`, `rofimoji`

## License

MIT — see [LICENSE](LICENSE).
