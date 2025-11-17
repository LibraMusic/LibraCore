CREATE TABLE IF NOT EXISTS tracks (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  isrc TEXT,
  title TEXT,
  artist_ids TEXT, -- JSON (json array)
  album_ids TEXT, -- JSON (json array)
  primary_album_id TEXT,
  track_number INTEGER,
  duration INTEGER,
  description TEXT,
  release_date TEXT,
  lyrics TEXT, -- JSON
  listen_count INTEGER,
  favorite_count INTEGER,
  addition_date INTEGER,
  tags TEXT, -- JSON (json array)
  additional_meta BLOB, -- JSONB (json object)
  permissions BLOB, -- JSONB (json object)
  linked_item_ids TEXT, -- JSON (json array)
  content_source TEXT,
  metadata_source TEXT,
  lyric_sources BLOB -- JSONB (json object)
);

CREATE TABLE IF NOT EXISTS albums (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  upc TEXT,
  ean TEXT,
  title TEXT,
  artist_ids TEXT, -- JSON (json array)
  track_ids TEXT, -- JSON (json array)
  description TEXT,
  release_date TEXT,
  listen_count INTEGER,
  favorite_count INTEGER,
  addition_date INTEGER,
  tags TEXT, -- JSON (json array)
  additional_meta BLOB, -- JSONB (json object)
  permissions BLOB, -- JSONB (json object)
  linked_item_ids TEXT, -- JSON (json array)
  metadata_source TEXT
);

CREATE TABLE IF NOT EXISTS videos (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  title TEXT,
  artist_ids TEXT, -- JSON (json array)
  duration INTEGER,
  description TEXT,
  release_date TEXT,
  subtitles BLOB, -- JSONB (json object)
  watch_count INTEGER,
  favorite_count INTEGER,
  addition_date INTEGER,
  tags TEXT, -- JSON (json array)
  additional_meta BLOB, -- JSONB (json object)
  permissions BLOB, -- JSONB (json object)
  linked_item_ids TEXT, -- JSON (json array)
  content_source TEXT,
  metadata_source TEXT,
  lyric_sources BLOB -- JSONB (json object)
);

CREATE TABLE IF NOT EXISTS artists (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  name TEXT,
  album_ids TEXT, -- JSON (json array)
  track_ids TEXT, -- JSON (json array)
  description TEXT,
  creation_date TEXT,
  listen_count INTEGER,
  favorite_count INTEGER,
  addition_date INTEGER,
  tags TEXT, -- JSON (json array)
  additional_meta BLOB, -- JSONB (json object)
  permissions BLOB, -- JSONB (json object)
  linked_item_ids TEXT, -- JSON (json array)
  metadata_source TEXT
);

CREATE TABLE IF NOT EXISTS playlists (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  title TEXT,
  track_ids TEXT, -- JSON (json array)
  listen_count INTEGER,
  favorite_count INTEGER,
  description TEXT,
  creation_date TEXT,
  addition_date INTEGER,
  tags TEXT, -- JSON (json array)
  additional_meta BLOB, -- JSONB (json object)
  permissions BLOB, -- JSONB (json object)
  metadata_source TEXT
);

CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  email TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  display_name TEXT,
  description TEXT,
  listened_to BLOB, -- JSONB (json object)
  favorites TEXT, -- JSON (json array)
  public_view_count INTEGER,
  creation_date INTEGER,
  permissions BLOB, -- JSONB (json object)
  linked_artist_id TEXT,
  linked_sources BLOB -- JSONB (json object)
);

CREATE TABLE IF NOT EXISTS auth_providers (
  user_id TEXT,
  provider TEXT,
  provider_user_id TEXT,
  UNIQUE (user_id, provider)
);

CREATE TABLE IF NOT EXISTS blacklisted_tokens (
  token TEXT PRIMARY KEY,
  expiration TEXT
);
