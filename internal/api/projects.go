package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

func getProjectHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	var result struct {
		ID          uuid.UUID  `json:"id"`
		Name        string     `json:"name"`
		Description *string    `json:"description,omitempty"`
		CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
	}

	err = conn.QueryRow(c.Request.Context(),
		`SELECT id, name, description, created_by, created_at, updated_at
		 FROM projects
		 WHERE tenant_id = $1 AND id = $2`,
		t.ID, projectID,
	).Scan(&result.ID, &result.Name, &result.Description, &result.CreatedBy, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func updateProjectHandler(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	_, err = conn.Exec(c.Request.Context(),
		`UPDATE projects
		 SET name = COALESCE(NULLIF($1, ''), name),
		     description = COALESCE(NULLIF($2, ''), description),
		     updated_at = now()
		 WHERE tenant_id = $3 AND id = $4`,
		req.Name, req.Description, t.ID, projectID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project updated"})
}

func deleteProjectHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	_, err = conn.Exec(c.Request.Context(),
		`DELETE FROM projects WHERE tenant_id = $1 AND id = $2`,
		t.ID, projectID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
}
