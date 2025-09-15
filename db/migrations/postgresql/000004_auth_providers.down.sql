BEGIN;

ALTER TABLE auth_providers RENAME TO oauth_providers;
ALTER TABLE oauth_providers ADD COLUMN id TEXT;

COMMIT;
