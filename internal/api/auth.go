package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/auth"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/datastore"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/notification"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/queries"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/rbac"
	"go.uber.org/zap"
)

type authHandler struct {
	pool     *datastore.Pool
	tm       *auth.TokenManager
	emailSvc *notification.Service
	logger   *zap.Logger
}

func newAuthHandler(pool *datastore.Pool, tm *auth.TokenManager, emailSvc *notification.Service, logger *zap.Logger) *authHandler {
	return &authHandler{
		pool:     pool,
		tm:       tm,
		emailSvc: emailSvc,
		logger:   logger,
	}
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name"`
}

func (h *authHandler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := auth.GenerateVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID uuid.UUID
	var emailVerified bool

	err = tx.QueryRow(ctx,
		queries.AuthSelectUserByEmailForUpdate,
		req.Email,
	).Scan(&userID, &emailVerified)

	switch {
	case err == nil:
		if emailVerified {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		// 邮箱存在但未验证：允许重新注册，更新密码和姓名后重新发邮件
		_, err = tx.Exec(ctx,
			queries.AuthUpdateUserPasswordAndName,
			passwordHash, req.Name, userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case errors.Is(err, pgx.ErrNoRows):
		// 新用户
		err = tx.QueryRow(ctx,
			queries.AuthInsertUser,
			req.Email, req.Name, passwordHash,
		).Scan(&userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 清除该用户旧的验证 token，插入新的
	_, err = tx.Exec(ctx, queries.AuthDeleteEmailVerificationsByUser, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(ctx,
		queries.AuthInsertEmailVerification,
		userID, token,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 同步发送验证邮件；发送失败视为注册失败的一部分，前端可明确得知。
	if h.emailSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "email service not configured"})
		return
	}
	if err := h.emailSvc.SendVerificationEmail(req.Email, token); err != nil {
		h.logger.Error("failed to send verification email",
			zap.String("to", req.Email),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send verification email"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id":    userID,
		"email_sent": true,
		"message":    "registration successful, please verify your email",
	})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *authHandler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	var user struct {
		ID            uuid.UUID
		PasswordHash  string
		EmailVerified bool
		IsSuperAdmin  bool
	}
	err := h.pool.QueryRow(ctx,
		queries.AuthSelectUserForLogin,
		req.Email,
	).Scan(&user.ID, &user.PasswordHash, &user.EmailVerified, &user.IsSuperAdmin)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if !user.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "email not verified"})
		return
	}

	accessToken, err := h.tm.IssueAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	role := rbac.RoleFreeUser
	if user.IsSuperAdmin {
		role = rbac.RoleSuperAdmin
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    86400,
		"user_id":       user.ID,
		"is_superadmin": user.IsSuperAdmin,
		"role":          role,
	})
}

func (h *authHandler) verifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
		return
	}

	ctx := c.Request.Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID uuid.UUID
	err = tx.QueryRow(ctx,
		queries.AuthSelectEmailVerification,
		token,
	).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	_, err = tx.Exec(ctx,
		queries.AuthMarkEmailVerificationUsed,
		token,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(ctx,
		queries.AuthVerifyUserEmail,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
}

type resendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// resendVerification 为未验证用户重新发送验证邮件。
func (h *authHandler) resendVerification(c *gin.Context) {
	var req resendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.emailSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "email service not configured"})
		return
	}

	token, err := auth.GenerateVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID uuid.UUID
	var emailVerified bool
	err = tx.QueryRow(ctx,
		queries.AuthSelectUserByEmailForUpdate,
		req.Email,
	).Scan(&userID, &emailVerified)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or already verified email"})
		return
	}
	if emailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or already verified email"})
		return
	}

	_, err = tx.Exec(ctx, queries.AuthDeleteEmailVerificationsByUser, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, err = tx.Exec(ctx,
		queries.AuthInsertEmailVerification,
		userID, token,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.emailSvc.SendVerificationEmail(req.Email, token); err != nil {
		h.logger.Error("failed to resend verification email",
			zap.String("to", req.Email),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send verification email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email_sent": true,
		"message":    "verification email sent",
	})
}
