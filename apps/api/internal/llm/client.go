package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

func New(baseURL, apiKey, model string) *Client {
	transport := &http.Transport{
		// 显式禁用代理；Proxy==nil 时 Go 仍会读 HTTP_PROXY
		Proxy: func(*http.Request) (*url.URL, error) { return nil, nil },
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
		HTTP: &http.Client{
			Timeout:   120 * time.Second,
			Transport: transport,
		},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.BaseURL != "" && c.APIKey != "" && c.Model != ""
}

type Message struct {
	Role             string `json:"role"`
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

type chatRequest struct {
	Model          string    `json:"model"`
	Messages       []Message `json:"messages"`
	Temperature    float64   `json:"temperature"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) ChatJSON(system, user string) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("LLM 未配置：请设置 JA_LLM_BASE_URL / JA_LLM_API_KEY / JA_LLM_MODEL")
	}
	reqBody := chatRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.3,
		ResponseFormat: &struct {
			Type string `json:"type"`
		}{Type: "json_object"},
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("调用中转站失败: %w", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("中转站返回 %d: %s", res.StatusCode, truncate(string(body), 400))
	}
	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("解析中转站响应失败: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("中转站错误: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("中转站未返回内容")
	}
	msg := parsed.Choices[0].Message
	content := strings.TrimSpace(msg.Content)
	if content == "" {
		// 部分推理模型只回 reasoning_content
		content = strings.TrimSpace(msg.ReasoningContent)
	}
	if content == "" {
		return "", fmt.Errorf("中转站返回空内容")
	}
	return content, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
