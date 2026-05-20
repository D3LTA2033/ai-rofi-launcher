package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Provider struct {
	APIKey string `yaml:"api_key,omitempty"`
	URL    string `yaml:"url,omitempty"`
	Fast   string `yaml:"fast"`
	Deep   string `yaml:"deep"`
}

type Config struct {
	DefaultProvider string              `yaml:"default_provider"`
	Providers       map[string]Provider `yaml:"providers"`
}

var providerOrder = []string{"anthropic", "openai", "groq", "ollama", "ollama_cloud"}

var providerLabels = map[string]string{
	"anthropic":    "Anthropic",
	"openai":       "OpenAI",
	"groq":         "Groq",
	"ollama":       "Ollama (local)",
	"ollama_cloud": "Ollama Cloud",
}

func DefaultConfig() Config {
	return Config{
		DefaultProvider: "anthropic",
		Providers: map[string]Provider{
			"anthropic": {
				Fast: "claude-haiku-4-5-20251001",
				Deep: "claude-sonnet-4-6",
			},
			"openai": {
				Fast: "gpt-4o-mini",
				Deep: "gpt-4o",
			},
			"groq": {
				Fast: "llama-3.1-8b-instant",
				Deep: "llama-3.3-70b-versatile",
			},
			"ollama": {
				URL:  "http://localhost:11434",
				Fast: "llama3.2:3b",
				Deep: "llama3.1:8b",
			},
			"ollama_cloud": {
				URL:  "https://ollama.com",
				Fast: "gpt-oss:20b",
				Deep: "gpt-oss:120b",
			},
		},
	}
}

func ConfigDir() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "ai-rofi-launcher")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ai-rofi-launcher")
}

func ConfigPath() string  { return filepath.Join(ConfigDir(), "config.yaml") }
func LegacyPath() string  { return filepath.Join(ConfigDir(), "config") }

func Load() (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			if mig, mok := tryMigrate(); mok {
				return mig, nil
			}
			return cfg, nil
		}
		return cfg, err
	}
	var loaded Config
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		return cfg, err
	}
	mergeDefaults(&loaded, cfg)
	return loaded, nil
}

func mergeDefaults(c *Config, d Config) {
	if c.DefaultProvider == "" {
		c.DefaultProvider = d.DefaultProvider
	}
	if c.Providers == nil {
		c.Providers = map[string]Provider{}
	}
	for k, dp := range d.Providers {
		p, ok := c.Providers[k]
		if !ok {
			c.Providers[k] = dp
			continue
		}
		if p.Fast == "" {
			p.Fast = dp.Fast
		}
		if p.Deep == "" {
			p.Deep = dp.Deep
		}
		if p.URL == "" && dp.URL != "" {
			p.URL = dp.URL
		}
		c.Providers[k] = p
	}
}

func (c Config) Save() error {
	if err := os.MkdirAll(ConfigDir(), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	header := []byte("# ai-rofi-launcher · YAML config\n# Run `launch --config` to edit interactively.\n\n")
	if err := os.WriteFile(ConfigPath(), append(header, data...), 0o600); err != nil {
		return err
	}
	return nil
}

func (c Config) Export() string {
	out := ""
	add := func(k, v string) {
		if v == "" {
			return
		}
		out += fmt.Sprintf("%s=%s\n", k, shellQuote(v))
	}
	addKey := func(envName, v string) {
		if v == "" {
			return
		}
		out += fmt.Sprintf("export %s=%s\n", envName, shellQuote(v))
	}
	out += fmt.Sprintf("LAUNCH_PROVIDER=%s\n", shellQuote(c.DefaultProvider))
	a := c.Providers["anthropic"]
	addKey("ANTHROPIC_API_KEY", a.APIKey)
	add("LAUNCH_ANTHROPIC_FAST", a.Fast)
	add("LAUNCH_ANTHROPIC_DEEP", a.Deep)
	o := c.Providers["openai"]
	addKey("OPENAI_API_KEY", o.APIKey)
	add("LAUNCH_OPENAI_FAST", o.Fast)
	add("LAUNCH_OPENAI_DEEP", o.Deep)
	g := c.Providers["groq"]
	addKey("GROQ_API_KEY", g.APIKey)
	add("LAUNCH_GROQ_FAST", g.Fast)
	add("LAUNCH_GROQ_DEEP", g.Deep)
	l := c.Providers["ollama"]
	add("LAUNCH_OLLAMA_URL", l.URL)
	add("LAUNCH_OLLAMA_FAST", l.Fast)
	add("LAUNCH_OLLAMA_DEEP", l.Deep)
	cl := c.Providers["ollama_cloud"]
	addKey("OLLAMA_API_KEY", cl.APIKey)
	add("LAUNCH_OLLAMA_CLOUD_URL", cl.URL)
	add("LAUNCH_OLLAMA_CLOUD_FAST", cl.Fast)
	add("LAUNCH_OLLAMA_CLOUD_DEEP", cl.Deep)
	return out
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	safe := true
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' ||
			r == '-' || r == '_' || r == '.' || r == '/' || r == ':' || r == '@' || r == '+' || r == '=') {
			safe = false
			break
		}
	}
	if safe {
		return s
	}
	q := []rune{'\''}
	for _, r := range s {
		if r == '\'' {
			q = append(q, '\'', '\\', '\'', '\'')
		} else {
			q = append(q, r)
		}
	}
	q = append(q, '\'')
	return string(q)
}
