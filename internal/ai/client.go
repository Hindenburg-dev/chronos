package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client struct {
	platform string
	apiKey   string
	model    string
	client   *http.Client
}

func NewClient(platform, apiKey, model string) *Client {
	return &Client{
		platform: platform,
		apiKey:   apiKey,   
		model:    model,
		client:   &http.Client{Timeout: 60 * time.Second},  
	}
}

func (c *Client) Chat(messages []Message) (string, error) {
	reqBody := map[string]any{
		"model":    c.model,
		"messages": messages,
	}
	jsonData, _ := json.Marshal(reqBody)

	url := "https://api.deepseek.com/chat/completions"

	switch c.platform {
	case "ChatGPT", "OpenAI":
		url = "https://api.openai.com/v1/chat/completions"
	case "Gemini":
		url = "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"
	case "Claude":
		url = "https://api.anthropic.com/v1/messages"
	case "Kimi":
		url = "https://api.moonshot.cn/v1/chat/completions"
	case "Qwen":
		url = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	case "Deepseek":
		url = "https://api.deepseek.com/chat/completions"
	}

	
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
  
	json.Unmarshal(body, &result)
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("AI 返回异常: %s", string(body))
	}
	return result.Choices[0].Message.Content, nil
}  