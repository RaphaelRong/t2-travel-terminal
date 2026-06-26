package queries

const (
	// TenantListSubscriptions 查询当前用户的全部租户订阅。
	TenantListSubscriptions = `SELECT t.id, t.name, t.slug, t.status,
	        t.subscribed_at, t.expires_at, t.auto_renew,
	        p.id, p.name, p.role_key,
	        pp.id, pp.duration_months, pp.price, pp.currency,
	        m.role
	 FROM tenants t
	 JOIN memberships m ON t.id = m.tenant_id
	 LEFT JOIN plans p ON t.plan_id = p.id
	 LEFT JOIN plan_pricing pp ON t.pricing_id = pp.id
	 WHERE m.user_id = $1::uuid
	 ORDER BY t.subscribed_at DESC NULLS LAST`

	// TenantValidatePricingForPlan 校验定价是否属于指定计划且处于活跃状态。
	TenantValidatePricingForPlan = `SELECT p.name, pp.duration_months
	 FROM plan_pricing pp
	 JOIN plans p ON p.id = pp.plan_id
	 WHERE pp.id = $1 AND pp.plan_id = $2 AND pp.status = 'active'`

	// TenantInsertTenant 创建租户。
	TenantInsertTenant = `INSERT INTO tenants (name, slug, plan_id, pricing_id, subscribed_at, expires_at, auto_renew, created_by)
	 VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, $8)
	 RETURNING id`

	// TenantInsertOwnerMembership 为租户创建 owner 成员关系。
	TenantInsertOwnerMembership = `INSERT INTO memberships (tenant_id, user_id, role)
	 VALUES ($1, $2, 'owner')`

	// TenantSelectByID 查询单个租户详情。
	TenantSelectByID = `SELECT t.id, t.name, t.slug, t.status,
	        t.plan_id, p.name, t.pricing_id, pp.duration_months, pp.price,
	        t.subscribed_at, t.expires_at, t.auto_renew,
	        t.created_by, t.created_at
	 FROM tenants t
	 LEFT JOIN plans p ON t.plan_id = p.id
	 LEFT JOIN plan_pricing pp ON t.pricing_id = pp.id
	 WHERE t.id = $1`

	// TenantUpdate 更新租户名称与 slug。
	TenantUpdate = `UPDATE tenants
	 SET name = COALESCE(NULLIF($1, ''), name),
	     slug = COALESCE(NULLIF($2, ''), slug),
	     updated_at = now()
	 WHERE id = $3`

	// TenantDelete 删除租户。
	TenantDelete = `DELETE FROM tenants WHERE id = $1`

	// TenantSelectPricingDuration 查询指定计划的活跃定价时长。
	TenantSelectPricingDuration = `SELECT duration_months FROM plan_pricing WHERE id = $1 AND plan_id = $2 AND status = 'active'`

	// TenantUpdatePlan 更新租户订阅计划与定价。
	TenantUpdatePlan = `UPDATE tenants
	 SET plan_id = $1,
	     pricing_id = $2,
	     subscribed_at = $3,
	     expires_at = $4,
	     auto_renew = $5,
	     updated_at = now()
	 WHERE id = $6`
)
