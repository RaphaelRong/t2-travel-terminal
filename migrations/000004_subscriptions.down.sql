-- 000004_subscriptions.down.sql
-- 回退订阅计划相关改动

ALTER TABLE tenants
    DROP COLUMN IF EXISTS plan_id,
    DROP COLUMN IF EXISTS pricing_id,
    DROP COLUMN IF EXISTS subscribed_at,
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS auto_renew;

ALTER TABLE users
    DROP COLUMN IF EXISTS is_superadmin;

DROP TABLE IF EXISTS plan_pricing;
DROP TABLE IF EXISTS plans;
