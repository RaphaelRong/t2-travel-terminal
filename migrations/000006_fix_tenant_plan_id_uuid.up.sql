-- 000006_fix_tenant_plan_id_uuid.up.sql
-- Ensure tenants.plan_id is a uuid so joins to plans.id do not fail with text = uuid.

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
