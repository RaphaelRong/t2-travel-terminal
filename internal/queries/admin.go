package queries

const (
	// AdminSelectSuperAdmin 查询用户是否为超级管理员。
	AdminSelectSuperAdmin = `SELECT is_superadmin FROM users WHERE id = $1`

	// AdminInsertPlan 创建计划。
	AdminInsertPlan = `INSERT INTO plans (name, description)
	 VALUES ($1, NULLIF($2, ''))
	 RETURNING id`

	// AdminUpdatePlan 更新计划。
	AdminUpdatePlan = `UPDATE plans
	 SET name = COALESCE(NULLIF($1, ''), name),
	     description = COALESCE(NULLIF($2, ''), description),
	     status = COALESCE(NULLIF($3, ''), status),
	     updated_at = now()
	 WHERE id = $4`

	// AdminDeletePlan 删除计划。
	AdminDeletePlan = `DELETE FROM plans WHERE id = $1`

	// AdminInsertPricing 创建定价。
	AdminInsertPricing = `INSERT INTO plan_pricing (plan_id, duration_months, price, currency)
	 VALUES ($1, $2, $3, $4)
	 RETURNING id`

	// AdminUpdatePricing 更新定价。
	AdminUpdatePricing = `UPDATE plan_pricing
	 SET duration_months = COALESCE(NULLIF($1, 0), duration_months),
	     price = COALESCE(NULLIF($2, 0), price),
	     currency = COALESCE(NULLIF($3, ''), currency),
	     status = COALESCE(NULLIF($4, ''), status),
	     updated_at = now()
	 WHERE id = $5`

	// AdminDeletePricing 删除定价。
	AdminDeletePricing = `DELETE FROM plan_pricing WHERE id = $1`

	// AdminListUsers 查询所有用户及其订阅（SuperAdmin）。
	AdminListUsers = `SELECT u.id, u.email, u.name, u.email_verified, u.is_superadmin, u.created_at,
	        t.id, t.name, p.name, pp.duration_months, pp.price, pp.currency,
	        t.subscribed_at, t.expires_at, t.status
	 FROM users u
	 LEFT JOIN memberships m ON m.user_id = u.id
	 LEFT JOIN tenants t ON t.id = m.tenant_id
	 LEFT JOIN plans p ON p.id = t.plan_id
	 LEFT JOIN plan_pricing pp ON pp.id = t.pricing_id
	 ORDER BY u.created_at DESC, t.subscribed_at DESC`

	// AdminListPlans 查询所有计划及完整定价（SuperAdmin）。
	AdminListPlans = `SELECT p.id, p.name, p.description, p.status,
	        pp.id, pp.duration_months, pp.price, pp.currency, pp.status
	 FROM plans p
	 LEFT JOIN plan_pricing pp ON pp.plan_id = p.id
	 ORDER BY p.name, pp.duration_months`
)
