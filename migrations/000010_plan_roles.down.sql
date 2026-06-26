-- 000010_plan_roles.down.sql
-- 回滚：移除 plans.role_key 字段。

ALTER TABLE plans
    DROP CONSTRAINT IF EXISTS plans_role_key_check;

ALTER TABLE plans
    DROP COLUMN IF EXISTS role_key;
