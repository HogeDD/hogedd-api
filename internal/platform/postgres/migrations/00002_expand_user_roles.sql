-- +goose Up
DROP INDEX users_email_lower_unique;

ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('owner', 'admin', 'member'));

-- +goose Down
-- This migration is intentionally forward-only. Removing roles or restoring email
-- uniqueness can discard valid identities and must be handled by a reviewed data migration.
SELECT 1;
