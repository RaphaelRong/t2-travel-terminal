-- 000004_subscriptions.up.sql
-- 将 Tenant 重新定义为「订阅计划实例」，并增加 SuperAdmin 与计划管理相关表。

-- 订阅计划定义表（由 SuperAdmin 管理）
CREATE TABLE IF NOT EXISTS plans (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    description text,
    status text DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- 订阅计划定价表：一个计划可对应多种时长/价格
CREATE TABLE IF NOT EXISTS plan_pricing (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id uuid NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    duration_months int NOT NULL CHECK (duration_months > 0),
    price decimal(10,2) NOT NULL CHECK (price >= 0),
    currency text DEFAULT 'USD',
    status text DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    UNIQUE (plan_id, duration_months)
);

-- 用户表增加系统级 SuperAdmin 标记
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_superadmin boolean NOT NULL DEFAULT false;

-- tenants 视为「订阅计划实例」，补充订阅关系字段
ALTER TABLE tenants
    ADD COLUMN IF NOT EXISTS pricing_id uuid REFERENCES plan_pricing(id),
    ADD COLUMN IF NOT EXISTS subscribed_at timestamptz,
    ADD COLUMN IF NOT EXISTS expires_at timestamptz,
    ADD COLUMN IF NOT EXISTS auto_renew boolean DEFAULT false;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'tenants'
          AND column_name = 'plan_id'
    ) THEN
        ALTER TABLE tenants ADD COLUMN plan_id uuid;
    ELSIF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'tenants'
          AND column_name = 'plan_id'
          AND udt_name <> 'uuid'
    ) THEN
        ALTER TABLE tenants ALTER COLUMN plan_id DROP DEFAULT;
        ALTER TABLE tenants
            ALTER COLUMN plan_id TYPE uuid
            USING (
                CASE
                    WHEN plan_id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
                    THEN plan_id::uuid
                    ELSE NULL
                END
            );
    END IF;

    UPDATE tenants
    SET plan_id = NULL
    WHERE plan_id IS NOT NULL
      AND NOT EXISTS (
          SELECT 1
          FROM plans
          WHERE plans.id = tenants.plan_id
      );

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'tenants_plan_id_fkey'
          AND conrelid = 'public.tenants'::regclass
    ) THEN
        ALTER TABLE tenants
            ADD CONSTRAINT tenants_plan_id_fkey
            FOREIGN KEY (plan_id) REFERENCES plans(id);
    END IF;
END $$;

-- 插入默认订阅计划（默认英文内容）
INSERT INTO plans (id, name, description, status, role_key) VALUES
    (gen_random_uuid(), 'Free Trial', 'For individuals or small teams to try T2 core features', 'active', 'free_user'),
    (gen_random_uuid(), 'Basic', 'For growing teams with more projects and reports', 'active', 'paid_user'),
    (gen_random_uuid(), 'Advanced', 'For large enterprises with advanced data and dedicated support', 'active', 'premium_paid_user')
ON CONFLICT DO NOTHING;

-- 为默认计划插入定价
WITH free_plan AS (SELECT id FROM plans WHERE name = 'Free Trial' LIMIT 1),
     basic_plan AS (SELECT id FROM plans WHERE name = 'Basic' LIMIT 1),
     advanced_plan AS (SELECT id FROM plans WHERE name = 'Advanced' LIMIT 1)
INSERT INTO plan_pricing (plan_id, duration_months, price, currency)
SELECT id, 1, 0, 'USD' FROM free_plan
UNION ALL
SELECT id, 1, 29, 'USD' FROM basic_plan
UNION ALL
SELECT id, 12, 290, 'USD' FROM basic_plan
UNION ALL
SELECT id, 1, 99, 'USD' FROM advanced_plan
UNION ALL
SELECT id, 12, 990, 'USD' FROM advanced_plan
ON CONFLICT DO NOTHING;

-- 创建系统管理员账号，邮箱已验证，无需再点击验证链接
INSERT INTO users (id, email, name, password_hash, email_verified, is_superadmin)
VALUES (
    gen_random_uuid(),
    'Admin@super.com',
    'SuperAdmin',
    '$2a$10$keGW9/L/Y9ENaZ5tOBrxguVBboCs5nhRRr4vJaCDkedRYK1EwGQG2',
    true,
    true
)
ON CONFLICT (email) DO NOTHING;
