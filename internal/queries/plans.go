package queries

const (
	// PlansListActive 返回所有活跃计划及其活跃定价。
	PlansListActive = `SELECT p.id, p.name, p.description,
	        pp.id, pp.duration_months, pp.price, pp.currency
	 FROM plans p
	 LEFT JOIN plan_pricing pp ON pp.plan_id = p.id AND pp.status = 'active'
	 WHERE p.status = 'active'
	 ORDER BY p.name, pp.duration_months`
)
