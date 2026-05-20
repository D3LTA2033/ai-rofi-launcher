package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

//go:embed web
var webFS embed.FS

const systemPrompt = "You are a warm, thoughtful conversational assistant. Be direct but kind. Use plain text — no markdown headers, no preamble. Write like a smart friend talking, not a corporate FAQ. Keep responses conversational length unless the user asks for depth."

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Provider string        `json:"provider,omitempty"`
	Messages []ChatMessage `json:"messages"`
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	_ = cmd.Start()
}

func runServe(port int) error {
	if port == 0 {
		port = 8765
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := fmt.Sprintf("http://%s", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("server already running — opening " + url)
		openBrowser(url)
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/api/config", handleConfig)
	mux.HandleFunc("/api/chat", handleChat)

	go func() {
		time.Sleep(250 * time.Millisecond)
		fmt.Println("Talkroom running at " + url)
		openBrowser(url)
	}()

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}
	return srv.Serve(listener)
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := webFS.ReadFile("web/index.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := Load()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	provs := []string{}
	models := map[string]string{}
	for _, name := range providerOrder {
		p := cfg.Providers[name]
		if name == "ollama" || p.APIKey != "" {
			provs = append(provs, name)
			models[name] = p.Fast
		}
	}
	if len(provs) == 0 {
		provs = providerOrder
		for _, name := range providerOrder {
			models[name] = cfg.Providers[name].Fast
		}
	}
	def := cfg.DefaultProvider
	have := false
	for _, p := range provs {
		if p == def {
			have = true
			break
		}
	}
	if !have && len(provs) > 0 {
		def = provs[0]
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"providers": provs,
		"default":   def,
		"models":    models,
	})
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	cfg, err := Load()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	provider := req.Provider
	if provider == "" {
		provider = cfg.DefaultProvider
	}
	p, ok := cfg.Providers[provider]
	if !ok {
		http.Error(w, "unknown provider: "+provider, 400)
		return
	}
	model := p.Fast
	if model == "" {
		http.Error(w, "no fast model configured for "+provider, 400)
		return
	}

	reply, err := callProvider(provider, p, model, req.Messages)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{{"type": "text", "text": "error: " + err.Error()}},
		})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{
		"content": []map[string]any{{"type": "text", "text": reply}},
	})
}

func callProvider(name string, p Provider, model string, msgs []ChatMessage) (string, error) {
	switch name {
	case "anthropic":
		return callAnthropic(p.APIKey, model, msgs)
	case "openai":
		return callOpenAICompat("https://api.openai.com/v1/chat/completions", p.APIKey, model, msgs)
	case "groq":
		return callOpenAICompat("https://api.groq.com/openai/v1/chat/completions", p.APIKey, model, msgs)
	case "ollama":
		url := p.URL
		if url == "" {
			url = "http://localhost:11434"
		}
		return callOllama(url, "", model, msgs)
	case "ollama_cloud":
		url := p.URL
		if url == "" {
			url = "https://ollama.com"
		}
		return callOllama(url, p.APIKey, model, msgs)
	}
	return "", fmt.Errorf("unknown provider")
}

func httpPost(url string, headers map[string]string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func callAnthropic(key, model string, msgs []ChatMessage) (string, error) {
	if key == "" {
		return "", fmt.Errorf("no ANTHROPIC_API_KEY configured")
	}
	body, _ := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 1024,
		"system":     systemPrompt,
		"messages":   msgs,
	})
	data, err := httpPost("https://api.anthropic.com/v1/messages", map[string]string{
		"x-api-key":         key,
		"anthropic-version": "2023-06-01",
		"content-type":      "application/json",
	}, body)
	if err != nil {
		return "", err
	}
	var r struct {
		Content []struct {
			Type, Text string
		} `json:"content"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return "", fmt.Errorf("decode: %s", string(data))
	}
	if r.Error.Message != "" {
		return "", fmt.Errorf(r.Error.Message)
	}
	for _, c := range r.Content {
		if c.Type == "text" {
			return c.Text, nil
		}
	}
	return "", fmt.Errorf("no text in response")
}

func callOpenAICompat(url, key, model string, msgs []ChatMessage) (string, error) {
	if key == "" {
		return "", fmt.Errorf("missing API key")
	}
	full := append([]ChatMessage{{Role: "system", Content: systemPrompt}}, msgs...)
	body, _ := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 1024,
		"messages":   full,
	})
	data, err := httpPost(url, map[string]string{
		"Authorization": "Bearer " + key,
		"content-type":  "application/json",
	}, body)
	if err != nil {
		return "", err
	}
	var r struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return "", fmt.Errorf("decode: %s", string(data))
	}
	if r.Error.Message != "" {
		return "", fmt.Errorf(r.Error.Message)
	}
	if len(r.Choices) > 0 {
		return r.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no choices in response")
}

func callOllama(baseURL, key, model string, msgs []ChatMessage) (string, error) {
	full := append([]ChatMessage{{Role: "system", Content: systemPrompt}}, msgs...)
	body, _ := json.Marshal(map[string]any{
		"model":    model,
		"stream":   false,
		"messages": full,
	})
	headers := map[string]string{"content-type": "application/json"}
	if key != "" {
		headers["Authorization"] = "Bearer " + key
	}
	data, err := httpPost(baseURL+"/api/chat", headers, body)
	if err != nil {
		return "", err
	}
	var r struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return "", fmt.Errorf("decode: %s", string(data))
	}
	if r.Error != "" {
		return "", fmt.Errorf(r.Error)
	}
	return r.Message.Content, nil
}
