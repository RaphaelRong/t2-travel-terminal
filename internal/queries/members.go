package queries

const (
	// MembersListByTenant 查询租户成员列表。
	MembersListByTenant = `SELECT u.id, u.email, u.name, m.role, m.joined_at
	 FROM memberships m
	 JOIN users u ON m.user_id = u.id
	 WHERE m.tenant_id = $1
	 ORDER BY m.joined_at`

	// MembersInsertInvite 创建租户邀请。
	MembersInsertInvite = `INSERT INTO tenant_invites (tenant_id, email, role, token, expires_at)
	 VALUES ($1, $2, $3, $4, now() + interval '7 days')`

	// MembersSelectInvite 查询未使用且未过期的邀请 token。
	MembersSelectInvite = `SELECT tenant_id, email, role FROM tenant_invites
	 WHERE token = $1
	   AND used_at IS NULL
	   AND expires_at > now()
	 FOR UPDATE`

	// MembersSelectUserEmail 查询用户邮箱。
	MembersSelectUserEmail = `SELECT email FROM users WHERE id = $1`

	// MembersUpsertMembership 接受邀请时插入或忽略已存在的成员关系。
	MembersUpsertMembership = `INSERT INTO memberships (tenant_id, user_id, role)
	 VALUES ($1, $2, $3)
	 ON CONFLICT (tenant_id, user_id) DO NOTHING`

	// MembersMarkInviteUsed 标记邀请 token 已使用。
	MembersMarkInviteUsed = `UPDATE tenant_invites SET used_at = now() WHERE token = $1`

	// MembersRemove 移除租户成员（不能移除 owner）。
	MembersRemove = `DELETE FROM memberships
	 WHERE tenant_id = $1 AND user_id = $2 AND role != 'owner'`

	// MembersUpdateRole 更新成员角色（不能修改 owner）。
	MembersUpdateRole = `UPDATE memberships
	 SET role = $1
	 WHERE tenant_id = $2 AND user_id = $3 AND role != 'owner'`
)
