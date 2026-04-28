package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DeepSeekClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewDeepSeekClient(apiKey, baseURL string) *DeepSeekClient {
	return &DeepSeekClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *DeepSeekClient) Chat(ctx context.Context, prompt string) (string, error) {
	reqBody := chatRequest{
		Model: "deepseek-chat",
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("deepseek 返回错误: %s", string(b))
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("deepseek 返回空结果")
	}

	return chatResp.Choices[0].Message.Content, nil
}
