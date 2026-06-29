package runtime

// StreamEventType 表示流式事件类型。
type StreamEventType string

const (
	// StreamEventAssistantMessage 表示 assistant 产生了一条消息。
	StreamEventAssistantMessage StreamEventType = "assistant_message"
	// StreamEventToolMessage 表示执行了一条工具并产生结果。
	StreamEventToolMessage StreamEventType = "tool_message"
	// StreamEventDone 表示 Agent Run 结束。
	StreamEventDone StreamEventType = "done"
	// StreamEventError 表示运行过程中出现错误。
	StreamEventError StreamEventType = "error"
)

// StreamEvent 是 SSE 推送的事件结构。
type StreamEvent struct {
	Type StreamEventType        `json:"type"`
	Data map[string]interface{} `json:"data"`
}
