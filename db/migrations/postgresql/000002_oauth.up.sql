CREATE TABLE IF NOT EXISTS oauth_providers (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  provider TEXT,
  provider_user_id TEXT,
  UNIQUE(user_id, provider)
);