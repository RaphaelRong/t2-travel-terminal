package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/auth"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/queries"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

func listMembersHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	rows, err := conn.Query(c.Request.Context(),
		queries.MembersListByTenant,
		t.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type memberResp struct {
		UserID   uuid.UUID `json:"user_id"`
		Email    string    `json:"email"`
		Name     *string   `json:"name,omitempty"`
		Role     string    `json:"role"`
		JoinedAt time.Time `json:"joined_at"`
	}

	var result []memberResp
	for rows.Next() {
		var m memberResp
		if err := rows.Scan(&m.UserID, &m.Email, &m.Name, &m.Role, &m.JoinedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result = append(result, m)
	}

	c.JSON(http.StatusOK, gin.H{"members": result})
}

type inviteRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin member"`
}

func inviteMemberHandler(c *gin.Context) {
	var req inviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	token, err := auth.GenerateVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = conn.Exec(c.Request.Context(),
		queries.MembersInsertInvite,
		t.ID, req.Email, req.Role, token,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":      token,
		"invite_url": "/api/v1/invites/accept?token=" + token,
	})
}

func acceptInviteHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
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

	var invite struct {
		TenantID uuid.UUID
		Email    string
		Role     string
	}
	err = tx.QueryRow(ctx,
		queries.MembersSelectInvite,
		token,
	).Scan(&invite.TenantID, &invite.Email, &invite.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired invite"})
		return
	}

	var userEmail string
	err = tx.QueryRow(ctx,
		queries.MembersSelectUserEmail,
		userID,
	).Scan(&userEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if userEmail != invite.Email {
		c.JSON(http.StatusForbidden, gin.H{"error": "invite email does not match your account email"})
		return
	}

	_, err = tx.Exec(ctx,
		queries.MembersUpsertMembership,
		invite.TenantID, userID, invite.Role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(ctx,
		queries.MembersMarkInviteUsed,
		token,
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
		"tenant_id": invite.TenantID,
		"role":      invite.Role,
	})
}

func removeMemberHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	targetUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// 不能移除 owner（更严格的逻辑可以在应用层做）
	_, err = conn.Exec(c.Request.Context(),
		queries.MembersRemove,
		t.ID, targetUserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member removed"})
}

type updateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

func updateMemberRoleHandler(c *gin.Context) {
	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	targetUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	_, err = conn.Exec(c.Request.Context(),
		queries.MembersUpdateRole,
		req.Role, t.ID, targetUserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"role": req.Role})
}
