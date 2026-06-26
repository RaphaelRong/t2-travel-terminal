package queries

const (
	// AuthSelectUserByEmailForUpdate 根据邮箱查询用户并加锁（注册/重发验证邮件）。
	AuthSelectUserByEmailForUpdate = `SELECT id, email_verified FROM users WHERE email = $1 FOR UPDATE`

	// AuthUpdateUserPasswordAndName 更新已存在但未验证用户的密码和姓名。
	AuthUpdateUserPasswordAndName = `UPDATE users
	 SET password_hash = $1,
	     name = NULLIF($2, ''),
	     updated_at = now()
	 WHERE id = $3`

	// AuthInsertUser 创建新用户。
	AuthInsertUser = `INSERT INTO users (email, name, password_hash, email_verified)
	 VALUES ($1, NULLIF($2, ''), $3, false)
	 RETURNING id`

	// AuthDeleteEmailVerificationsByUser 清除某用户的全部邮件验证 token。
	AuthDeleteEmailVerificationsByUser = `DELETE FROM email_verifications WHERE user_id = $1`

	// AuthInsertEmailVerification 插入新的邮件验证 token。
	AuthInsertEmailVerification = `INSERT INTO email_verifications (user_id, token, expires_at)
	 VALUES ($1, $2, now() + interval '24 hours')`

	// AuthSelectUserForLogin 登录时查询用户凭据与验证状态。
	AuthSelectUserForLogin = `SELECT id, password_hash, email_verified, is_superadmin FROM users WHERE email = $1`

	// AuthSelectEmailVerification 查询未使用且未过期的验证 token。
	AuthSelectEmailVerification = `SELECT user_id FROM email_verifications
	 WHERE token = $1
	   AND used_at IS NULL
	   AND expires_at > now()
	 FOR UPDATE`

	// AuthMarkEmailVerificationUsed 标记验证 token 已使用。
	AuthMarkEmailVerificationUsed = `UPDATE email_verifications SET used_at = now() WHERE token = $1`

	// AuthVerifyUserEmail 将用户标记为已验证。
	AuthVerifyUserEmail = `UPDATE users SET email_verified = true, email_verified_at = now() WHERE id = $1`
)
