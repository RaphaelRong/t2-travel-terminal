package queries

const (
	// CommonSetCurrentUserID 设置 RLS 会话变量：当前用户 ID。
	CommonSetCurrentUserID = `SELECT set_config('app.current_user_id', $1, false)`

	// CommonSetCurrentTenantID 设置 RLS 会话变量：当前租户 ID。
	CommonSetCurrentTenantID = `SELECT set_config('app.current_tenant_id', $1, false)`

	// CommonResetCurrentUserID 重置 RLS 会话变量：当前用户 ID。
	CommonResetCurrentUserID = `SELECT set_config('app.current_user_id', '', false)`

	// CommonResetCurrentTenantID 重置 RLS 会话变量：当前租户 ID。
	CommonResetCurrentTenantID = `SELECT set_config('app.current_tenant_id', '', false)`

	// CommonSelectMembershipRole 查询用户在指定租户中的角色。
	CommonSelectMembershipRole = `SELECT role FROM memberships WHERE tenant_id = $1 AND user_id = $2`

	// CommonSelectMembershipAndPlan 查询用户在指定租户中的成员角色及该租户订阅的计划 role_key。
	CommonSelectMembershipAndPlan = `SELECT m.role, COALESCE(p.role_key, '')
	 FROM memberships m
	 LEFT JOIN tenants t ON t.id = m.tenant_id
	 LEFT JOIN plans p ON p.id = t.plan_id
	 WHERE m.tenant_id = $1 AND m.user_id = $2`

	// CommonCreateSchemaMigrationsTable 创建迁移记录表。
	CommonCreateSchemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

	// CommonSelectAppliedVersions 查询已应用的迁移版本。
	CommonSelectAppliedVersions = `SELECT version FROM schema_migrations`

	// CommonSelectAppliedVersionsDesc 按版本降序查询已应用的迁移。
	CommonSelectAppliedVersionsDesc = `SELECT version FROM schema_migrations ORDER BY version DESC`

	// CommonInsertSchemaMigration 记录已应用的迁移版本。
	CommonInsertSchemaMigration = `INSERT INTO schema_migrations (version) VALUES ($1)`

	// CommonDeleteSchemaMigration 删除迁移记录。
	CommonDeleteSchemaMigration = `DELETE FROM schema_migrations WHERE version = $1`

	// CommonVerifyUserEmail 直接验证指定邮箱的用户（管理脚本）。
	CommonVerifyUserEmail = `UPDATE users SET email_verified = true, email_verified_at = now() WHERE email = $1`

	// CommonMarkEmailVerificationUsedByEmail 根据邮箱标记验证 token 已使用（管理脚本）。
	CommonMarkEmailVerificationUsedByEmail = `UPDATE email_verifications SET used_at = now() WHERE user_id = (SELECT id FROM users WHERE email = $1)`
)
