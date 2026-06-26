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
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/queries"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/rbac"
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
			queries.CommonSetCurrentUserID,
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

		// 首先确定用户是否为 SuperAdmin。该信息在全局上下文和租户上下文中都需要，
		// 因此无论是否带 X-Tenant-ID 都先查询一次。
		var isSuperAdmin bool
		err = conn.QueryRow(c.Request.Context(),
			queries.AdminSelectSuperAdmin,
			userID,
		).Scan(&isSuperAdmin)
		if err != nil {
			logger.Error("failed to lookup superadmin status", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to verify user role"})
			return
		}

		effectiveRole := rbac.RoleFreeUser
		if isSuperAdmin {
			effectiveRole = rbac.RoleSuperAdmin
		}

		// 可选：解析并验证 X-Tenant-ID，并据此调整普通用户的 effective role。
		tenantIDStr := c.GetHeader("X-Tenant-ID")
		if tenantIDStr != "" {
			tenantID, err := uuid.Parse(tenantIDStr)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid X-Tenant-ID"})
				return
			}

			if isSuperAdmin {
				// SuperAdmin 可进入任意租户上下文，不依赖 memberships 表。
				// memberships 表启用 RLS，直接查询会导致非该租户成员的 SuperAdmin 被拒绝。
				if _, err := conn.Exec(c.Request.Context(),
					queries.CommonSetCurrentTenantID,
					tenantID.String(),
				); err != nil {
					logger.Error("failed to set current_tenant_id", zap.Error(err))
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to set tenant context"})
					return
				}
				SetTenant(c, Tenant{
					ID:            tenantID,
					Role:          string(rbac.TenantOwner),
					EffectiveRole: rbac.RoleSuperAdmin,
				})
			} else {
				var membershipRole, planRoleKey string
				err = conn.QueryRow(c.Request.Context(),
					queries.CommonSelectMembershipAndPlan,
					tenantID, userID,
				).Scan(&membershipRole, &planRoleKey)

				if err == pgx.ErrNoRows {
					// X-Tenant-ID 指向的租户当前用户无权访问。
					// 对于不需要租户上下文的接口（如 /me）不应直接拒绝，
					// 继续执行但不设置租户上下文；需要租户的接口由 RequireTenant 拦截。
					logger.Debug("X-Tenant-ID points to a tenant the user is not a member of", zap.String("tenant_id", tenantID.String()))
				} else if err != nil {
					logger.Error("failed to lookup membership", zap.Error(err))
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "membership lookup failed"})
					return
				} else {
					if _, err := conn.Exec(c.Request.Context(),
						queries.CommonSetCurrentTenantID,
						tenantID.String(),
					); err != nil {
						logger.Error("failed to set current_tenant_id", zap.Error(err))
						c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to set tenant context"})
						return
					}

					effectiveRole = rbac.RoleKeyToRole(planRoleKey)
					SetTenant(c, Tenant{
						ID:            tenantID,
						Role:          membershipRole,
						Plan:          planRoleKey,
						EffectiveRole: effectiveRole,
					})
				}
			}
		}

		SetEffectiveRole(c, effectiveRole)
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

// RequireRole 要求当前用户在当前租户中拥有指定成员角色之一。
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

// RequireEffectiveRole 要求当前用户的系统级 effective role 在允许列表内。
func RequireEffectiveRole(allowed ...rbac.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := EffectiveRoleFromContext(c)
		for _, allowedRole := range allowed {
			if role == allowedRole {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

// RequireSubscriptionAtLeast 要求当前用户的订阅级别不低于指定角色。
// SuperAdmin 始终通过。
func RequireSubscriptionAtLeast(minRole rbac.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := EffectiveRoleFromContext(c)
		if !rbac.IsAtLeastSubscription(role, minRole) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "subscription plan insufficient"})
			return
		}
		c.Next()
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
	_, _ = conn.Exec(ctx, queries.CommonResetCurrentUserID)
	_, _ = conn.Exec(ctx, queries.CommonResetCurrentTenantID)
	conn.Release()
}
