// Package tenant 提供多租户上下文管理：
//   - 从 HTTP 请求解析当前用户和当前租户
//   - 在 gin.Context 中保存数据库连接、用户 ID、租户信息
//   - 提供辅助函数供 handler 读取这些信息
package tenant

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ctxKey string

const (
	ctxUserKey   ctxKey = "tenant.user_id"
	ctxTenantKey ctxKey = "tenant.tenant"
	ctxConnKey   ctxKey = "tenant.conn"
)

// Tenant 表示当前请求所归属的租户及成员角色。
type Tenant struct {
	ID   uuid.UUID
	Role string
}

// SetUserID 在 gin 上下文中保存当前用户 ID。
func SetUserID(c *gin.Context, id uuid.UUID) {
	c.Set(string(ctxUserKey), id)
}

// UserIDFromContext 读取当前用户 ID，若不存在会 panic，请确保 middleware 已正确注入。
func UserIDFromContext(c *gin.Context) uuid.UUID {
	v, exists := c.Get(string(ctxUserKey))
	if !exists {
		panic("tenant: user id not found in context")
	}
	return v.(uuid.UUID)
}

// SetTenant 在 gin 上下文中保存当前租户。
func SetTenant(c *gin.Context, t Tenant) {
	c.Set(string(ctxTenantKey), t)
}

// TenantFromContext 读取当前租户，第二个返回值表示是否存在。
func TenantFromContext(c *gin.Context) (Tenant, bool) {
	v, exists := c.Get(string(ctxTenantKey))
	if !exists {
		return Tenant{}, false
	}
	return v.(Tenant), true
}

// SetConn 在 gin 上下文中保存当前数据库连接。
func SetConn(c *gin.Context, conn *pgxpool.Conn) {
	c.Set(string(ctxConnKey), conn)
}

// ConnFromContext 读取当前数据库连接，若不存在会 panic。
func ConnFromContext(c *gin.Context) *pgxpool.Conn {
	v, exists := c.Get(string(ctxConnKey))
	if !exists {
		panic("tenant: db connection not found in context")
	}
	return v.(*pgxpool.Conn)
}

// HasTenant 判断当前请求是否已经绑定租户。
func HasTenant(c *gin.Context) bool {
	_, exists := c.Get(string(ctxTenantKey))
	return exists
}
