package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

func listTenantsHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	rows, err := conn.Query(c.Request.Context(),
		`SELECT t.id, t.name, t.slug, t.plan_id, m.role
		 FROM tenants t
		 JOIN memberships m ON t.id = m.tenant_id
		 WHERE m.user_id = $1
		 ORDER BY t.name`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type tenantResp struct {
		ID     uuid.UUID `json:"id"`
		Name   string    `json:"name"`
		Slug   *string   `json:"slug,omitempty"`
		PlanID string    `json:"plan_id"`
		Role   string    `json:"role"`
	}
	var result []tenantResp
	for rows.Next() {
		var t tenantResp
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.PlanID, &t.Role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result = append(result, t)
	}

	c.JSON(http.StatusOK, gin.H{"tenants": result})
}

func listProjectsHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	rows, err := conn.Query(c.Request.Context(),
		`SELECT id, name, description, created_by, created_at
		 FROM projects
		 WHERE tenant_id = $1
		 ORDER BY created_at DESC`,
		t.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type projectResp struct {
		ID          uuid.UUID  `json:"id"`
		Name        string     `json:"name"`
		Description *string    `json:"description,omitempty"`
		CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
		CreatedAt   time.Time  `json:"created_at"`
	}
	var result []projectResp
	for rows.Next() {
		var p projectResp
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result = append(result, p)
	}

	c.JSON(http.StatusOK, gin.H{"projects": result})
}

func createProjectHandler(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)
	userID := tenant.UserIDFromContext(c)

	description := req.Description

	var id uuid.UUID
	err := conn.QueryRow(c.Request.Context(),
		`INSERT INTO projects (tenant_id, name, description, created_by)
		 VALUES ($1, $2, NULLIF($3, ''), $4)
		 RETURNING id`,
		t.ID, req.Name, description, userID,
	).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}
