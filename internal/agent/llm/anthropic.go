package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
)

// AnthropicProvider 支持 Anthropic Messages API。
type AnthropicProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewAnthropicProvider 创建 Anthropic provider。
func NewAnthropicProvider(baseURL, apiKey string) *AnthropicProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	return &AnthropicProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

type anthropicChatRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
	System    string             `json:"system,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type anthropicChatResponse struct {
	Content []struct {
		Type  string         `json:"type"`
		Text  string         `json:"text,omitempty"`
		ID    string         `json:"id,omitempty"`
		Name  string         `json:"name,omitempty"`
		Input map[string]any `json:"input,omitempty"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (p *AnthropicProvider) Chat(ctx context.Context, model string, messages []Message, tools []ToolDefinition) (*Response, error) {
	var system string
	var chatMessages []anthropicMessage

	for _, m := range messages {
		switch m.Role {
		case RoleSystem:
			system += m.Content + "\n"
		case RoleUser, RoleAssistant:
			chatMessages = append(chatMessages, anthropicMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		case RoleTool:
			chatMessages = append(chatMessages, anthropicMessage{
				Role:    "user",
				Content: fmt.Sprintf("Tool %s result: %s", m.ToolName, toolResultToString(m.ToolResult)),
			})
		}
	}

	reqBody := anthropicChatRequest{
		Model:     model,
		MaxTokens: 4096,
		Messages:  chatMessages,
		System:    strings.TrimSpace(system),
	}
	if len(tools) > 0 {
		reqBody.Tools = toAnthropicTools(tools)
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/messages", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("LLM returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var decoded anthropicChatResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("invalid LLM response: %w", err)
	}
	if decoded.Error != nil {
		return nil, fmt.Errorf("LLM error: %s", decoded.Error.Message)
	}

	var content string
	var toolCalls []domain.ToolCall
	for _, block := range decoded.Content {
		switch block.Type {
		case "text":
			content += block.Text
		case "tool_use":
			args, _ := json.Marshal(block.Input)
			toolCalls = append(toolCalls, domain.ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: domain.ToolCallFunction{
					Name:      block.Name,
					Arguments: string(args),
				},
			})
		}
	}

	finishReason := FinishReasonStop
	switch decoded.StopReason {
	case "tool_use":
		finishReason = FinishReasonToolCalls
	case "max_tokens":
		finishReason = FinishReasonLength
	}

	return &Response{
		Content:      content,
		ToolCalls:    toolCalls,
		FinishReason: finishReason,
		Usage: Usage{
			PromptTokens:     decoded.Usage.InputTokens,
			CompletionTokens: decoded.Usage.OutputTokens,
			TotalTokens:      decoded.Usage.InputTokens + decoded.Usage.OutputTokens,
		},
	}, nil
}

func toAnthropicTools(tools []ToolDefinition) []anthropicTool {
	result := make([]anthropicTool, len(tools))
	for i, t := range tools {
		result[i] = anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		}
	}
	return result
}
