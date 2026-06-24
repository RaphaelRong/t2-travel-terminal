package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/auth"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/datastore"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/notification"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
	"go.uber.org/zap"
)

// RegisterRoutes sets up all API routes.
func RegisterRoutes(r *gin.Engine, logger *zap.Logger, pool *datastore.Pool, tm *auth.TokenManager) {
	emailSvc := notification.NewService(logger)
	authHandler := newAuthHandler(pool, tm, emailSvc)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "t2-travel-terminal",
		})
	})

	api := r.Group("/api/v1")
	{
		// 公开接口
		api.POST("/auth/register", authHandler.register)
		api.POST("/auth/login", authHandler.login)
		api.GET("/auth/verify", authHandler.verifyEmail)
		api.GET("/plans", listPlansHandler)

		api.GET("/indices", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "regional indices endpoint (placeholder)",
			})
		})

		api.GET("/mcp/registry", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"servers": []string{
					"hotel-content-mcp",
					"property-connector-mcp",
					"regional-index-mcp",
				},
			})
		})

		// 需要登录的接口
		authGroup := api.Group("")
		authGroup.Use(tenant.Middleware(pool, logger, tm))
		{
			authGroup.GET("/me", getMeHandler)
			authGroup.PUT("/me", updateMeHandler)
			authGroup.DELETE("/me", deleteMeHandler)

			authGroup.POST("/invites/accept", acceptInviteHandler)

			// 租户相关（可在无 X-Tenant-ID 时创建/列出）
			authGroup.GET("/tenants", listTenantsHandler)
			authGroup.POST("/tenants", createTenantHandler)

			// 必须在某个租户上下文里的接口
			tenantScope := authGroup.Group("")
			tenantScope.Use(tenant.RequireTenant())
			{
				// 当前租户信息
				tenantScope.GET("/tenants/current", getTenantHandler)
				tenantScope.PUT("/tenants/current", tenant.RequireRole("owner", "admin"), updateTenantHandler)
				tenantScope.DELETE("/tenants/current", tenant.RequireRole("owner"), deleteTenantHandler)
				tenantScope.PUT("/tenants/current/plan", tenant.RequireRole("owner", "admin"), updateTenantPlanHandler)

				// 成员管理
				tenantScope.GET("/tenants/current/members", listMembersHandler)
				tenantScope.POST("/tenants/current/members", tenant.RequireRole("owner", "admin"), inviteMemberHandler)
				tenantScope.PUT("/tenants/current/members/:user_id/role", tenant.RequireRole("owner", "admin"), updateMemberRoleHandler)
				tenantScope.DELETE("/tenants/current/members/:user_id", tenant.RequireRole("owner", "admin"), removeMemberHandler)

				// 项目管理
				tenantScope.GET("/projects", listProjectsHandler)
				tenantScope.POST("/projects", createProjectHandler)
				tenantScope.GET("/projects/:project_id", getProjectHandler)
				tenantScope.PUT("/projects/:project_id", updateProjectHandler)
				tenantScope.DELETE("/projects/:project_id", deleteProjectHandler)
			}
		}
	}

	logger.Info("API routes registered")
}
