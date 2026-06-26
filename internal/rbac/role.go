// Package rbac 提供 T2 统一的角色与权限模型。
//
// 系统角色（system-wide）由用户身份或当前订阅计划决定：
//   - superadmin      : 系统管理员，拥有全部权限
//   - free_user       : 当前租户订阅 Free Trial
//   - paid_user       : 当前租户订阅 Basic
//   - premium_paid_user: 当前租户订阅 Advanced
//
// 租户角色（tenant-scoped）保留在 memberships.role 中，用于租户内管理：
//   - owner / admin / member
package rbac

// Role 表示用户的系统级角色。
type Role string

const (
	RoleSuperAdmin  Role = "superadmin"
	RoleFreeUser    Role = "free_user"
	RolePaidUser    Role = "paid_user"
	RolePremiumUser Role = "premium_paid_user"

	// TenantOwner 租户所有者
	TenantOwner Role = "owner"
	// TenantAdmin 租户管理员
	TenantAdmin Role = "admin"
	// TenantMember 普通成员
	TenantMember Role = "member"
)

// SubscriptionRank 定义订阅角色的等级，数值越大权限越高。
var SubscriptionRank = map[Role]int{
	RoleFreeUser:    1,
	RolePaidUser:    2,
	RolePremiumUser: 3,
	RoleSuperAdmin:  99,
}

// IsSuperAdmin 判断角色是否为系统管理员。
func IsSuperAdmin(role Role) bool {
	return role == RoleSuperAdmin
}

// IsSubscriptionRole 判断是否为订阅类角色。
func IsSubscriptionRole(role Role) bool {
	rank, ok := SubscriptionRank[role]
	return ok && rank > 0 && rank < 99
}

// IsAtLeastSubscription 判断角色订阅级别是否不低于 required。
// SuperAdmin 永远返回 true。
func IsAtLeastSubscription(role, required Role) bool {
	if role == RoleSuperAdmin {
		return true
	}
	r, ok1 := SubscriptionRank[role]
	req, ok2 := SubscriptionRank[required]
	if !ok1 || !ok2 {
		return false
	}
	return r >= req
}

// RoleKeyToRole 将 plans.role_key 映射为系统角色。
func RoleKeyToRole(roleKey string) Role {
	switch roleKey {
	case "paid_user":
		return RolePaidUser
	case "premium_paid_user":
		return RolePremiumUser
	case "free_user":
		return RoleFreeUser
	default:
		// 未知 role_key 按最受限处理
		return RoleFreeUser
	}
}

// PlanNameToRole 将订阅计划名称映射为系统角色（兼容旧数据/未设置 role_key 的场景）。
func PlanNameToRole(planName string) Role {
	switch planName {
	case "Basic":
		return RolePaidUser
	case "Advanced":
		return RolePremiumUser
	case "Free Trial":
		return RoleFreeUser
	default:
		return RoleFreeUser
	}
}
