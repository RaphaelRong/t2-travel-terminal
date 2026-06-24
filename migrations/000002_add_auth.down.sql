-- 000002_add_auth.down.sql

DROP TABLE IF EXISTS email_verifications;

ALTER TABLE users
    DROP COLUMN IF EXISTS password_hash,
    DROP COLUMN IF EXISTS email_verified,
    DROP COLUMN IF EXISTS email_verified_at;
