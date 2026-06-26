package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/auth"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/config"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/datastore"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/notification"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
	"go.uber.org/zap"
)

// RegisterRoutes sets up all API routes.
func RegisterRoutes(r *gin.Engine, logger *zap.Logger, pool *datastore.Pool, tm *auth.TokenManager, cfg *config.Config) {
	emailSvc, err := notification.NewService(notification.Config{
		Host:                cfg.SMTPHost,
		Port:                cfg.SMTPPort,
		Username:            cfg.SMTPUsername,
		Password:            cfg.SMTPPassword,
		From:                cfg.SMTPFrom,
		FromName:            cfg.SMTPFromName,
		Insecure:            cfg.SMTPInsecure,
		AuthMethod:          cfg.SMTPAuthMethod,
		VerificationBaseURL: cfg.EmailVerificationBaseURL,
	}, logger)
	if err != nil {
		logger.Warn("email service not configured; verification emails will not be sent", zap.Error(err))
		emailSvc = nil
	}

	authHandler := newAuthHandler(pool, tm, emailSvc, logger)
	plansHandler := newPlansHandler(pool)
	adminHandler := newAdminHandler()
	hubHandler := newHubHandler(cfg)

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
		api.POST("/auth/resend-verification", authHandler.resendVerification)
		api.GET("/plans", plansHandler.listPlansHandler)
		api.GET("/hub/providers", hubHandler.listProviders)
		api.GET("/hub/skills/ticketmaster/manifest", hubHandler.ticketmasterManifest)

		api.GET("/indices", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "regional indices endpoint",
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
			authGroup.GET("/llm-profiles", listLLMProfilesHandler)
			authGroup.POST("/llm-profiles", createLLMProfileHandler)
			authGroup.POST("/llm-profiles/fetch-models", fetchLLMModelsHandler)
			authGroup.PUT("/llm-profiles/:profile_id", updateLLMProfileHandler)
			authGroup.DELETE("/llm-profiles/:profile_id", deleteLLMProfileHandler)

			authGroup.POST("/invites/accept", acceptInviteHandler)

			// SuperAdmin 管理接口
			adminGroup := authGroup.Group("/admin")
			adminGroup.Use(requireSuperAdmin())
			{
				adminGroup.GET("/users", adminHandler.listUsers)
				adminGroup.GET("/plans", adminHandler.listAdminPlans)
				adminGroup.GET("/projects", adminHandler.listSystemProjects)

				adminGroup.POST("/plans", adminHandler.createPlan)
				adminGroup.PUT("/plans/:plan_id", adminHandler.updatePlan)
				adminGroup.DELETE("/plans/:plan_id", adminHandler.deletePlan)
				adminGroup.POST("/projects", adminHandler.createSystemProject)
				adminGroup.PUT("/projects/:project_id", adminHandler.updateSystemProject)
				adminGroup.DELETE("/projects/:project_id", adminHandler.deleteSystemProject)
				adminGroup.GET("/projects/:project_id/capabilities", listProjectCapabilitiesHandler)
				adminGroup.POST("/projects/:project_id/capabilities", createProjectCapabilityHandler)
				adminGroup.PUT("/projects/:project_id/capabilities/:capability_id", updateProjectCapabilityHandler)
				adminGroup.DELETE("/projects/:project_id/capabilities/:capability_id", deleteProjectCapabilityHandler)
				adminGroup.GET("/projects/:project_id/integrations", listProjectIntegrationsHandler)
				adminGroup.POST("/projects/:project_id/integrations", createProjectIntegrationHandler)
				adminGroup.PUT("/projects/:project_id/integrations/:integration_id", updateProjectIntegrationHandler)
				adminGroup.DELETE("/projects/:project_id/integrations/:integration_id", deleteProjectIntegrationHandler)
				adminGroup.POST("/projects/:project_id/integrations/:integration_id/sync", syncProjectIntegrationHandler)

				adminGroup.POST("/plans/:plan_id/pricing", adminHandler.createPricing)
				adminGroup.PUT("/plans/:plan_id/pricing/:pricing_id", adminHandler.updatePricing)
				adminGroup.DELETE("/plans/:plan_id/pricing/:pricing_id", adminHandler.deletePricing)
			}

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

				// 第三方功能 Hub：租户级凭证与 Skill 执行
				tenantScope.GET("/hub/provider-credentials", hubHandler.listProviderCredentials)
				tenantScope.PUT("/hub/provider-credentials/:provider_id", tenant.RequireRole("owner", "admin"), hubHandler.upsertProviderCredential)
				tenantScope.DELETE("/hub/provider-credentials/:provider_id", tenant.RequireRole("owner", "admin"), hubHandler.deleteProviderCredential)
				tenantScope.POST("/hub/skills/ticketmaster/search-events", hubHandler.ticketmasterSearchEvents)

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
				tenantScope.GET("/projects/:project_id/capabilities", listProjectCapabilitiesHandler)
				tenantScope.POST("/projects/:project_id/capabilities", createProjectCapabilityHandler)
				tenantScope.PUT("/projects/:project_id/capabilities/:capability_id", updateProjectCapabilityHandler)
				tenantScope.DELETE("/projects/:project_id/capabilities/:capability_id", deleteProjectCapabilityHandler)
				tenantScope.GET("/projects/:project_id/integrations", listProjectIntegrationsHandler)
				tenantScope.POST("/projects/:project_id/integrations", createProjectIntegrationHandler)
				tenantScope.PUT("/projects/:project_id/integrations/:integration_id", updateProjectIntegrationHandler)
				tenantScope.DELETE("/projects/:project_id/integrations/:integration_id", deleteProjectIntegrationHandler)
				tenantScope.POST("/projects/:project_id/integrations/:integration_id/sync", syncProjectIntegrationHandler)
			}
		}
	}

	logger.Info("API routes registered")
}
