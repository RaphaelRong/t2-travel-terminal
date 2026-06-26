-- 000006_fix_tenant_plan_id_uuid.down.sql
-- Keep this migration irreversible to avoid reintroducing the text = uuid mismatch.

DO $$
BEGIN
    NULL;
END $$;
