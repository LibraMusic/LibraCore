BEGIN;

DROP TABLE IF EXISTS tracks;
DROP TABLE IF EXISTS albums;
DROP TABLE IF EXISTS videos;
DROP TABLE IF EXISTS artists;
DROP TABLE IF EXISTS playlists;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS blacklisted_tokens;

COMMIT;