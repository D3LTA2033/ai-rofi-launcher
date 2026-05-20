package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

var (
	bg       = lipgloss.Color("#1e1e2e")
	bgAlt    = lipgloss.Color("#313244")
	fg       = lipgloss.Color("#cdd6f4")
	fgDim    = lipgloss.Color("#a6adc8")
	accent   = lipgloss.Color("#89b4fa")
	accent2  = lipgloss.Color("#cba6f7")
	good     = lipgloss.Color("#a6e3a1")
	warn     = lipgloss.Color("#f9e2af")
	dim      = lipgloss.Color("#6c7086")

	titleStyle    = lipgloss.NewStyle().Foreground(accent2).Bold(true).Padding(0, 1)
	subtitleStyle = lipgloss.NewStyle().Foreground(fgDim).Padding(0, 1)
	labelStyle    = lipgloss.NewStyle().Foreground(fg).Width(24)
	valueStyle    = lipgloss.NewStyle().Foreground(accent)
	setStyle      = lipgloss.NewStyle().Foreground(good)
	unsetStyle    = lipgloss.NewStyle().Foreground(dim).Italic(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(accent2).Bold(true)
	sectionStyle  = lipgloss.NewStyle().Foreground(accent).Bold(true).Underline(true).MarginTop(1)
	frameStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(1, 2)
	helpStyle     = lipgloss.NewStyle().Foreground(dim).MarginTop(1)
	dirtyStyle    = lipgloss.NewStyle().Foreground(warn).Bold(true)
	editStyle     = lipgloss.NewStyle().Foreground(accent2).Bold(true)
)

type fieldKind int

const (
	kindProvider fieldKind = iota
	kindKey
	kindText
)

type field struct {
	section  string
	label    string
	kind     fieldKind
	secret   bool
	get      func(*Config) string
	set      func(*Config, string)
	choices  []string
}

func buildFields() []field {
	pf := func(prov, attr string) (func(*Config) string, func(*Config, string)) {
		return func(c *Config) string {
				p := c.Providers[prov]
				switch attr {
				case "api_key":
					return p.APIKey
				case "url":
					return p.URL
				case "fast":
					return p.Fast
				case "deep":
					return p.Deep
				}
				return ""
			},
			func(c *Config, v string) {
				p := c.Providers[prov]
				switch attr {
				case "api_key":
					p.APIKey = v
				case "url":
					p.URL = v
				case "fast":
					p.Fast = v
				case "deep":
					p.Deep = v
				}
				c.Providers[prov] = p
			}
	}

	out := []field{
		{
			section: "general",
			label:   "Default provider",
			kind:    kindProvider,
			choices: providerOrder,
			get:     func(c *Config) string { return c.DefaultProvider },
			set:     func(c *Config, v string) { c.DefaultProvider = v },
		},
	}
	for _, p := range providerOrder {
		label := providerLabels[p]
		if p != "ollama" {
			g, s := pf(p, "api_key")
			out = append(out, field{section: label, label: "API key", kind: kindKey, secret: true, get: g, set: s})
		}
		if p == "ollama" || p == "ollama_cloud" {
			g, s := pf(p, "url")
			out = append(out, field{section: label, label: "URL", kind: kindText, get: g, set: s})
		}
		gf, sf := pf(p, "fast")
		gd, sd := pf(p, "deep")
		out = append(out,
			field{section: label, label: "Fast model", kind: kindText, get: gf, set: sf},
			field{section: label, label: "Deep model", kind: kindText, get: gd, set: sd},
		)
	}
	return out
}

type model struct {
	cfg     Config
	orig    Config
	fields  []field
	cursor  int
	editing bool
	input   textinput.Model
	dirty   bool
	status  string
	width   int
	height  int
	picker  bool
	pickIdx int
}

func initialModel() (model, error) {
	cfg, err := Load()
	if err != nil {
		return model{}, err
	}
	ti := textinput.New()
	ti.Prompt = "› "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(accent2)
	ti.TextStyle = lipgloss.NewStyle().Foreground(fg)
	ti.CharLimit = 256
	ti.Width = 50
	return model{
		cfg:    cfg,
		orig:   deepCopy(cfg),
		fields: buildFields(),
		input:  ti,
	}, nil
}

func deepCopy(c Config) Config {
	out := Config{DefaultProvider: c.DefaultProvider, Providers: map[string]Provider{}}
	for k, v := range c.Providers {
		out.Providers[k] = v
	}
	return out
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.picker {
			return m.updatePicker(msg)
		}
		if m.editing {
			return m.updateEditing(msg)
		}
		return m.updateList(msg)
	}
	return m, nil
}

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.fields)-1 {
			m.cursor++
		}
	case "g":
		m.cursor = 0
	case "G":
		m.cursor = len(m.fields) - 1
	case "enter":
		f := m.fields[m.cursor]
		if f.kind == kindProvider {
			m.picker = true
			cur := f.get(&m.cfg)
			m.pickIdx = 0
			for i, c := range f.choices {
				if c == cur {
					m.pickIdx = i
					break
				}
			}
			return m, nil
		}
		m.editing = true
		m.input.SetValue(f.get(&m.cfg))
		if f.secret {
			m.input.EchoMode = textinput.EchoPassword
			m.input.EchoCharacter = '●'
		} else {
			m.input.EchoMode = textinput.EchoNormal
		}
		m.input.Focus()
		return m, textinput.Blink
	case "c":
		f := m.fields[m.cursor]
		if f.secret {
			f.set(&m.cfg, "")
			m.dirty = !cfgEqual(m.cfg, m.orig)
			m.status = "cleared " + f.section + " · " + f.label
		}
	case "s":
		if err := m.cfg.Save(); err != nil {
			m.status = "save failed: " + err.Error()
		} else {
			m.orig = deepCopy(m.cfg)
			m.dirty = false
			m.status = "saved to " + ConfigPath()
		}
	}
	return m, nil
}

func (m model) updateEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editing = false
		m.input.Blur()
		return m, nil
	case "enter":
		f := m.fields[m.cursor]
		f.set(&m.cfg, m.input.Value())
		m.editing = false
		m.input.Blur()
		m.dirty = !cfgEqual(m.cfg, m.orig)
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	f := m.fields[m.cursor]
	switch msg.String() {
	case "esc", "q":
		m.picker = false
		return m, nil
	case "up", "k":
		if m.pickIdx > 0 {
			m.pickIdx--
		}
	case "down", "j":
		if m.pickIdx < len(f.choices)-1 {
			m.pickIdx++
		}
	case "enter":
		f.set(&m.cfg, f.choices[m.pickIdx])
		m.picker = false
		m.dirty = !cfgEqual(m.cfg, m.orig)
	}
	return m, nil
}

func cfgEqual(a, b Config) bool {
	if a.DefaultProvider != b.DefaultProvider {
		return false
	}
	if len(a.Providers) != len(b.Providers) {
		return false
	}
	for k, va := range a.Providers {
		vb, ok := b.Providers[k]
		if !ok || va != vb {
			return false
		}
	}
	return true
}

func (m model) View() string {
	var b strings.Builder

	title := titleStyle.Render(" ai-rofi-launcher · config")
	hint := subtitleStyle.Render("YAML at " + ConfigPath())
	if m.dirty {
		title += " " + dirtyStyle.Render("● unsaved")
	}
	b.WriteString(title + "\n")
	b.WriteString(hint + "\n\n")

	if m.picker {
		f := m.fields[m.cursor]
		b.WriteString(sectionStyle.Render("choose " + f.label))
		b.WriteString("\n\n")
		for i, c := range f.choices {
			line := "  " + providerLabels[c]
			if i == m.pickIdx {
				line = cursorStyle.Render("▸ " + providerLabels[c])
			}
			b.WriteString(line + "\n")
		}
		b.WriteString(helpStyle.Render("\n[↑↓] navigate  [↵] select  [esc] cancel"))
		return frameStyle.Render(b.String())
	}

	lastSection := ""
	for i, f := range m.fields {
		if f.section != lastSection {
			b.WriteString(sectionStyle.Render(f.section) + "\n")
			lastSection = f.section
		}
		cur := f.get(&m.cfg)
		val := renderValue(f, cur)
		line := labelStyle.Render("  " + f.label) + " " + val
		if i == m.cursor {
			if m.editing {
				line = labelStyle.Render(editStyle.Render("▸ " + f.label)) + " " + m.input.View()
			} else {
				line = cursorStyle.Render("▸ ") + labelStyle.Render(f.label) + " " + val
			}
		}
		b.WriteString(line + "\n")
	}

	help := "[↑↓] navigate  [↵] edit  [c] clear key  [s] save  [q] quit"
	if m.status != "" {
		help = m.status + "   ·   " + help
	}
	b.WriteString(helpStyle.Render("\n" + help))

	return frameStyle.Render(b.String())
}

func renderValue(f field, v string) string {
	switch f.kind {
	case kindKey:
		if v == "" {
			return unsetStyle.Render("(not set)")
		}
		return setStyle.Render("●●●●●●●● set")
	case kindProvider:
		return valueStyle.Render(providerLabels[v])
	default:
		if v == "" {
			return unsetStyle.Render("(empty)")
		}
		return valueStyle.Render(v)
	}
}

func runTUI() error {
	m, err := initialModel()
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func _unused() string { return fmt.Sprint("") }
