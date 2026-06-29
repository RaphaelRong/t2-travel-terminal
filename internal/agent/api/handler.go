package api

import (
	"strings"

	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/god"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/runtime"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/store"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/config"
)

// Handler 聚合 Agent 相关的所有 API 处理函数。
type Handler struct {
	soulStore     store.SoulStore
	memoryStore   store.MemoryStore
	sessionStore  store.SessionStore
	profileStore  store.ProfileStore
	godStore      store.GodConfigStore
	userdataStore store.UserDataStore
	projectStore  store.ProjectStore
	godLoader     *god.Loader
	runner        *runtime.Runner
}

// NewHandler 创建一个新的 Agent Handler。
func NewHandler(cfg *config.Config) *Handler {
	soulStore := store.NewPGSoulStore()
	godStore := store.NewPGGodConfigStore()
	localBaseURL := inferLocalBaseURL(cfg.ServerAddr)
	return &Handler{
		soulStore:     soulStore,
		memoryStore:   store.NewPGMemoryStore(),
		sessionStore:  store.NewPGSessionStore(),
		profileStore:  store.NewPGProfileStore(),
		godStore:      godStore,
		userdataStore: store.NewPGUserDataStore(),
		projectStore:  store.NewPGProjectStore(),
		godLoader:     god.NewLoader(godStore, soulStore),
		runner:        runtime.NewRunner(localBaseURL),
	}
}

// inferLocalBaseURL 根据 server_addr 推断本机 Base URL，用于执行指向 /api/v1/ 的 Skill。
func inferLocalBaseURL(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "http://localhost:8080"
	}
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		return "http://" + addr
	}
	return addr
}
