package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

type createTenantRequest struct {
	Name   string `json:"name" binding:"required"`
	Slug   string `json:"slug"`
	PlanID string `json:"plan_id"`
}

func createTenantHandler(c *gin.Context) {
	var req createTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	planID := req.PlanID
	if planID == "" {
		planID = "free"
	}

	tx, err := conn.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	var tenantID uuid.UUID
	err = tx.QueryRow(c.Request.Context(),
		`INSERT INTO tenants (name, slug, plan_id, created_by)
		 VALUES ($1, NULLIF($2, ''), $3, $4)
		 RETURNING id`,
		req.Name, req.Slug, planID, userID,
	).Scan(&tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(c.Request.Context(),
		`INSERT INTO memberships (tenant_id, user_id, role)
		 VALUES ($1, $2, 'owner')`,
		tenantID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      tenantID,
		"name":    req.Name,
		"plan_id": planID,
		"role":    "owner",
	})
}

func getTenantHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	var result struct {
		ID        uuid.UUID  `json:"id"`
		Name      string     `json:"name"`
		Slug      *string    `json:"slug,omitempty"`
		PlanID    string     `json:"plan_id"`
		Status    string     `json:"status"`
		CreatedBy *uuid.UUID `json:"created_by,omitempty"`
		CreatedAt time.Time  `json:"created_at"`
	}

	err := conn.QueryRow(c.Request.Context(),
		`SELECT id, name, slug, plan_id, status, created_by, created_at
		 FROM tenants WHERE id = $1`,
		t.ID,
	).Scan(&result.ID, &result.Name, &result.Slug, &result.PlanID, &result.Status, &result.CreatedBy, &result.CreatedAt)
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
		`UPDATE tenants
		 SET name = COALESCE(NULLIF($1, ''), name),
		     slug = COALESCE(NULLIF($2, ''), slug),
		     updated_at = now()
		 WHERE id = $3`,
		req.Name, req.Slug, t.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tenant updated"})
}

func deleteTenantHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	_, err := conn.Exec(c.Request.Context(),
		`DELETE FROM tenants WHERE id = $1`,
		t.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tenant deleted"})
}

func updateTenantPlanHandler(c *gin.Context) {
	var req struct {
		PlanID string `json:"plan_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	_, err := conn.Exec(c.Request.Context(),
		`UPDATE tenants SET plan_id = $1, updated_at = now() WHERE id = $2`,
		req.PlanID, t.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan_id": req.PlanID})
}
