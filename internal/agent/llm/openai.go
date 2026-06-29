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

// OpenAIProvider 支持 OpenAI-compatible API（包括 custom base_url）。
type OpenAIProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewOpenAIProvider 创建 OpenAI provider。
func NewOpenAIProvider(baseURL, apiKey string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

type openAIChatRequest struct {
	Model    string              `json:"model"`
	Messages []openAIMessage     `json:"messages"`
	Tools    []openAITool        `json:"tools,omitempty"`
}

type openAIMessage struct {
	Role       string              `json:"role"`
	Content    string              `json:"content,omitempty"`
	ToolCalls  []openAIToolCall    `json:"tool_calls,omitempty"`
	ToolCallID string              `json:"tool_call_id,omitempty"`
	Name       string              `json:"name,omitempty"`
}

type openAIToolCall struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	Function openAIFunctionCall   `json:"function"`
}

type openAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAITool struct {
	Type     string              `json:"type"`
	Function openAIToolFunction   `json:"function"`
}

type openAIToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Role         string           `json:"role"`
			Content      string           `json:"content"`
			ToolCalls    []openAIToolCall `json:"tool_calls"`
			Reasoning    string           `json:"reasoning,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (p *OpenAIProvider) Chat(ctx context.Context, model string, messages []Message, tools []ToolDefinition) (*Response, error) {
	reqBody := openAIChatRequest{
		Model:    model,
		Messages: toOpenAIMessages(messages),
	}
	if len(tools) > 0 {
		reqBody.Tools = toOpenAITools(tools)
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

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

	var decoded openAIChatResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("invalid LLM response: %w", err)
	}
	if decoded.Error != nil {
		return nil, fmt.Errorf("LLM error: %s", decoded.Error.Message)
	}
	if len(decoded.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	choice := decoded.Choices[0]
	return &Response{
		Content:          choice.Message.Content,
		ToolCalls:        fromOpenAIToolCalls(choice.Message.ToolCalls),
		ReasoningContent: choice.Message.Reasoning,
		FinishReason:     choice.FinishReason,
		Usage: Usage{
			PromptTokens:     decoded.Usage.PromptTokens,
			CompletionTokens: decoded.Usage.CompletionTokens,
			TotalTokens:      decoded.Usage.TotalTokens,
		},
	}, nil
}

func toOpenAIMessages(messages []Message) []openAIMessage {
	result := make([]openAIMessage, len(messages))
	for i, m := range messages {
		msg := openAIMessage{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		if m.Role == RoleTool {
			msg.Role = "tool"
			msg.Content = toolResultToString(m.ToolResult)
		}
		if len(m.ToolCalls) > 0 {
			msg.ToolCalls = toOpenAIToolCalls(m.ToolCalls)
		}
		result[i] = msg
	}
	return result
}

func toOpenAITools(tools []ToolDefinition) []openAITool {
	result := make([]openAITool, len(tools))
	for i, t := range tools {
		result[i] = openAITool{
			Type: "function",
			Function: openAIToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		}
	}
	return result
}

func toOpenAIToolCalls(calls []domain.ToolCall) []openAIToolCall {
	result := make([]openAIToolCall, len(calls))
	for i, tc := range calls {
		result[i] = openAIToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: openAIFunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	return result
}

func fromOpenAIToolCalls(calls []openAIToolCall) []domain.ToolCall {
	result := make([]domain.ToolCall, len(calls))
	for i, tc := range calls {
		result[i] = domain.ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			Function: domain.ToolCallFunction{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		}
	}
	return result
}

func toolResultToString(result map[string]any) string {
	if result == nil {
		return ""
	}
	b, _ := json.Marshal(result)
	return string(b)
}
