# ai-rofi-launcher

A universal launcher for Linux that routes whatever you type to the right tool — Claude, calculator, web search, shell, app launcher — all behind one hotkey.

```
┌──────────────────────────────────────────────────┐
│  λ  ask anything…                                │
│    ai    deep   = calc   g/gh/yt/w/d   > shell   │
└──────────────────────────────────────────────────┘
```

## Features

- **`?` Claude (Haiku)** — fast answers, copied straight to your clipboard
- **`??` Claude (Sonnet)** — heavier thinking when you need it
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

## Get an API key

https://console.anthropic.com/settings/keys → Create Key → paste during install.

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
export ANTHROPIC_API_KEY=sk-ant-...
export LAUNCH_MODEL_FAST=claude-haiku-4-5-20251001
export LAUNCH_MODEL_DEEP=claude-sonnet-4-6
```

Themes live in `~/.config/ai-rofi-launcher/themes/` — edit `prompt.rasi` and `view.rasi` to taste.

## Examples

| Type                              | What happens                              |
|-----------------------------------|-------------------------------------------|
| `? best way to debug a memleak`   | Claude answers, copied to clipboard       |
| `?? design a rate limiter at 10k rps` | Sonnet for a harder question          |
| `= 1.5 BTC to USD`                | qalc with currency                        |
| `42 * sqrt(7)`                    | auto-detected math                        |
| `g rust async traits`             | Google in your browser                    |
| `gh tokio runtime`                | GitHub code search                        |
| `> df -h /`                       | runs it, popup with output                |
| `https://news.ycombinator.com`    | opens URL                                 |
| `firefox`                         | drun fallback launches the app            |

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
