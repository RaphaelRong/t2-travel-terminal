package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	projectsync "github.com/t2-travel-terminal/t2-travel-terminal/internal/project"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/queries"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/rbac"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

type subscriptionPricing struct {
	ID             uuid.UUID `json:"id"`
	DurationMonths int       `json:"duration_months"`
	Price          float64   `json:"price"`
	Currency       string    `json:"currency"`
}

type subscriptionResp struct {
	ID            uuid.UUID            `json:"id"`
	Name          string               `json:"name"`
	Slug          *string              `json:"slug,omitempty"`
	PlanID        *uuid.UUID           `json:"plan_id,omitempty"`
	PlanName      string               `json:"plan_name"`
	EffectiveRole rbac.Role            `json:"effective_role"`
	Pricing       *subscriptionPricing `json:"pricing,omitempty"`
	Role          string               `json:"role"`
	SubscribedAt  *time.Time           `json:"subscribed_at,omitempty"`
	ExpiresAt     *time.Time           `json:"expires_at,omitempty"`
	AutoRenew     bool                 `json:"auto_renew"`
	Status        string               `json:"status"`
}

func listTenantsHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)
	ctx := c.Request.Context()

	rows, err := conn.Query(ctx,
		queries.TenantListSubscriptions,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var result []subscriptionResp
	for rows.Next() {
		var s subscriptionResp
		var planID *uuid.UUID
		var planName *string
		var planRoleKey *string
		var pricingID *uuid.UUID
		var durationMonths *int
		var price *float64
		var currency *string
		var subscribedAt *time.Time
		var expiresAt *time.Time

		if err := rows.Scan(
			&s.ID, &s.Name, &s.Slug, &s.Status,
			&subscribedAt, &expiresAt, &s.AutoRenew,
			&planID, &planName, &planRoleKey,
			&pricingID, &durationMonths, &price, &currency,
			&s.Role,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if planID != nil {
			s.PlanID = planID
		}
		if planName != nil {
			s.PlanName = *planName
		}
		if planRoleKey != nil && *planRoleKey != "" {
			s.EffectiveRole = rbac.RoleKeyToRole(*planRoleKey)
		} else if planName != nil {
			// 兼容旧数据：未设置 role_key 时回退到计划名称映射
			s.EffectiveRole = rbac.PlanNameToRole(*planName)
		}
		if pricingID != nil {
			s.Pricing = &subscriptionPricing{
				ID:             *pricingID,
				DurationMonths: *durationMonths,
				Price:          *price,
				Currency:       *currency,
			}
		}
		if subscribedAt != nil {
			s.SubscribedAt = subscribedAt
		}
		if expiresAt != nil {
			s.ExpiresAt = expiresAt
		}

		result = append(result, s)
	}

	c.JSON(http.StatusOK, gin.H{"subscriptions": result})
}

type createTenantRequest struct {
	Slug      string    `json:"slug"`
	PlanID    uuid.UUID `json:"plan_id" binding:"required"`
	PricingID uuid.UUID `json:"pricing_id" binding:"required"`
	AutoRenew bool      `json:"auto_renew"`
}

func createTenantHandler(c *gin.Context) {
	var req createTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)
	ctx := c.Request.Context()

	tx, err := conn.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 校验 pricing 属于 plan，并读取计划名称和时长
	var planName string
	var durationMonths int
	err = tx.QueryRow(ctx,
		queries.TenantValidatePricingForPlan,
		req.PricingID, req.PlanID,
	).Scan(&planName, &durationMonths)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan or pricing"})
		return
	}

	subName := fmt.Sprintf("%s (%d months)", planName, durationMonths)

	var tenantID uuid.UUID
	subscribedAt := time.Now()
	expiresAt := subscribedAt.AddDate(0, durationMonths, 0)

	err = tx.QueryRow(ctx,
		queries.TenantInsertTenant,
		subName, req.Slug, req.PlanID, req.PricingID, subscribedAt, expiresAt, req.AutoRenew, userID,
	).Scan(&tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(ctx,
		queries.TenantInsertOwnerMembership,
		tenantID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":            tenantID,
		"name":          subName,
		"plan_id":       req.PlanID,
		"pricing_id":    req.PricingID,
		"subscribed_at": subscribedAt,
		"expires_at":    expiresAt,
		"auto_renew":    req.AutoRenew,
		"role":          "owner",
	})
}

func getTenantHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	var result struct {
		ID             uuid.UUID  `json:"id"`
		Name           string     `json:"name"`
		Slug           *string    `json:"slug,omitempty"`
		PlanID         *uuid.UUID `json:"plan_id,omitempty"`
		PlanName       *string    `json:"plan_name,omitempty"`
		PricingID      *uuid.UUID `json:"pricing_id,omitempty"`
		DurationMonths *int       `json:"duration_months,omitempty"`
		Price          *float64   `json:"price,omitempty"`
		Status         string     `json:"status"`
		SubscribedAt   *time.Time `json:"subscribed_at,omitempty"`
		ExpiresAt      *time.Time `json:"expires_at,omitempty"`
		AutoRenew      bool       `json:"auto_renew"`
		CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
		CreatedAt      time.Time  `json:"created_at"`
	}

	err := conn.QueryRow(c.Request.Context(),
		queries.TenantSelectByID,
		t.ID,
	).Scan(&result.ID, &result.Name, &result.Slug, &result.Status,
		&result.PlanID, &result.PlanName, &result.PricingID, &result.DurationMonths, &result.Price,
		&result.SubscribedAt, &result.ExpiresAt, &result.AutoRenew,
		&result.CreatedBy, &result.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func updateTenantHandler(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	_, err := conn.Exec(c.Request.Context(),
		queries.TenantUpdate,
		req.Name, req.Slug, t.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription updated"})
}

func deleteTenantHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	_, err := conn.Exec(c.Request.Context(),
		queries.TenantDelete,
		t.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription deleted"})
}

func updateTenantPlanHandler(c *gin.Context) {
	var req struct {
		PlanID    uuid.UUID `json:"plan_id" binding:"required"`
		PricingID uuid.UUID `json:"pricing_id" binding:"required"`
		AutoRenew bool      `json:"auto_renew"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)
	ctx := c.Request.Context()

	tx, err := conn.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var durationMonths int
	err = tx.QueryRow(ctx,
		queries.TenantSelectPricingDuration,
		req.PricingID, req.PlanID,
	).Scan(&durationMonths)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan or pricing"})
		return
	}

	subscribedAt := time.Now()
	expiresAt := subscribedAt.AddDate(0, durationMonths, 0)

	_, err = tx.Exec(ctx,
		queries.TenantUpdatePlan,
		req.PlanID, req.PricingID, subscribedAt, expiresAt, req.AutoRenew, t.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plan_id":       req.PlanID,
		"pricing_id":    req.PricingID,
		"subscribed_at": subscribedAt,
		"expires_at":    expiresAt,
		"auto_renew":    req.AutoRenew,
	})
}

func listProjectsHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	tx, err := conn.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	projects, err := projectsync.NewSyncService().ListAccessibleProjects(c.Request.Context(), tx, t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]projectResp, 0, len(projects))
	for _, project := range projects {
		result = append(result, toProjectResp(project))
	}

	if err := hydrateProjectDetails(c, tx, result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": result})
}

func createProjectHandler(c *gin.Context) {
	var req projectUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	normalizeProjectDefaults(&req)

	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)
	userID := tenant.UserIDFromContext(c)

	tx, err := conn.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	id, err := projectsync.NewSyncService().CreateProject(c.Request.Context(), tx, projectInputFromReq(req, t.ID, "tenant", userID))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}
