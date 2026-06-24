package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/api"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/auth"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/config"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/datastore"
	"go.uber.org/zap"
)

// Server wraps the HTTP server and dependencies.
type Server struct {
	cfg    *config.Config
	logger *zap.Logger
	pool   *datastore.Pool
	router *gin.Engine
}

// New creates a new Server instance.
func New(cfg *config.Config, logger *zap.Logger, pool *datastore.Pool) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	tm, err := auth.NewTokenManager(cfg.JWTSecret)
	if err != nil {
		logger.Fatal("failed to create token manager", zap.Error(err))
	}

	api.RegisterRoutes(r, logger, pool, tm)

	return &Server{
		cfg:    cfg,
		logger: logger,
		pool:   pool,
		router: r,
	}
}

// Run starts the HTTP server and gracefully shuts down on context cancellation.
func (s *Server) Run(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:    s.cfg.ServerAddr,
		Handler: s.router,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
