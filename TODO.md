# T2-TravelTerminal TODO

## 待实现

### 用户密码修改功能

- **背景**：当前 `migrations/000004_subscriptions.up.sql` 中硬编码了默认 SuperAdmin 账号
  `Admin@super.com`，初始密码为 `123456`（bcrypt hash 已公开）。
- **目标**：允许用户（尤其是 SuperAdmin 和租户成员）在登录后自行修改密码。
- **建议实现点**：
  1. 后端新增 `POST /api/v1/me/password` 或 `PUT /api/v1/me/password` 接口。
  2. 要求提供旧密码 + 新密码 + 确认新密码。
  3. 新密码需校验强度（长度、复杂度）。
  4. 用 `bcrypt` 重新计算 `password_hash` 并更新数据库。
  5. 前端在个人设置/Profile 页面增加"修改密码"表单。
  6. （可选）SuperAdmin 首次登录后强制修改密码。
