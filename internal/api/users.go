package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

func getMeHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	var user struct {
		ID            string  `json:"id"`
		Email         string  `json:"email"`
		Name          *string `json:"name,omitempty"`
		EmailVerified bool    `json:"email_verified"`
	}

	err := conn.QueryRow(c.Request.Context(),
		`SELECT id, email, name, email_verified FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.EmailVerified)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func updateMeHandler(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	_, err := conn.Exec(c.Request.Context(),
		`UPDATE users SET name = NULLIF($1, ''), updated_at = now() WHERE id = $2`,
		req.Name, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

func deleteMeHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)
	ctx := c.Request.Context()

	tx, err := conn.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 解除用户创建的外部引用，避免外键约束阻止删除
	if _, err := tx.Exec(ctx, `UPDATE projects SET created_by = NULL WHERE created_by = $1`, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := tx.Exec(ctx, `UPDATE tenants SET created_by = NULL WHERE created_by = $1`, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// memberships 会通过 ON DELETE CASCADE 自动清理
	if _, err := tx.Exec(ctx, `DELETE FROM email_verifications WHERE user_id = $1`, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
