package runtime

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/god"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/llm"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/store"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/tools"
)

// Runner 是 Agent 运行循环的 orchestrator。
type Runner struct {
	llmFactory    *llm.ClientFactory
	memoryStore   store.MemoryStore
	userdataStore store.UserDataStore
	sessionStore  store.SessionStore
	projectStore  store.ProjectStore
	guardrail     *Guardrail
	planner       *Planner
	projectLoader *ProjectLoader
	localBaseURL  string
}

// NewRunner 创建 Runner。localBaseURL 用于执行指向本机 API 的 Skill（如 /api/v1/hub/...）。
func NewRunner(localBaseURL string) *Runner {
	projectStore := store.NewPGProjectStore()
	return &Runner{
		llmFactory:    llm.NewClientFactory(),
		memoryStore:   store.NewPGMemoryStore(),
		userdataStore: store.NewPGUserDataStore(),
		sessionStore:  store.NewPGSessionStore(),
		projectStore:  projectStore,
		guardrail:     NewGuardrail(),
		planner:       NewPlanner(),
		projectLoader: NewProjectLoader(projectStore, localBaseURL),
		localBaseURL:  localBaseURL,
	}
}

// RunRequest 是运行一次 Agent 的请求。
type RunRequest struct {
	UserID        uuid.UUID
	TenantID      *uuid.UUID
	SessionID     uuid.UUID
	Profile       *domain.UserProfile
	GodScope      *domain.GodScope
	Soul          *domain.Soul
	Messages      []*domain.Message
	LLMProfileID  uuid.UUID
	OverrideModel string
}

// RunResponse 是 Agent 运行的结果。
type RunResponse struct {
	UserMessage      *domain.Message
	AssistantMessage *domain.Message
	ToolMessages     []*domain.Message
	Error            error
}

// Run 执行一次 Agent Run。若 stream 不为 nil，会在关键节点推送 StreamEvent。
func (r *Runner) Run(ctx context.Context, q store.Querier, req RunRequest, stream chan<- StreamEvent) (*RunResponse, error) {
	// 1. 加载 LLM profile 并创建 provider
	profile, err := r.llmFactory.LoadProfile(ctx, q, req.UserID, req.LLMProfileID)
	if err != nil {
		r.emit(stream, StreamEventError, map[string]interface{}{"error": fmt.Sprintf("load llm profile: %v", err)})
		return nil, fmt.Errorf("load llm profile: %w", err)
	}
	provider, err := r.llmFactory.CreateProvider(profile)
	if err != nil {
		r.emit(stream, StreamEventError, map[string]interface{}{"error": fmt.Sprintf("create llm provider: %v", err)})
		return nil, fmt.Errorf("create llm provider: %w", err)
	}

	// 2. 构建 AgentRun
	maxIter := 30
	if req.GodScope != nil && req.GodScope.MaxIterations > 0 {
		maxIter = req.GodScope.MaxIterations
	}

	run := &domain.AgentRun{
		SessionID:  req.SessionID,
		UserID:     req.UserID,
		TenantID:   req.TenantID,
		Profile:    req.Profile,
		Soul:       req.Soul,
		GodScope:   req.GodScope,
		Budget:     &domain.IterationBudget{MaxTotal: maxIter},
		Guardrails: domain.NewGuardrailController(),
		State:      domain.RunStateExecuting,
	}

	// 3. 加载工具
	registry := tools.NewRegistry(req.GodScope)
	registry.MustRegister(tools.NewWorkflowTool())
	registry.MustRegister(tools.NewMemoryTool(r.memoryStore))
	registry.MustRegister(tools.NewExecuteSQLTool(r.userdataStore))
	registry.MustRegister(tools.NewDoneTool())

	// 3.1 加载会话关联的 Project Capability 工具
	projectTools, err := r.projectLoader.LoadTools(ctx, q, req.SessionID)
	if err != nil {
		r.emit(stream, StreamEventError, map[string]interface{}{"error": fmt.Sprintf("load project tools: %v", err)})
		return nil, fmt.Errorf("load project tools: %w", err)
	}
	for _, pt := range projectTools {
		if err := registry.Register(pt); err != nil {
			// 被 GodScope 禁止或其他原因注册失败时，记录但继续
			r.emit(stream, StreamEventError, map[string]interface{}{"error": fmt.Sprintf("register project tool %s: %v", pt.Name(), err)})
		}
	}

	// 4. 构建 LLM messages，注入 system prompt
	llmMessages := r.buildLLMMessages(req.GodScope, req.Soul, req.Messages)

	// 5. 主循环
	var assistantMsg *domain.Message
	var toolMessages []*domain.Message

	for run.Budget.Remaining() > 0 {
		if !run.Budget.Consume() {
			break
		}

		model := profile.DefaultModel
		if req.OverrideModel != "" {
			model = req.OverrideModel
		}

		resp, err := provider.Chat(ctx, model, llmMessages, registry.ToLLMDefinitions())
		if err != nil {
			r.emit(stream, StreamEventError, map[string]interface{}{"error": fmt.Sprintf("llm chat: %v", err)})
			return nil, fmt.Errorf("llm chat: %w", err)
		}

		// 记录并保存 assistant message
		assistantMsg = &domain.Message{
			SessionID:        req.SessionID,
			Role:             domain.MessageRoleAssistant,
			Content:          resp.Content,
			ToolCalls:        resp.ToolCalls,
			ReasoningContent: resp.ReasoningContent,
		}
		if err := r.sessionStore.CreateMessage(ctx, q, assistantMsg); err != nil {
			r.emit(stream, StreamEventError, map[string]interface{}{"error": fmt.Sprintf("save assistant message: %v", err)})
			return nil, fmt.Errorf("save assistant message: %w", err)
		}
		r.emit(stream, StreamEventAssistantMessage, map[string]interface{}{
			"id":                assistantMsg.ID.String(),
			"role":              string(assistantMsg.Role),
			"content":           assistantMsg.Content,
			"tool_calls":        assistantMsg.ToolCalls,
			"reasoning_content": assistantMsg.ReasoningContent,
		})

		if len(resp.ToolCalls) == 0 {
			run.State = domain.RunStateCompleted
			break
		}

		// 执行工具调用并保存/推送结果
		toolResults := r.executeToolCalls(ctx, q, registry, run, resp.ToolCalls, stream)
		toolMessages = append(toolMessages, toolResults...)

		// 更新 LLM messages
		llmMessages = append(llmMessages, llm.FromDomainMessage(assistantMsg))
		for _, tm := range toolResults {
			llmMessages = append(llmMessages, llm.FromDomainMessage(tm))
		}

		if run.State == domain.RunStateCompleted {
			break
		}
	}

	budgetReached := run.State != domain.RunStateCompleted && assistantMsg != nil
	if budgetReached {
		assistantMsg.Content += "\n\n[Iteration budget reached]"
	}

	r.emit(stream, StreamEventDone, map[string]interface{}{
		"state":          run.State,
		"budget_reached": budgetReached,
		"tool_count":     len(toolMessages),
	})

	return &RunResponse{
		AssistantMessage: assistantMsg,
		ToolMessages:     toolMessages,
	}, nil
}

func (r *Runner) buildLLMMessages(scope *domain.GodScope, soul *domain.Soul, history []*domain.Message) []llm.Message {
	llmMessages := make([]llm.Message, 0, len(history)+1)

	systemPrompt := god.BuildSystemPrompt(scope, soul)
	if systemPrompt != "" {
		llmMessages = append(llmMessages, llm.Message{
			Role:    llm.RoleSystem,
			Content: systemPrompt,
		})
	}

	for _, m := range history {
		llmMessages = append(llmMessages, llm.FromDomainMessage(m))
	}
	return llmMessages
}

func (r *Runner) executeToolCalls(ctx context.Context, q store.Querier, registry *tools.Registry, run *domain.AgentRun, calls []domain.ToolCall, stream chan<- StreamEvent) []*domain.Message {
	var results []*domain.Message

	for _, tc := range calls {
		tool, ok := registry.Get(tc.Function.Name)
		if !ok {
			msg := &domain.Message{
				SessionID:  run.SessionID,
				Role:       domain.MessageRoleTool,
				ToolCallID: tc.ID,
				ToolName:   tc.Function.Name,
				ToolResult: map[string]interface{}{"error": fmt.Sprintf("tool %s not found", tc.Function.Name)},
			}
			_ = r.sessionStore.CreateMessage(ctx, q, msg)
			r.emit(stream, StreamEventToolMessage, map[string]interface{}{"tool_message": msg})
			results = append(results, msg)
			continue
		}

		var args map[string]interface{}
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)

		// GodScope 校验
		if run.GodScope != nil {
			if err := run.GodScope.CanExecute(tool.Name(), tool.Domain()); err != nil {
				msg := &domain.Message{
					SessionID:  run.SessionID,
					Role:       domain.MessageRoleTool,
					ToolCallID: tc.ID,
					ToolName:   tc.Function.Name,
					ToolResult: map[string]interface{}{"error": err.Error()},
				}
				_ = r.sessionStore.CreateMessage(ctx, q, msg)
				r.emit(stream, StreamEventToolMessage, map[string]interface{}{"tool_message": msg})
				results = append(results, msg)
				continue
			}
		}

		runCtx := &tools.Context{
			UserID:    run.UserID,
			TenantID:  run.TenantID,
			SessionID: run.SessionID,
			Run:       run,
			Querier:   q,
		}

		res, err := tool.Execute(ctx, args, runCtx)
		failed := err != nil
		var resultMap map[string]interface{}
		if failed {
			resultMap = map[string]interface{}{"error": err.Error()}
		} else {
			resultMap = normalizeResult(res)
		}

		// 护栏检查
		decision := r.guardrail.AfterCall(run.Guardrails, tool.Name(), args, res, failed)
		if decision.Action == "halt" {
			resultMap["guardrail"] = decision.Message
			run.State = domain.RunStateCompleted
		}

		msg := &domain.Message{
			SessionID:  run.SessionID,
			Role:       domain.MessageRoleTool,
			ToolCallID: tc.ID,
			ToolName:   tc.Function.Name,
			ToolResult: resultMap,
		}
		if err := r.sessionStore.CreateMessage(ctx, q, msg); err != nil {
			msg.ToolResult = map[string]interface{}{"error": fmt.Sprintf("save tool message: %v", err)}
		}
		r.emit(stream, StreamEventToolMessage, map[string]interface{}{"tool_message": msg})
		results = append(results, msg)
	}

	return results
}

func normalizeResult(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{"result": v}
}

func (r *Runner) emit(stream chan<- StreamEvent, typ StreamEventType, data map[string]interface{}) {
	if stream == nil {
		return
	}
	select {
	case stream <- StreamEvent{Type: typ, Data: data}:
	default:
		// 消费者未就绪时丢弃事件，避免阻塞运行循环
	}
}

// BuildSystemPrompt 是 god.BuildSystemPrompt 的别名，便于 runtime 使用。
func BuildSystemPrompt(scope *domain.GodScope, soul *domain.Soul) string {
	return god.BuildSystemPrompt(scope, soul)
}
