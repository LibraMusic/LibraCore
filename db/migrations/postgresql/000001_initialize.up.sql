BEGIN;

CREATE TABLE IF NOT EXISTS tracks (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  isrc TEXT,
  title TEXT,
  artist_ids TEXT[],
  album_ids TEXT[],
  primary_album_id TEXT,
  track_number INT,
  duration INT,
  description TEXT,
  release_date TEXT,
  lyrics jsonb,
  listen_count INT,
  favorite_count INT,
  addition_date BIGINT,
  tags TEXT[],
  additional_meta jsonb,
  permissions jsonb,
  linked_item_ids TEXT[],
  content_source TEXT,
  metadata_source TEXT,
  lyric_sources jsonb
);

CREATE TABLE IF NOT EXISTS albums (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  upc TEXT,
  ean TEXT,
  title TEXT,
  artist_ids TEXT[],
  track_ids TEXT[],
  description TEXT,
  release_date TEXT,
  listen_count INT,
  favorite_count INT,
  addition_date BIGINT,
  tags TEXT[],
  additional_meta jsonb,
  permissions jsonb,
  linked_item_ids TEXT[],
  metadata_source TEXT
);

CREATE TABLE IF NOT EXISTS videos (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  title TEXT,
  artist_ids TEXT[],
  duration INT,
  description TEXT,
  release_date TEXT,
  subtitles jsonb,
  watch_count INT,
  favorite_count INT,
  addition_date BIGINT,
  tags TEXT[],
  additional_meta jsonb,
  permissions jsonb,
  linked_item_ids TEXT[],
  content_source TEXT,
  metadata_source TEXT,
  lyric_sources jsonb
);

CREATE TABLE IF NOT EXISTS artists (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  name TEXT,
  album_ids TEXT[],
  track_ids TEXT[],
  description TEXT,
  creation_date TEXT,
  listen_count INT,
  favorite_count INT,
  addition_date BIGINT,
  tags TEXT[],
  additional_meta jsonb,
  permissions jsonb,
  linked_item_ids TEXT[],
  metadata_source TEXT
);

CREATE TABLE IF NOT EXISTS playlists (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  title TEXT,
  track_ids TEXT[],
  listen_count INT,
  favorite_count INT,
  description TEXT,
  creation_date TEXT,
  addition_date BIGINT,
  tags TEXT[],
  additional_meta jsonb,
  permissions jsonb,
  metadata_source TEXT
);

CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  email TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  display_name TEXT,
  description TEXT,
  listened_to jsonb,
  favorites TEXT[],
  public_view_count INT,
  creation_date BIGINT,
  permissions jsonb,
  linked_artist_id TEXT,
  linked_sources jsonb
);

CREATE TABLE IF NOT EXISTS auth_providers (
  user_id TEXT,
  provider TEXT,
  provider_user_id TEXT,
  UNIQUE (user_id, provider)
);

CREATE TABLE IF NOT EXISTS blacklisted_tokens (
  token TEXT PRIMARY KEY,
  expiration TIMESTAMP
);

COMMIT;
