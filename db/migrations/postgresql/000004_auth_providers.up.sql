BEGIN;

ALTER TABLE oauth_providers DROP COLUMN id;
ALTER TABLE oauth_providers RENAME TO auth_providers;

COMMIT;
