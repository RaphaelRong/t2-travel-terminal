-- 000011_project_integrations.down.sql

ALTER TABLE project_capabilities
    DROP COLUMN IF EXISTS external_name,
    DROP COLUMN IF EXISTS integration_id;

DROP TABLE IF EXISTS project_integrations;
