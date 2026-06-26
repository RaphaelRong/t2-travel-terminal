package queries

const (
	// UsersSelectProfile 查询当前用户资料。
	UsersSelectProfile = `SELECT id, email, name, email_verified FROM users WHERE id = $1`

	// UsersUpdateProfile 更新当前用户姓名。
	UsersUpdateProfile = `UPDATE users SET name = NULLIF($1, ''), updated_at = now() WHERE id = $2`

	// UsersNullifyProjectsCreatedBy 删除用户前将其创建的项目外键置空。
	UsersNullifyProjectsCreatedBy = `UPDATE projects SET created_by = NULL WHERE created_by = $1`

	// UsersNullifyTenantsCreatedBy 删除用户前将其创建的租户外键置空。
	UsersNullifyTenantsCreatedBy = `UPDATE tenants SET created_by = NULL WHERE created_by = $1`

	// UsersDeleteEmailVerifications 删除用户的邮件验证记录。
	UsersDeleteEmailVerifications = `DELETE FROM email_verifications WHERE user_id = $1`

	// UsersDeleteUser 删除用户。
	UsersDeleteUser = `DELETE FROM users WHERE id = $1`
)
