package tenant

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/auth"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/datastore"
	"go.uber.org/zap"
)

// Middleware 解析当前用户身份（JWT 或 X-User-ID 开发头），
// 验证可选的 X-Tenant-ID，并将已设置 RLS 会话变量的数据库连接注入 gin 上下文。
func Middleware(pool *datastore.Pool, logger *zap.Logger, tm *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := resolveUserID(c, tm)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		conn, err := pool.Acquire(c.Request.Context())
		if err != nil {
			logger.Error("failed to acquire db connection", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "database unavailable"})
			return
		}

		// 设置会话级 RLS 变量（当前用户）
		if _, err := conn.Exec(c.Request.Context(),
			"SELECT set_config('app.current_user_id', $1, false)",
			userID.String(),
		); err != nil {
			resetAndRelease(conn)
			logger.Error("failed to set current_user_id", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to set user context"})
			return
		}

		// 请求结束后重置会话变量并归还连接，避免污染连接池中的下一个请求
		defer resetAndRelease(conn)

		SetUserID(c, userID)
		SetConn(c, conn)

		// 可选：解析并验证 X-Tenant-ID
		tenantIDStr := c.GetHeader("X-Tenant-ID")
		if tenantIDStr != "" {
			tenantID, err := uuid.Parse(tenantIDStr)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid X-Tenant-ID"})
				return
			}

			var role string
			err = conn.QueryRow(c.Request.Context(),
				"SELECT role FROM memberships WHERE tenant_id = $1 AND user_id = $2",
				tenantID, userID,
			).Scan(&role)

			if err == pgx.ErrNoRows {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "not a member of this tenant"})
				return
			}
			if err != nil {
				logger.Error("failed to lookup membership", zap.Error(err))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "membership lookup failed"})
				return
			}

			if _, err := conn.Exec(c.Request.Context(),
				"SELECT set_config('app.current_tenant_id', $1, false)",
				tenantID.String(),
			); err != nil {
				logger.Error("failed to set current_tenant_id", zap.Error(err))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to set tenant context"})
				return
			}

			SetTenant(c, Tenant{ID: tenantID, Role: role})
		}

		c.Next()
	}
}

// RequireTenant 要求当前请求必须包含有效的 X-Tenant-ID。
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasTenant(c) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "X-Tenant-ID required"})
			return
		}
		c.Next()
	}
}

// RequireRole 要求当前用户在当前租户中拥有指定角色之一。
func RequireRole(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasTenant(c) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "X-Tenant-ID required"})
			return
		}
		t, _ := TenantFromContext(c)
		for _, role := range allowed {
			if t.Role == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

// resolveUserID 优先从 Authorization: Bearer <jwt> 解析用户，
// 否则尝试 X-User-ID 开发头。
func resolveUserID(c *gin.Context, tm *auth.TokenManager) (uuid.UUID, error) {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		return tm.ParseAccessToken(tokenString)
	}

	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid X-User-ID")
		}
		return userID, nil
	}

	return uuid.Nil, fmt.Errorf("missing Authorization or X-User-ID")
}

// resetAndRelease 在连接归还前清空 RLS 会话变量。
func resetAndRelease(conn *pgxpool.Conn) {
	ctx := context.Background()
	_, _ = conn.Exec(ctx, "SELECT set_config('app.current_user_id', '', false)")
	_, _ = conn.Exec(ctx, "SELECT set_config('app.current_tenant_id', '', false)")
	conn.Release()
}
