package main

import (
	"bufio"
	"os"
	"strings"
)

func tryMigrate() (Config, bool) {
	f, err := os.Open(LegacyPath())
	if err != nil {
		return Config{}, false
	}
	defer f.Close()

	cfg := DefaultConfig()
	sc := bufio.NewScanner(f)
	found := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.Index(line, "=")
		if eq < 0 {
			continue
		}
		key := line[:eq]
		val := unquote(line[eq+1:])

		switch key {
		case "LAUNCH_PROVIDER":
			cfg.DefaultProvider = val
			found = true
		case "ANTHROPIC_API_KEY":
			p := cfg.Providers["anthropic"]
			p.APIKey = val
			cfg.Providers["anthropic"] = p
			found = true
		case "OPENAI_API_KEY":
			p := cfg.Providers["openai"]
			p.APIKey = val
			cfg.Providers["openai"] = p
			found = true
		case "GROQ_API_KEY":
			p := cfg.Providers["groq"]
			p.APIKey = val
			cfg.Providers["groq"] = p
			found = true
		case "OLLAMA_API_KEY":
			p := cfg.Providers["ollama_cloud"]
			p.APIKey = val
			cfg.Providers["ollama_cloud"] = p
			found = true
		case "LAUNCH_ANTHROPIC_FAST", "LAUNCH_MODEL_FAST":
			p := cfg.Providers["anthropic"]
			p.Fast = val
			cfg.Providers["anthropic"] = p
		case "LAUNCH_ANTHROPIC_DEEP", "LAUNCH_MODEL_DEEP":
			p := cfg.Providers["anthropic"]
			p.Deep = val
			cfg.Providers["anthropic"] = p
		case "LAUNCH_OPENAI_FAST":
			p := cfg.Providers["openai"]
			p.Fast = val
			cfg.Providers["openai"] = p
		case "LAUNCH_OPENAI_DEEP":
			p := cfg.Providers["openai"]
			p.Deep = val
			cfg.Providers["openai"] = p
		case "LAUNCH_GROQ_FAST":
			p := cfg.Providers["groq"]
			p.Fast = val
			cfg.Providers["groq"] = p
		case "LAUNCH_GROQ_DEEP":
			p := cfg.Providers["groq"]
			p.Deep = val
			cfg.Providers["groq"] = p
		case "LAUNCH_OLLAMA_URL":
			p := cfg.Providers["ollama"]
			p.URL = val
			cfg.Providers["ollama"] = p
		case "LAUNCH_OLLAMA_FAST":
			p := cfg.Providers["ollama"]
			p.Fast = val
			cfg.Providers["ollama"] = p
		case "LAUNCH_OLLAMA_DEEP":
			p := cfg.Providers["ollama"]
			p.Deep = val
			cfg.Providers["ollama"] = p
		case "LAUNCH_OLLAMA_CLOUD_URL":
			p := cfg.Providers["ollama_cloud"]
			p.URL = val
			cfg.Providers["ollama_cloud"] = p
		case "LAUNCH_OLLAMA_CLOUD_FAST":
			p := cfg.Providers["ollama_cloud"]
			p.Fast = val
			cfg.Providers["ollama_cloud"] = p
		case "LAUNCH_OLLAMA_CLOUD_DEEP":
			p := cfg.Providers["ollama_cloud"]
			p.Deep = val
			cfg.Providers["ollama_cloud"] = p
		}
	}
	return cfg, found
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
