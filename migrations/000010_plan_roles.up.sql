-- 000010_plan_roles.up.sql
-- 为订阅计划增加系统角色标识，使权限控制不再依赖计划名称字符串。

ALTER TABLE plans
    ADD COLUMN IF NOT EXISTS role_key text;

-- 约束：role_key 必须是预定义的系统角色之一
ALTER TABLE plans
    DROP CONSTRAINT IF EXISTS plans_role_key_check;

ALTER TABLE plans
    ADD CONSTRAINT plans_role_key_check
    CHECK (role_key IN ('free_user', 'paid_user', 'premium_paid_user'));

-- 为现有默认计划设置 role_key
UPDATE plans
SET role_key = 'free_user'
WHERE name = 'Free Trial';

UPDATE plans
SET role_key = 'paid_user'
WHERE name = 'Basic';

UPDATE plans
SET role_key = 'premium_paid_user'
WHERE name = 'Advanced';
