package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	projectsync "github.com/t2-travel-terminal/t2-travel-terminal/internal/project"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/queries"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/rbac"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

var systemTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type adminHandler struct{}

func newAdminHandler() *adminHandler {
	return &adminHandler{}
}

// requireSuperAdmin 要求当前登录用户为系统管理员。
// 统一使用 tenant 上下文中解析出的 effective role 进行校验。
func requireSuperAdmin() gin.HandlerFunc {
	return tenant.RequireEffectiveRole(rbac.RoleSuperAdmin)
}

type createPlanRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (h *adminHandler) createPlan(c *gin.Context) {
	var req createPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	var id uuid.UUID
	err := conn.QueryRow(c.Request.Context(),
		queries.AdminInsertPlan,
		req.Name, req.Description,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

type updatePlanRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (h *adminHandler) updatePlan(c *gin.Context) {
	planID, err := uuid.Parse(c.Param("plan_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan_id"})
		return
	}

	var req updatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	_, err = conn.Exec(c.Request.Context(),
		queries.AdminUpdatePlan,
		req.Name, req.Description, req.Status, planID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "plan updated"})
}

func (h *adminHandler) deletePlan(c *gin.Context) {
	planID, err := uuid.Parse(c.Param("plan_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan_id"})
		return
	}

	conn := tenant.ConnFromContext(c)
	_, err = conn.Exec(c.Request.Context(),
		queries.AdminDeletePlan,
		planID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "plan deleted"})
}

type createPricingRequest struct {
	DurationMonths int     `json:"duration_months" binding:"required,min=1"`
	Price          float64 `json:"price" binding:"required,min=0"`
	Currency       string  `json:"currency"`
}

func (h *adminHandler) createPricing(c *gin.Context) {
	planID, err := uuid.Parse(c.Param("plan_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan_id"})
		return
	}

	var req createPricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	conn := tenant.ConnFromContext(c)
	var id uuid.UUID
	err = conn.QueryRow(c.Request.Context(),
		queries.AdminInsertPricing,
		planID, req.DurationMonths, req.Price, currency,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

type updatePricingRequest struct {
	DurationMonths int     `json:"duration_months" binding:"min=1"`
	Price          float64 `json:"price" binding:"min=0"`
	Currency       string  `json:"currency"`
	Status         string  `json:"status"`
}

func (h *adminHandler) updatePricing(c *gin.Context) {
	pricingID, err := uuid.Parse(c.Param("pricing_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pricing_id"})
		return
	}

	var req updatePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	_, err = conn.Exec(c.Request.Context(),
		queries.AdminUpdatePricing,
		req.DurationMonths, req.Price, req.Currency, req.Status, pricingID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pricing updated"})
}

func (h *adminHandler) deletePricing(c *gin.Context) {
	pricingID, err := uuid.Parse(c.Param("pricing_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pricing_id"})
		return
	}

	conn := tenant.ConnFromContext(c)
	_, err = conn.Exec(c.Request.Context(),
		queries.AdminDeletePricing,
		pricingID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pricing deleted"})
}

func (h *adminHandler) listSystemProjects(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	ctx := c.Request.Context()

	tx, err := conn.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	projects, err := projectsync.NewSyncService().ListSystemProjects(ctx, tx)
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

func (h *adminHandler) createSystemProject(c *gin.Context) {
	var req projectUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	normalizeProjectDefaults(&req)

	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)
	ctx := c.Request.Context()

	tx, err := conn.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	id, err := projectsync.NewSyncService().CreateProject(ctx, tx, projectInputFromReq(req, systemTenantID, "system", userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *adminHandler) updateSystemProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	var req projectUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	normalizeProjectDefaults(&req)

	conn := tenant.ConnFromContext(c)
	ctx := c.Request.Context()

	tx, err := conn.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := projectsync.NewSyncService().UpdateProject(ctx, tx, projectID, projectInputFromReq(req, uuid.Nil, "", uuid.Nil)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project updated"})
}

func (h *adminHandler) deleteSystemProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	conn := tenant.ConnFromContext(c)
	if err := projectsync.NewSyncService().DeleteProject(c.Request.Context(), conn, projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
}

// listUsers 返回所有用户及其订阅计划列表（SuperAdmin 专用）。
func (h *adminHandler) listUsers(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	ctx := c.Request.Context()

	rows, err := conn.Query(ctx,
		queries.AdminListUsers,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type subscriptionSummary struct {
		ID             *uuid.UUID `json:"id,omitempty"`
		Name           *string    `json:"name,omitempty"`
		PlanName       *string    `json:"plan_name,omitempty"`
		DurationMonths *int       `json:"duration_months,omitempty"`
		Price          *float64   `json:"price,omitempty"`
		Currency       *string    `json:"currency,omitempty"`
		SubscribedAt   *time.Time `json:"subscribed_at,omitempty"`
		ExpiresAt      *time.Time `json:"expires_at,omitempty"`
		Status         *string    `json:"status,omitempty"`
	}

	type userResp struct {
		ID            uuid.UUID             `json:"id"`
		Email         string                `json:"email"`
		Name          *string               `json:"name,omitempty"`
		EmailVerified bool                  `json:"email_verified"`
		IsSuperAdmin  bool                  `json:"is_superadmin"`
		CreatedAt     time.Time             `json:"created_at"`
		Subscriptions []subscriptionSummary `json:"subscriptions"`
	}

	users := make(map[uuid.UUID]*userResp)
	for rows.Next() {
		var userID uuid.UUID
		var email string
		var name *string
		var emailVerified bool
		var isSuperAdmin bool
		var createdAt time.Time

		var subID *uuid.UUID
		var subName *string
		var planName *string
		var durationMonths *int
		var price *float64
		var currency *string
		var subscribedAt *time.Time
		var expiresAt *time.Time
		var status *string

		if err := rows.Scan(
			&userID, &email, &name, &emailVerified, &isSuperAdmin, &createdAt,
			&subID, &subName, &planName, &durationMonths, &price, &currency,
			&subscribedAt, &expiresAt, &status,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if _, ok := users[userID]; !ok {
			users[userID] = &userResp{
				ID:            userID,
				Email:         email,
				Name:          name,
				EmailVerified: emailVerified,
				IsSuperAdmin:  isSuperAdmin,
				CreatedAt:     createdAt,
				Subscriptions: []subscriptionSummary{},
			}
		}

		if subID != nil {
			users[userID].Subscriptions = append(users[userID].Subscriptions, subscriptionSummary{
				ID:             subID,
				Name:           subName,
				PlanName:       planName,
				DurationMonths: durationMonths,
				Price:          price,
				Currency:       currency,
				SubscribedAt:   subscribedAt,
				ExpiresAt:      expiresAt,
				Status:         status,
			})
		}
	}

	result := make([]userResp, 0, len(users))
	for _, u := range users {
		result = append(result, *u)
	}

	c.JSON(http.StatusOK, gin.H{"users": result})
}

// listAdminPlans 返回所有计划（含 inactive）及完整定价，供 SuperAdmin 管理界面使用。
func (h *adminHandler) listAdminPlans(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	ctx := c.Request.Context()

	rows, err := conn.Query(ctx,
		queries.AdminListPlans,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type pricingResp struct {
		ID             uuid.UUID `json:"id"`
		DurationMonths int       `json:"duration_months"`
		Price          float64   `json:"price"`
		Currency       string    `json:"currency"`
		Status         string    `json:"status"`
	}

	type planResp struct {
		ID          uuid.UUID     `json:"id"`
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Status      string        `json:"status"`
		Pricing     []pricingResp `json:"pricing"`
	}

	plans := make(map[uuid.UUID]*planResp)
	for rows.Next() {
		var planID uuid.UUID
		var name, description, status string
		var pricingID *uuid.UUID
		var durationMonths *int
		var price *float64
		var currency, pricingStatus *string

		if err := rows.Scan(
			&planID, &name, &description, &status,
			&pricingID, &durationMonths, &price, &currency, &pricingStatus,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if _, ok := plans[planID]; !ok {
			plans[planID] = &planResp{
				ID:          planID,
				Name:        name,
				Description: description,
				Status:      status,
				Pricing:     []pricingResp{},
			}
		}

		if pricingID != nil {
			plans[planID].Pricing = append(plans[planID].Pricing, pricingResp{
				ID:             *pricingID,
				DurationMonths: *durationMonths,
				Price:          *price,
				Currency:       *currency,
				Status:         *pricingStatus,
			})
		}
	}

	result := make([]planResp, 0, len(plans))
	for _, p := range plans {
		result = append(result, *p)
	}

	c.JSON(http.StatusOK, gin.H{"plans": result})
}
