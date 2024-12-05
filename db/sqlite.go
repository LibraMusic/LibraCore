package db

import (
	"database/sql"
	"errors"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-json"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"

	"github.com/LibraMusic/LibraCore/config"
	"github.com/LibraMusic/LibraCore/types"
	"github.com/LibraMusic/LibraCore/utils"
)

type SQLiteDatabase struct {
	sqlDB *sql.DB
}

func ConnectSQLite() (*SQLiteDatabase, error) {
	result := &SQLiteDatabase{}
	err := result.Connect()
	return result, err
}

func (db *SQLiteDatabase) Connect() error {
	log.Info("Connecting to SQLite...")
	dbPath := config.Conf.Database.SQLite.Path
	if !filepath.IsAbs(dbPath) && utils.DataDir != "" {
		dbPath = filepath.Join(utils.DataDir, dbPath)
	}
	sqlDB, err := sql.Open("sqlite3", dbPath)
	db.sqlDB = sqlDB
	if err != nil {
		return err
	}

	if err = db.createTracksTable(); err != nil {
		return err
	}
	if err = db.createAlbumsTable(); err != nil {
		return err
	}
	if err = db.createVideosTable(); err != nil {
		return err
	}
	if err = db.createArtistsTable(); err != nil {
		return err
	}
	if err = db.createPlaylistsTable(); err != nil {
		return err
	}
	if err = db.createUsersTable(); err != nil {
		return err
	}
	if err = db.createBlacklistedTokensTable(); err != nil {
		return err
	}

	return nil
}

func (db *SQLiteDatabase) createTracksTable() error {
	_, err := db.sqlDB.Exec(`
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
  `)
	return err
}

func (db *SQLiteDatabase) createAlbumsTable() error {
	_, err := db.sqlDB.Exec(`
	  CREATE TABLE IF NOT EXISTS albums (
		  id TEXT PRIMARY KEY,
		  user_id TEXT,
		  upc TEXT,
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
	`)
	return err
}

func (db *SQLiteDatabase) createVideosTable() error {
	_, err := db.sqlDB.Exec(`
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
	`)
	return err
}

func (db *SQLiteDatabase) createArtistsTable() error {
	_, err := db.sqlDB.Exec(`
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
	`)
	return err
}

func (db *SQLiteDatabase) createPlaylistsTable() error {
	_, err := db.sqlDB.Exec(`
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
	`)
	return err
}

func (db *SQLiteDatabase) createUsersTable() error {
	_, err := db.sqlDB.Exec(`
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
	`)
	return err
}

func (db *SQLiteDatabase) createBlacklistedTokensTable() error {
	_, err := db.sqlDB.Exec(`
	  CREATE TABLE IF NOT EXISTS blacklisted_tokens (
		  token TEXT PRIMARY KEY,
		  expiration TEXT
	  );
	`)
	return err
}

func (db *SQLiteDatabase) Close() error {
	log.Info("Closing SQLite connection...")
	return db.sqlDB.Close()
}

func (*SQLiteDatabase) EngineName() string {
	return "SQLite"
}

func (db *SQLiteDatabase) MigrateUp(steps int) error {
	d, err := iofs.New(migrationsFS, "migrations/sqlite")
	if err != nil {
		return err
	}

	driver, err := sqlite3.WithInstance(db.sqlDB, &sqlite3.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", d, "sqlite", driver)
	if err != nil {
		return err
	}

	if steps <= 0 {
		return m.Up()
	} else {
		return m.Steps(steps)
	}
}

func (db *SQLiteDatabase) MigrateDown(steps int) error {
	d, err := iofs.New(migrationsFS, "migrations/sqlite")
	if err != nil {
		return err
	}

	driver, err := sqlite3.WithInstance(db.sqlDB, &sqlite3.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", d, "sqlite", driver)
	if err != nil {
		return err
	}

	if steps <= 0 {
		return m.Down()
	} else {
		return m.Steps(-steps)
	}
}

func (db *SQLiteDatabase) GetAllTracks() ([]types.Track, error) {
	var tracks []types.Track
	rows, err := db.sqlDB.Query("SELECT * FROM tracks;")
	if err != nil {
		return tracks, err
	}
	defer rows.Close()

	for rows.Next() {
		track := types.Track{}
		var artistIDs, albumIDs, tags, linkedItemIDs string
		var additionalMeta, permissions, lyrics, lyricSources string

		err = rows.Scan(
			&track.ID, &track.UserID, &track.ISRC, &track.Title,
			&artistIDs, &albumIDs, &track.PrimaryAlbumID, &track.TrackNumber,
			&track.Duration, &track.Description, &track.ReleaseDate, &lyrics,
			&track.ListenCount, &track.FavoriteCount, &track.AdditionDate,
			&tags, &additionalMeta, &permissions, &linkedItemIDs,
			&track.ContentSource, &track.MetadataSource, &lyricSources,
		)
		if err != nil {
			return tracks, err
		}

		// Parse JSON fields
		if err = json.Unmarshal([]byte(artistIDs), &track.ArtistIDs); err != nil {
			return tracks, errors.New("failed to parse artist_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(albumIDs), &track.AlbumIDs); err != nil {
			return tracks, errors.New("failed to parse album_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(tags), &track.Tags); err != nil {
			return tracks, errors.New("failed to parse tags: " + err.Error())
		}
		if err = json.Unmarshal([]byte(linkedItemIDs), &track.LinkedItemIDs); err != nil {
			return tracks, errors.New("failed to parse linked_item_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(additionalMeta), &track.AdditionalMeta); err != nil {
			return tracks, errors.New("failed to parse additional_meta: " + err.Error())
		}
		if err = json.Unmarshal([]byte(permissions), &track.Permissions); err != nil {
			return tracks, errors.New("failed to parse permissions: " + err.Error())
		}
		if err = json.Unmarshal([]byte(lyrics), &track.Lyrics); err != nil {
			return tracks, errors.New("failed to parse lyrics: " + err.Error())
		}
		if err = json.Unmarshal([]byte(lyricSources), &track.LyricSources); err != nil {
			return tracks, errors.New("failed to parse lyric_sources: " + err.Error())
		}

		tracks = append(tracks, track)
	}

	if err = rows.Err(); err != nil {
		return tracks, err
	}

	return tracks, err
}

func (db *SQLiteDatabase) GetTracks(userID string) ([]types.Track, error) {
	var tracks []types.Track
	rows, err := db.sqlDB.Query("SELECT * FROM tracks WHERE user_id = ?;", userID)
	if err != nil {
		return tracks, err
	}
	defer rows.Close()

	for rows.Next() {
		track := types.Track{}
		var artistIDs, albumIDs, tags, linkedItemIDs string
		var additionalMeta, permissions, lyrics, lyricSources string

		err = rows.Scan(
			&track.ID, &track.UserID, &track.ISRC, &track.Title,
			&artistIDs, &albumIDs, &track.PrimaryAlbumID, &track.TrackNumber,
			&track.Duration, &track.Description, &track.ReleaseDate, &lyrics,
			&track.ListenCount, &track.FavoriteCount, &track.AdditionDate,
			&tags, &additionalMeta, &permissions, &linkedItemIDs,
			&track.ContentSource, &track.MetadataSource, &lyricSources,
		)
		if err != nil {
			return tracks, err
		}

		// Parse JSON fields
		if err = json.Unmarshal([]byte(artistIDs), &track.ArtistIDs); err != nil {
			return tracks, errors.New("failed to parse artist_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(albumIDs), &track.AlbumIDs); err != nil {
			return tracks, errors.New("failed to parse album_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(tags), &track.Tags); err != nil {
			return tracks, errors.New("failed to parse tags: " + err.Error())
		}
		if err = json.Unmarshal([]byte(linkedItemIDs), &track.LinkedItemIDs); err != nil {
			return tracks, errors.New("failed to parse linked_item_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(additionalMeta), &track.AdditionalMeta); err != nil {
			return tracks, errors.New("failed to parse additional_meta: " + err.Error())
		}
		if err = json.Unmarshal([]byte(permissions), &track.Permissions); err != nil {
			return tracks, errors.New("failed to parse permissions: " + err.Error())
		}
		if err = json.Unmarshal([]byte(lyrics), &track.Lyrics); err != nil {
			return tracks, errors.New("failed to parse lyrics: " + err.Error())
		}
		if err = json.Unmarshal([]byte(lyricSources), &track.LyricSources); err != nil {
			return tracks, errors.New("failed to parse lyric_sources: " + err.Error())
		}

		tracks = append(tracks, track)
	}

	if err = rows.Err(); err != nil {
		return tracks, err
	}

	return tracks, err
}

func (db *SQLiteDatabase) GetTrack(id string) (types.Track, error) {
	track := types.Track{}
	var artistIDs, albumIDs, tags, linkedItemIDs string
	var additionalMeta, permissions, lyrics, lyricSources string

	row := db.sqlDB.QueryRow("SELECT * FROM tracks WHERE id = ?;", id)
	err := row.Scan(
		&track.ID, &track.UserID, &track.ISRC, &track.Title,
		&artistIDs, &albumIDs, &track.PrimaryAlbumID, &track.TrackNumber,
		&track.Duration, &track.Description, &track.ReleaseDate, &lyrics,
		&track.ListenCount, &track.FavoriteCount, &track.AdditionDate,
		&tags, &additionalMeta, &permissions, &linkedItemIDs,
		&track.ContentSource, &track.MetadataSource, &lyricSources,
	)
	if err != nil {
		return track, err
	}

	// Parse JSON fields
	if err = json.Unmarshal([]byte(artistIDs), &track.ArtistIDs); err != nil {
		return track, errors.New("failed to parse artist_ids: " + err.Error())
	}
	if err = json.Unmarshal([]byte(albumIDs), &track.AlbumIDs); err != nil {
		return track, errors.New("failed to parse album_ids: " + err.Error())
	}
	if err = json.Unmarshal([]byte(tags), &track.Tags); err != nil {
		return track, errors.New("failed to parse tags: " + err.Error())
	}
	if err = json.Unmarshal([]byte(linkedItemIDs), &track.LinkedItemIDs); err != nil {
		return track, errors.New("failed to parse linked_item_ids: " + err.Error())
	}
	if err = json.Unmarshal([]byte(additionalMeta), &track.AdditionalMeta); err != nil {
		return track, errors.New("failed to parse additional_meta: " + err.Error())
	}
	if err = json.Unmarshal([]byte(permissions), &track.Permissions); err != nil {
		return track, errors.New("failed to parse permissions: " + err.Error())
	}
	if err = json.Unmarshal([]byte(lyrics), &track.Lyrics); err != nil {
		return track, errors.New("failed to parse lyrics: " + err.Error())
	}
	if err = json.Unmarshal([]byte(lyricSources), &track.LyricSources); err != nil {
		return track, errors.New("failed to parse lyric_sources: " + err.Error())
	}

	return track, nil
}

func (db *SQLiteDatabase) AddTrack(track types.Track) error {
	// Convert JSON fields to strings
	artistIDs, err := json.Marshal(track.ArtistIDs)
	if err != nil {
		return errors.New("failed to marshal artist_ids: " + err.Error())
	}
	albumIDs, err := json.Marshal(track.AlbumIDs)
	if err != nil {
		return errors.New("failed to marshal album_ids: " + err.Error())
	}
	tags, err := json.Marshal(track.Tags)
	if err != nil {
		return errors.New("failed to marshal tags: " + err.Error())
	}
	linkedItemIDs, err := json.Marshal(track.LinkedItemIDs)
	if err != nil {
		return errors.New("failed to marshal linked_item_ids: " + err.Error())
	}
	additionalMeta, err := json.Marshal(track.AdditionalMeta)
	if err != nil {
		return errors.New("failed to marshal additional_meta: " + err.Error())
	}
	permissions, err := json.Marshal(track.Permissions)
	if err != nil {
		return errors.New("failed to marshal permissions: " + err.Error())
	}
	lyrics, err := json.Marshal(track.Lyrics)
	if err != nil {
		return errors.New("failed to marshal lyrics: " + err.Error())
	}
	lyricSources, err := json.Marshal(track.LyricSources)
	if err != nil {
		return errors.New("failed to marshal lyric_sources: " + err.Error())
	}

	_, err = db.sqlDB.Exec(`
	  INSERT INTO tracks (
	    id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
	  ) VALUES (
	    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	  );
	`,
		track.ID, track.UserID, track.ISRC, track.Title, string(artistIDs), string(albumIDs),
		track.PrimaryAlbumID, track.TrackNumber, track.Duration, track.Description,
		track.ReleaseDate, string(lyrics), track.ListenCount, track.FavoriteCount,
		track.AdditionDate, string(tags), string(additionalMeta), string(permissions),
		string(linkedItemIDs), track.ContentSource, track.MetadataSource, string(lyricSources),
	)

	return err
}

func (db *SQLiteDatabase) UpdateTrack(track types.Track) error {
	// Convert JSON fields to strings
	artistIDs, err := json.Marshal(track.ArtistIDs)
	if err != nil {
		return errors.New("failed to marshal artist_ids: " + err.Error())
	}
	albumIDs, err := json.Marshal(track.AlbumIDs)
	if err != nil {
		return errors.New("failed to marshal album_ids: " + err.Error())
	}
	tags, err := json.Marshal(track.Tags)
	if err != nil {
		return errors.New("failed to marshal tags: " + err.Error())
	}
	linkedItemIDs, err := json.Marshal(track.LinkedItemIDs)
	if err != nil {
		return errors.New("failed to marshal linked_item_ids: " + err.Error())
	}
	additionalMeta, err := json.Marshal(track.AdditionalMeta)
	if err != nil {
		return errors.New("failed to marshal additional_meta: " + err.Error())
	}
	permissions, err := json.Marshal(track.Permissions)
	if err != nil {
		return errors.New("failed to marshal permissions: " + err.Error())
	}
	lyrics, err := json.Marshal(track.Lyrics)
	if err != nil {
		return errors.New("failed to marshal lyrics: " + err.Error())
	}
	lyricSources, err := json.Marshal(track.LyricSources)
	if err != nil {
		return errors.New("failed to marshal lyric_sources: " + err.Error())
	}

	_, err = db.sqlDB.Exec(`
	  UPDATE tracks
	  SET user_id=?, isrc=?, title=?, artist_ids=?, album_ids=?, primary_album_id=?, 
	      track_number=?, duration=?, description=?, release_date=?, lyrics=?, 
	      listen_count=?, favorite_count=?, addition_date=?, tags=?, additional_meta=?, 
	      permissions=?, linked_item_ids=?, content_source=?, metadata_source=?, 
	      lyric_sources=?
	  WHERE id=?;
	`,
		track.UserID, track.ISRC, track.Title, string(artistIDs), string(albumIDs),
		track.PrimaryAlbumID, track.TrackNumber, track.Duration, track.Description,
		track.ReleaseDate, string(lyrics), track.ListenCount, track.FavoriteCount,
		track.AdditionDate, string(tags), string(additionalMeta), string(permissions),
		string(linkedItemIDs), track.ContentSource, track.MetadataSource,
		string(lyricSources), track.ID,
	)

	return err
}

func (db *SQLiteDatabase) DeleteTrack(id string) error {
	_, err := db.sqlDB.Exec("DELETE FROM tracks WHERE id = ?;", id)
	return err
}

func (db *SQLiteDatabase) GetAllAlbums() ([]types.Album, error) {
	var albums []types.Album
	rows, err := db.sqlDB.Query("SELECT * FROM albums;")
	if err != nil {
		return albums, err
	}
	defer rows.Close()

	for rows.Next() {
		album := types.Album{}
		var artistIDs, trackIDs, tags, linkedItemIDs string
		var additionalMeta, permissions string

		err = rows.Scan(
			&album.ID, &album.UserID, &album.UPC, &album.Title,
			&artistIDs, &trackIDs, &album.Description, &album.ReleaseDate,
			&album.ListenCount, &album.FavoriteCount, &album.AdditionDate,
			&tags, &additionalMeta, &permissions, &linkedItemIDs,
			&album.MetadataSource,
		)
		if err != nil {
			return albums, err
		}

		// Parse JSON fields
		if err = json.Unmarshal([]byte(artistIDs), &album.ArtistIDs); err != nil {
			return albums, errors.New("failed to parse artist_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(trackIDs), &album.TrackIDs); err != nil {
			return albums, errors.New("failed to parse track_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(tags), &album.Tags); err != nil {
			return albums, errors.New("failed to parse tags: " + err.Error())
		}
		if err = json.Unmarshal([]byte(linkedItemIDs), &album.LinkedItemIDs); err != nil {
			return albums, errors.New("failed to parse linked_item_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(additionalMeta), &album.AdditionalMeta); err != nil {
			return albums, errors.New("failed to parse additional_meta: " + err.Error())
		}
		if err = json.Unmarshal([]byte(permissions), &album.Permissions); err != nil {
			return albums, errors.New("failed to parse permissions: " + err.Error())
		}

		albums = append(albums, album)
	}

	if err = rows.Err(); err != nil {
		return albums, err
	}

	return albums, err
}

func (db *SQLiteDatabase) GetAlbums(userID string) ([]types.Album, error) {
	var albums []types.Album
	rows, err := db.sqlDB.Query("SELECT * FROM albums WHERE user_id = ?;", userID)
	if err != nil {
		return albums, err
	}

	for rows.Next() {
		album := types.Album{}
		var artistIDs, trackIDs, tags, linkedItemIDs string
		var additionalMeta, permissions string

		err = rows.Scan(
			&album.ID, &album.UserID, &album.UPC, &album.Title,
			&artistIDs, &trackIDs, &album.Description, &album.ReleaseDate,
			&album.ListenCount, &album.FavoriteCount, &album.AdditionDate,
			&tags, &additionalMeta, &permissions, &linkedItemIDs,
			&album.MetadataSource,
		)
		if err != nil {
			return albums, err
		}

		// Parse JSON fields
		if err = json.Unmarshal([]byte(artistIDs), &album.ArtistIDs); err != nil {
			return albums, errors.New("failed to parse artist_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(trackIDs), &album.TrackIDs); err != nil {
			return albums, errors.New("failed to parse track_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(tags), &album.Tags); err != nil {
			return albums, errors.New("failed to parse tags: " + err.Error())
		}
		if err = json.Unmarshal([]byte(linkedItemIDs), &album.LinkedItemIDs); err != nil {
			return albums, errors.New("failed to parse linked_item_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(additionalMeta), &album.AdditionalMeta); err != nil {
			return albums, errors.New("failed to parse additional_meta: " + err.Error())
		}
		if err = json.Unmarshal([]byte(permissions), &album.Permissions); err != nil {
			return albums, errors.New("failed to parse permissions: " + err.Error())
		}

		albums = append(albums, album)
	}

	if err = rows.Err(); err != nil {
		return albums, err
	}

	return albums, err
}

func (db *SQLiteDatabase) GetAlbum(id string) (types.Album, error) {
	album := types.Album{}
	var artistIDs, trackIDs, tags, linkedItemIDs string
	var additionalMeta, permissions string

	row := db.sqlDB.QueryRow("SELECT * FROM albums WHERE id = ?;", id)
	err := row.Scan(
		&album.ID, &album.UserID, &album.UPC, &album.Title,
		&artistIDs, &trackIDs, &album.Description, &album.ReleaseDate,
		&album.ListenCount, &album.FavoriteCount, &album.AdditionDate,
		&tags, &additionalMeta, &permissions, &linkedItemIDs,
		&album.MetadataSource,
	)
	if err != nil {
		return album, err
	}

	// Parse JSON fields
	if err = json.Unmarshal([]byte(artistIDs), &album.ArtistIDs); err != nil {
		return album, errors.New("failed to parse artist_ids: " + err.Error())
	}
	if err = json.Unmarshal([]byte(trackIDs), &album.TrackIDs); err != nil {
		return album, errors.New("failed to parse track_ids: " + err.Error())
	}
	if err = json.Unmarshal([]byte(tags), &album.Tags); err != nil {
		return album, errors.New("failed to parse tags: " + err.Error())
	}
	if err = json.Unmarshal([]byte(linkedItemIDs), &album.LinkedItemIDs); err != nil {
		return album, errors.New("failed to parse linked_item_ids: " + err.Error())
	}
	if err = json.Unmarshal([]byte(additionalMeta), &album.AdditionalMeta); err != nil {
		return album, errors.New("failed to parse additional_meta: " + err.Error())
	}
	if err = json.Unmarshal([]byte(permissions), &album.Permissions); err != nil {
		return album, errors.New("failed to parse permissions: " + err.Error())
	}

	return album, nil
}

func (db *SQLiteDatabase) AddAlbum(album types.Album) error {
	// Convert JSON fields to strings
	artistIDs, err := json.Marshal(album.ArtistIDs)
	if err != nil {
		return errors.New("failed to marshal artist_ids: " + err.Error())
	}
	trackIDs, err := json.Marshal(album.TrackIDs)
	if err != nil {
		return errors.New("failed to marshal track_ids: " + err.Error())
	}
	tags, err := json.Marshal(album.Tags)
	if err != nil {
		return errors.New("failed to marshal tags: " + err.Error())
	}
	linkedItemIDs, err := json.Marshal(album.LinkedItemIDs)
	if err != nil {
		return errors.New("failed to marshal linked_item_ids: " + err.Error())
	}
	additionalMeta, err := json.Marshal(album.AdditionalMeta)
	if err != nil {
		return errors.New("failed to marshal additional_meta: " + err.Error())
	}
	permissions, err := json.Marshal(album.Permissions)
	if err != nil {
		return errors.New("failed to marshal permissions: " + err.Error())
	}

	_, err = db.sqlDB.Exec(`
	  INSERT INTO albums (
	    id, user_id, upc, title, artist_ids, track_ids, description, release_date, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, metadata_source
	  ) VALUES (
	    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	  );
	`,
		album.ID, album.UserID, album.UPC, album.Title, string(artistIDs), string(trackIDs),
		album.Description, album.ReleaseDate, album.ListenCount, album.FavoriteCount,
		album.AdditionDate, string(tags), string(additionalMeta), string(permissions),
		string(linkedItemIDs), album.MetadataSource,
	)

	return err
}

func (db *SQLiteDatabase) UpdateAlbum(album types.Album) error {
	// Convert JSON fields to strings
	artistIDs, err := json.Marshal(album.ArtistIDs)
	if err != nil {
		return errors.New("failed to marshal artist_ids: " + err.Error())
	}
	trackIDs, err := json.Marshal(album.TrackIDs)
	if err != nil {
		return errors.New("failed to marshal track_ids: " + err.Error())
	}
	tags, err := json.Marshal(album.Tags)
	if err != nil {
		return errors.New("failed to marshal tags: " + err.Error())
	}
	linkedItemIDs, err := json.Marshal(album.LinkedItemIDs)
	if err != nil {
		return errors.New("failed to marshal linked_item_ids: " + err.Error())
	}
	additionalMeta, err := json.Marshal(album.AdditionalMeta)
	if err != nil {
		return errors.New("failed to marshal additional_meta: " + err.Error())
	}
	permissions, err := json.Marshal(album.Permissions)
	if err != nil {
		return errors.New("failed to marshal permissions: " + err.Error())
	}

	_, err = db.sqlDB.Exec(`
	  UPDATE albums
	  SET user_id=?, upc=?, title=?, artist_ids=?, track_ids=?, description=?, 
	      release_date=?, listen_count=?, favorite_count=?, addition_date=?, tags=?, 
	      additional_meta=?, permissions=?, linked_item_ids=?, metadata_source=?
	  WHERE id=?;
	`,
		album.UserID, album.UPC, album.Title, string(artistIDs), string(trackIDs),
		album.Description, album.ReleaseDate, album.ListenCount, album.FavoriteCount,
		album.AdditionDate, string(tags), string(additionalMeta), string(permissions),
		string(linkedItemIDs), album.MetadataSource, album.ID,
	)

	return err
}

func (db *SQLiteDatabase) DeleteAlbum(id string) error {
	_, err := db.sqlDB.Exec("DELETE FROM albums WHERE id = ?;", id)
	return err
}

func (db *SQLiteDatabase) GetAllVideos() ([]types.Video, error) {
	var videos []types.Video
	rows, err := db.sqlDB.Query("SELECT * FROM videos;")
	if err != nil {
		return videos, err
	}

	for rows.Next() {
		video := types.Video{}
		var artistIDs, tags, linkedItemIDs string
		var additionalMeta, permissions, subtitles, lyricSources string

		err = rows.Scan(
			&video.ID, &video.UserID, &video.Title, &artistIDs,
			&video.Duration, &video.Description, &video.ReleaseDate, &subtitles,
			&video.WatchCount, &video.FavoriteCount, &video.AdditionDate,
			&tags, &additionalMeta, &permissions, &linkedItemIDs,
			&video.ContentSource, &video.MetadataSource, &lyricSources,
		)
		if err != nil {
			return videos, err
		}

		// Parse JSON fields
		if err = json.Unmarshal([]byte(artistIDs), &video.ArtistIDs); err != nil {
			return videos, errors.New("failed to parse artist_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(tags), &video.Tags); err != nil {
			return videos, errors.New("failed to parse tags: " + err.Error())
		}
		if err = json.Unmarshal([]byte(linkedItemIDs), &video.LinkedItemIDs); err != nil {
			return videos, errors.New("failed to parse linked_item_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(additionalMeta), &video.AdditionalMeta); err != nil {
			return videos, errors.New("failed to parse additional_meta: " + err.Error())
		}
		if err = json.Unmarshal([]byte(permissions), &video.Permissions); err != nil {
			return videos, errors.New("failed to parse permissions: " + err.Error())
		}
		if err = json.Unmarshal([]byte(subtitles), &video.Subtitles); err != nil {
			return videos, errors.New("failed to parse subtitles: " + err.Error())
		}
		if err = json.Unmarshal([]byte(lyricSources), &video.LyricSources); err != nil {
			return videos, errors.New("failed to parse lyric_sources: " + err.Error())
		}

		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return videos, err
	}

	return videos, err
}

func (db *SQLiteDatabase) GetVideos(userID string) ([]types.Video, error) {
	var videos []types.Video
	rows, err := db.sqlDB.Query("SELECT * FROM videos WHERE user_id = ?;", userID)
	if err != nil {
		return videos, err
	}

	for rows.Next() {
		video := types.Video{}
		var artistIDs, tags, linkedItemIDs string
		var additionalMeta, permissions, subtitles, lyricSources string

		err = rows.Scan(
			&video.ID, &video.UserID, &video.Title, &artistIDs,
			&video.Duration, &video.Description, &video.ReleaseDate, &subtitles,
			&video.WatchCount, &video.FavoriteCount, &video.AdditionDate,
			&tags, &additionalMeta, &permissions, &linkedItemIDs,
			&video.ContentSource, &video.MetadataSource, &lyricSources,
		)
		if err != nil {
			return videos, err
		}

		// Parse JSON fields
		if err = json.Unmarshal([]byte(artistIDs), &video.ArtistIDs); err != nil {
			return videos, errors.New("failed to parse artist_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(tags), &video.Tags); err != nil {
			return videos, errors.New("failed to parse tags: " + err.Error())
		}
		if err = json.Unmarshal([]byte(linkedItemIDs), &video.LinkedItemIDs); err != nil {
			return videos, errors.New("failed to parse linked_item_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(additionalMeta), &video.AdditionalMeta); err != nil {
			return videos, errors.New("failed to parse additional_meta: " + err.Error())
		}
		if err = json.Unmarshal([]byte(permissions), &video.Permissions); err != nil {
			return videos, errors.New("failed to parse permissions: " + err.Error())
		}
		if err = json.Unmarshal([]byte(subtitles), &video.Subtitles); err != nil {
			return videos, errors.New("failed to parse subtitles: " + err.Error())
		}
		if err = json.Unmarshal([]byte(lyricSources), &video.LyricSources); err != nil {
			return videos, errors.New("failed to parse lyric_sources: " + err.Error())
		}

		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return videos, err
	}

	return videos, err
}

func (db *SQLiteDatabase) GetVideo(id string) (types.Video, error) {
	video := types.Video{}
	var artistIDs, tags, linkedItemIDs string
	var additionalMeta, permissions, subtitles, lyricSources string

	row := db.sqlDB.QueryRow("SELECT * FROM videos WHERE id = ?;", id)
	err := row.Scan(
		&video.ID, &video.UserID, &video.Title, &artistIDs,
		&video.Duration, &video.Description, &video.ReleaseDate, &subtitles,
		&video.WatchCount, &video.FavoriteCount, &video.AdditionDate,
		&tags, &additionalMeta, &permissions, &linkedItemIDs,
		&video.ContentSource, &video.MetadataSource, &lyricSources,
	)
	if err != nil {
		return video, err
	}

	// Parse JSON fields
	if err = json.Unmarshal([]byte(artistIDs), &video.ArtistIDs); err != nil {
		return video, errors.New("failed to parse artist_ids: " + err.Error())
	}
	if err = json.Unmarshal([]byte(tags), &video.Tags); err != nil {
		return video, errors.New("failed to parse tags: " + err.Error())
	}
	if err = json.Unmarshal([]byte(linkedItemIDs), &video.LinkedItemIDs); err != nil {
		return video, errors.New("failed to parse linked_item_ids: " + err.Error())
	}
	if err = json.Unmarshal([]byte(additionalMeta), &video.AdditionalMeta); err != nil {
		return video, errors.New("failed to parse additional_meta: " + err.Error())
	}
	if err = json.Unmarshal([]byte(permissions), &video.Permissions); err != nil {
		return video, errors.New("failed to parse permissions: " + err.Error())
	}
	if err = json.Unmarshal([]byte(subtitles), &video.Subtitles); err != nil {
		return video, errors.New("failed to parse subtitles: " + err.Error())
	}
	if err = json.Unmarshal([]byte(lyricSources), &video.LyricSources); err != nil {
		return video, errors.New("failed to parse lyric_sources: " + err.Error())
	}

	return video, nil
}

func (db *SQLiteDatabase) AddVideo(video types.Video) error {
	// Convert JSON fields to strings
	artistIDs, err := json.Marshal(video.ArtistIDs)
	if err != nil {
		return errors.New("failed to marshal artist_ids: " + err.Error())
	}
	tags, err := json.Marshal(video.Tags)
	if err != nil {
		return errors.New("failed to marshal tags: " + err.Error())
	}
	linkedItemIDs, err := json.Marshal(video.LinkedItemIDs)
	if err != nil {
		return errors.New("failed to marshal linked_item_ids: " + err.Error())
	}
	additionalMeta, err := json.Marshal(video.AdditionalMeta)
	if err != nil {
		return errors.New("failed to marshal additional_meta: " + err.Error())
	}
	permissions, err := json.Marshal(video.Permissions)
	if err != nil {
		return errors.New("failed to marshal permissions: " + err.Error())
	}
	subtitles, err := json.Marshal(video.Subtitles)
	if err != nil {
		return errors.New("failed to marshal subtitles: " + err.Error())
	}
	lyricSources, err := json.Marshal(video.LyricSources)
	if err != nil {
		return errors.New("failed to marshal lyric_sources: " + err.Error())
	}

	_, err = db.sqlDB.Exec(`
	  INSERT INTO videos (
	    id, user_id, title, artist_ids, duration, description, release_date, subtitles, watch_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
	  ) VALUES (
	   	?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	  );
	`,
		video.ID, video.UserID, video.Title, string(artistIDs), video.Duration,
		video.Description, video.ReleaseDate, string(subtitles), video.WatchCount,
		video.FavoriteCount, video.AdditionDate, string(tags), string(additionalMeta),
		string(permissions), string(linkedItemIDs), video.ContentSource,
		video.MetadataSource, string(lyricSources),
	)

	return err
}

func (db *SQLiteDatabase) UpdateVideo(video types.Video) error {
	// Convert JSON fields to strings
	artistIDs, err := json.Marshal(video.ArtistIDs)
	if err != nil {
		return errors.New("failed to marshal artist_ids: " + err.Error())
	}
	tags, err := json.Marshal(video.Tags)
	if err != nil {
		return errors.New("failed to marshal tags: " + err.Error())
	}
	linkedItemIDs, err := json.Marshal(video.LinkedItemIDs)
	if err != nil {
		return errors.New("failed to marshal linked_item_ids: " + err.Error())
	}
	additionalMeta, err := json.Marshal(video.AdditionalMeta)
	if err != nil {
		return errors.New("failed to marshal additional_meta: " + err.Error())
	}
	permissions, err := json.Marshal(video.Permissions)
	if err != nil {
		return errors.New("failed to marshal permissions: " + err.Error())
	}
	subtitles, err := json.Marshal(video.Subtitles)
	if err != nil {
		return errors.New("failed to marshal subtitles: " + err.Error())
	}
	lyricSources, err := json.Marshal(video.LyricSources)
	if err != nil {
		return errors.New("failed to marshal lyric_sources: " + err.Error())
	}

	_, err = db.sqlDB.Exec(`
	  UPDATE videos
	  SET user_id=?, title=?, artist_ids=?, duration=?, description=?, release_date=?, 
	      subtitles=?, watch_count=?, favorite_count=?, addition_date=?, tags=?, 
	      additional_meta=?, permissions=?, linked_item_ids=?, content_source=?, 
	      metadata_source=?, lyric_sources=?
	  WHERE id=?;
	`,
		video.UserID, video.Title, string(artistIDs), video.Duration, video.Description,
		video.ReleaseDate, string(subtitles), video.WatchCount, video.FavoriteCount,
		video.AdditionDate, string(tags), string(additionalMeta), string(permissions),
		string(linkedItemIDs), video.ContentSource, video.MetadataSource,
		string(lyricSources), video.ID,
	)

	return err
}

func (db *SQLiteDatabase) DeleteVideo(id string) error {
	_, err := db.sqlDB.Exec("DELETE FROM videos WHERE id = ?;", id)
	return err
}

func (db *SQLiteDatabase) GetAllArtists() ([]types.Artist, error) {
	var artists []types.Artist
	rows, err := db.sqlDB.Query("SELECT * FROM artists;")
	if err != nil {
		return artists, err
	}

	for rows.Next() {
		artist := types.Artist{}
		var albumIDs, trackIDs, tags, linkedItemIDs string
		var additionalMeta, permissions string

		err = rows.Scan(
			&artist.ID, &artist.UserID, &artist.Name, &albumIDs, &trackIDs,
			&artist.Description, &artist.CreationDate, &artist.ListenCount,
			&artist.FavoriteCount, &artist.AdditionDate, &tags,
			&additionalMeta, &permissions, &linkedItemIDs, &artist.MetadataSource,
		)
		if err != nil {
			return artists, err
		}

		// Parse JSON fields
		if err = json.Unmarshal([]byte(albumIDs), &artist.AlbumIDs); err != nil {
			return artists, errors.New("failed to parse album_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(trackIDs), &artist.TrackIDs); err != nil {
			return artists, errors.New("failed to parse track_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(tags), &artist.Tags); err != nil {
			return artists, errors.New("failed to parse tags: " + err.Error())
		}
		if err = json.Unmarshal([]byte(linkedItemIDs), &artist.LinkedItemIDs); err != nil {
			return artists, errors.New("failed to parse linked_item_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(additionalMeta), &artist.AdditionalMeta); err != nil {
			return artists, errors.New("failed to parse additional_meta: " + err.Error())
		}
		if err = json.Unmarshal([]byte(permissions), &artist.Permissions); err != nil {
			return artists, errors.New("failed to parse permissions: " + err.Error())
		}

		artists = append(artists, artist)
	}

	if err = rows.Err(); err != nil {
		return artists, err
	}

	return artists, err
}

func (db *SQLiteDatabase) GetArtists(userID string) ([]types.Artist, error) {
	var artists []types.Artist
	rows, err := db.sqlDB.Query("SELECT * FROM artists WHERE user_id = ?;", userID)
	if err != nil {
		return artists, err
	}

	for rows.Next() {
		artist := types.Artist{}
		var albumIDs, trackIDs, tags, linkedItemIDs string
		var additionalMeta, permissions string

		err = rows.Scan(
			&artist.ID, &artist.UserID, &artist.Name, &albumIDs, &trackIDs,
			&artist.Description, &artist.CreationDate, &artist.ListenCount,
			&artist.FavoriteCount, &artist.AdditionDate, &tags,
			&additionalMeta, &permissions, &linkedItemIDs, &artist.MetadataSource,
		)
		if err != nil {
			return artists, err
		}

		// Parse JSON fields
		if err = json.Unmarshal([]byte(albumIDs), &artist.AlbumIDs); err != nil {
			return artists, errors.New("failed to parse album_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(trackIDs), &artist.TrackIDs); err != nil {
			return artists, errors.New("failed to parse track_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(tags), &artist.Tags); err != nil {
			return artists, errors.New("failed to parse tags: " + err.Error())
		}
		if err = json.Unmarshal([]byte(linkedItemIDs), &artist.LinkedItemIDs); err != nil {
			return artists, errors.New("failed to parse linked_item_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(additionalMeta), &artist.AdditionalMeta); err != nil {
			return artists, errors.New("failed to parse additional_meta: " + err.Error())
		}
		if err = json.Unmarshal([]byte(permissions), &artist.Permissions); err != nil {
			return artists, errors.New("failed to parse permissions: " + err.Error())
		}

		artists = append(artists, artist)
	}

	if err = rows.Err(); err != nil {
		return artists, err
	}

	return artists, err
}

func (db *SQLiteDatabase) GetArtist(id string) (types.Artist, error) {
	artist := types.Artist{}
	var albumIDs, trackIDs, tags, linkedItemIDs string
	var additionalMeta, permissions string

	row := db.sqlDB.QueryRow("SELECT * FROM artists WHERE id = ?;", id)
	err := row.Scan(
		&artist.ID, &artist.UserID, &artist.Name, &albumIDs, &trackIDs,
		&artist.Description, &artist.CreationDate, &artist.ListenCount,
		&artist.FavoriteCount, &artist.AdditionDate, &tags,
		&additionalMeta, &permissions, &linkedItemIDs, &artist.MetadataSource,
	)
	if err != nil {
		return artist, err
	}

	// Parse JSON fields
	if err = json.Unmarshal([]byte(albumIDs), &artist.AlbumIDs); err != nil {
		return artist, errors.New("failed to parse album_ids: " + err.Error())
	}
	if err = json.Unmarshal([]byte(trackIDs), &artist.TrackIDs); err != nil {
		return artist, errors.New("failed to parse track_ids: " + err.Error())
	}
	if err = json.Unmarshal([]byte(tags), &artist.Tags); err != nil {
		return artist, errors.New("failed to parse tags: " + err.Error())
	}
	if err = json.Unmarshal([]byte(linkedItemIDs), &artist.LinkedItemIDs); err != nil {
		return artist, errors.New("failed to parse linked_item_ids: " + err.Error())
	}
	if err = json.Unmarshal([]byte(additionalMeta), &artist.AdditionalMeta); err != nil {
		return artist, errors.New("failed to parse additional_meta: " + err.Error())
	}
	if err = json.Unmarshal([]byte(permissions), &artist.Permissions); err != nil {
		return artist, errors.New("failed to parse permissions: " + err.Error())
	}

	return artist, nil
}

func (db *SQLiteDatabase) AddArtist(artist types.Artist) error {
	// Convert JSON fields to strings
	albumIDs, err := json.Marshal(artist.AlbumIDs)
	if err != nil {
		return errors.New("failed to marshal album_ids: " + err.Error())
	}
	trackIDs, err := json.Marshal(artist.TrackIDs)
	if err != nil {
		return errors.New("failed to marshal track_ids: " + err.Error())
	}
	tags, err := json.Marshal(artist.Tags)
	if err != nil {
		return errors.New("failed to marshal tags: " + err.Error())
	}
	linkedItemIDs, err := json.Marshal(artist.LinkedItemIDs)
	if err != nil {
		return errors.New("failed to marshal linked_item_ids: " + err.Error())
	}
	additionalMeta, err := json.Marshal(artist.AdditionalMeta)
	if err != nil {
		return errors.New("failed to marshal additional_meta: " + err.Error())
	}
	permissions, err := json.Marshal(artist.Permissions)
	if err != nil {
		return errors.New("failed to marshal permissions: " + err.Error())
	}

	_, err = db.sqlDB.Exec(`
	  INSERT INTO artists (
	    id, user_id, name, album_ids, track_ids, description, creation_date, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, metadata_source
	  ) VALUES (
	   	?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	  );
	`,
		artist.ID, artist.UserID, artist.Name, string(albumIDs), string(trackIDs),
		artist.Description, artist.CreationDate, artist.ListenCount, artist.FavoriteCount,
		artist.AdditionDate, string(tags), string(additionalMeta), string(permissions),
		string(linkedItemIDs), artist.MetadataSource,
	)

	return err
}

func (db *SQLiteDatabase) UpdateArtist(artist types.Artist) error {
	// Convert JSON fields to strings
	albumIDs, err := json.Marshal(artist.AlbumIDs)
	if err != nil {
		return errors.New("failed to marshal album_ids: " + err.Error())
	}
	trackIDs, err := json.Marshal(artist.TrackIDs)
	if err != nil {
		return errors.New("failed to marshal track_ids: " + err.Error())
	}
	tags, err := json.Marshal(artist.Tags)
	if err != nil {
		return errors.New("failed to marshal tags: " + err.Error())
	}
	linkedItemIDs, err := json.Marshal(artist.LinkedItemIDs)
	if err != nil {
		return errors.New("failed to marshal linked_item_ids: " + err.Error())
	}
	additionalMeta, err := json.Marshal(artist.AdditionalMeta)
	if err != nil {
		return errors.New("failed to marshal additional_meta: " + err.Error())
	}
	permissions, err := json.Marshal(artist.Permissions)
	if err != nil {
		return errors.New("failed to marshal permissions: " + err.Error())
	}

	_, err = db.sqlDB.Exec(`
	  UPDATE artists
	  SET user_id=?, name=?, album_ids=?, track_ids=?, description=?, creation_date=?,
	      listen_count=?, favorite_count=?, addition_date=?, tags=?, additional_meta=?,
		  permissions=?, linked_item_ids=?, metadata_source=?
	  WHERE id=?;
	`,
		artist.UserID, artist.Name, string(albumIDs), string(trackIDs),
		artist.Description, artist.CreationDate, artist.ListenCount, artist.FavoriteCount,
		artist.AdditionDate, string(tags), string(additionalMeta), string(permissions),
		string(linkedItemIDs), artist.MetadataSource, artist.ID,
	)

	return err
}

func (db *SQLiteDatabase) DeleteArtist(id string) error {
	_, err := db.sqlDB.Exec("DELETE FROM artists WHERE id = ?;", id)
	return err
}

func (db *SQLiteDatabase) GetAllPlaylists() ([]types.Playlist, error) {
	var playlists []types.Playlist
	rows, err := db.sqlDB.Query("SELECT * FROM playlists;")
	if err != nil {
		return playlists, err
	}

	for rows.Next() {
		playlist := types.Playlist{}
		var trackIDs, tags string
		var additionalMeta, permissions string

		err = rows.Scan(
			&playlist.ID, &playlist.UserID, &playlist.Title, &trackIDs,
			&playlist.ListenCount, &playlist.FavoriteCount, &playlist.Description,
			&playlist.CreationDate, &playlist.AdditionDate, &tags,
			&additionalMeta, &permissions, &playlist.MetadataSource,
		)
		if err != nil {
			return playlists, err
		}

		// Parse JSON fields
		if err = json.Unmarshal([]byte(trackIDs), &playlist.TrackIDs); err != nil {
			return playlists, errors.New("failed to parse track_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(tags), &playlist.Tags); err != nil {
			return playlists, errors.New("failed to parse tags: " + err.Error())
		}
		if err = json.Unmarshal([]byte(additionalMeta), &playlist.AdditionalMeta); err != nil {
			return playlists, errors.New("failed to parse additional_meta: " + err.Error())
		}
		if err = json.Unmarshal([]byte(permissions), &playlist.Permissions); err != nil {
			return playlists, errors.New("failed to parse permissions: " + err.Error())
		}

		playlists = append(playlists, playlist)
	}

	if err = rows.Err(); err != nil {
		return playlists, err
	}

	return playlists, err
}

func (db *SQLiteDatabase) GetPlaylists(userID string) ([]types.Playlist, error) {
	var playlists []types.Playlist
	rows, err := db.sqlDB.Query("SELECT * FROM playlists WHERE user_id = ?;", userID)
	if err != nil {
		return playlists, err
	}

	for rows.Next() {
		playlist := types.Playlist{}
		var trackIDs, tags string
		var additionalMeta, permissions string

		err = rows.Scan(
			&playlist.ID, &playlist.UserID, &playlist.Title, &trackIDs,
			&playlist.ListenCount, &playlist.FavoriteCount, &playlist.Description,
			&playlist.CreationDate, &playlist.AdditionDate, &tags,
			&additionalMeta, &permissions, &playlist.MetadataSource,
		)
		if err != nil {
			return playlists, err
		}

		// Parse JSON fields
		if err = json.Unmarshal([]byte(trackIDs), &playlist.TrackIDs); err != nil {
			return playlists, errors.New("failed to parse track_ids: " + err.Error())
		}
		if err = json.Unmarshal([]byte(tags), &playlist.Tags); err != nil {
			return playlists, errors.New("failed to parse tags: " + err.Error())
		}
		if err = json.Unmarshal([]byte(additionalMeta), &playlist.AdditionalMeta); err != nil {
			return playlists, errors.New("failed to parse additional_meta: " + err.Error())
		}
		if err = json.Unmarshal([]byte(permissions), &playlist.Permissions); err != nil {
			return playlists, errors.New("failed to parse permissions: " + err.Error())
		}

		playlists = append(playlists, playlist)
	}

	if err = rows.Err(); err != nil {
		return playlists, err
	}

	return playlists, err
}

func (db *SQLiteDatabase) GetPlaylist(id string) (types.Playlist, error) {
	playlist := types.Playlist{}
	var trackIDs, tags string
	var additionalMeta, permissions string

	row := db.sqlDB.QueryRow("SELECT * FROM playlists WHERE id = ?;", id)
	err := row.Scan(
		&playlist.ID, &playlist.UserID, &playlist.Title, &trackIDs,
		&playlist.ListenCount, &playlist.FavoriteCount, &playlist.Description,
		&playlist.CreationDate, &playlist.AdditionDate, &tags,
		&additionalMeta, &permissions, &playlist.MetadataSource,
	)
	if err != nil {
		return playlist, err
	}

	// Parse JSON fields
	if err = json.Unmarshal([]byte(trackIDs), &playlist.TrackIDs); err != nil {
		return playlist, errors.New("failed to parse track_ids: " + err.Error())
	}
	if err = json.Unmarshal([]byte(tags), &playlist.Tags); err != nil {
		return playlist, errors.New("failed to parse tags: " + err.Error())
	}
	if err = json.Unmarshal([]byte(additionalMeta), &playlist.AdditionalMeta); err != nil {
		return playlist, errors.New("failed to parse additional_meta: " + err.Error())
	}
	if err = json.Unmarshal([]byte(permissions), &playlist.Permissions); err != nil {
		return playlist, errors.New("failed to parse permissions: " + err.Error())
	}

	return playlist, nil
}

func (db *SQLiteDatabase) AddPlaylist(playlist types.Playlist) error {
	// Convert JSON fields to strings
	trackIDs, err := json.Marshal(playlist.TrackIDs)
	if err != nil {
		return errors.New("failed to marshal track_ids: " + err.Error())
	}
	tags, err := json.Marshal(playlist.Tags)
	if err != nil {
		return errors.New("failed to marshal tags: " + err.Error())
	}
	additionalMeta, err := json.Marshal(playlist.AdditionalMeta)
	if err != nil {
		return errors.New("failed to marshal additional_meta: " + err.Error())
	}
	permissions, err := json.Marshal(playlist.Permissions)
	if err != nil {
		return errors.New("failed to marshal permissions: " + err.Error())
	}

	_, err = db.sqlDB.Exec(`
	  INSERT INTO playlists (
	    id, user_id, title, track_ids, listen_count, favorite_count, description, 
	    creation_date, addition_date, tags, additional_meta, permissions, metadata_source
	  ) VALUES (
	   	?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	  );
	`,
		playlist.ID, playlist.UserID, playlist.Title, string(trackIDs),
		playlist.ListenCount, playlist.FavoriteCount, playlist.Description,
		playlist.CreationDate, playlist.AdditionDate, string(tags),
		string(additionalMeta), string(permissions), playlist.MetadataSource,
	)

	return err
}

func (db *SQLiteDatabase) UpdatePlaylist(playlist types.Playlist) error {
	// Convert JSON fields to strings
	trackIDs, err := json.Marshal(playlist.TrackIDs)
	if err != nil {
		return errors.New("failed to marshal track_ids: " + err.Error())
	}
	tags, err := json.Marshal(playlist.Tags)
	if err != nil {
		return errors.New("failed to marshal tags: " + err.Error())
	}
	additionalMeta, err := json.Marshal(playlist.AdditionalMeta)
	if err != nil {
		return errors.New("failed to marshal additional_meta: " + err.Error())
	}
	permissions, err := json.Marshal(playlist.Permissions)
	if err != nil {
		return errors.New("failed to marshal permissions: " + err.Error())
	}

	_, err = db.sqlDB.Exec(`
	  UPDATE playlists
	  SET user_id=?, title=?, track_ids=?, listen_count=?, favorite_count=?, description=?, 
	      creation_date=?, addition_date=?, tags=?, additional_meta=?, permissions=?, 
	      metadata_source=?
	  WHERE id=?;
	`,
		playlist.UserID, playlist.Title, string(trackIDs),
		playlist.ListenCount, playlist.FavoriteCount, playlist.Description,
		playlist.CreationDate, playlist.AdditionDate, string(tags),
		string(additionalMeta), string(permissions), playlist.MetadataSource,
		playlist.ID,
	)

	return err
}

func (db *SQLiteDatabase) DeletePlaylist(id string) error {
	_, err := db.sqlDB.Exec("DELETE FROM playlists WHERE id = ?;", id)
	return err
}

func (db *SQLiteDatabase) GetUsers() ([]types.User, error) {
	var users []types.User
	rows, err := db.sqlDB.Query("SELECT * FROM users;")
	if err != nil {
		return users, err
	}

	for rows.Next() {
		user := types.User{}
		var listenedTo, favorites, permissions, linkedSources string

		err = rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.DisplayName, &user.Description, &listenedTo, &favorites,
			&user.PublicViewCount, &user.CreationDate, &permissions,
			&user.LinkedArtistID, &linkedSources,
		)
		if err != nil {
			return users, err
		}

		// Parse JSON fields
		if err = json.Unmarshal([]byte(listenedTo), &user.ListenedTo); err != nil {
			return users, errors.New("failed to parse listened_to: " + err.Error())
		}
		if err = json.Unmarshal([]byte(favorites), &user.Favorites); err != nil {
			return users, errors.New("failed to parse favorites: " + err.Error())
		}
		if err = json.Unmarshal([]byte(permissions), &user.Permissions); err != nil {
			return users, errors.New("failed to parse permissions: " + err.Error())
		}
		if err = json.Unmarshal([]byte(linkedSources), &user.LinkedSources); err != nil {
			return users, errors.New("failed to parse linked_sources: " + err.Error())
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return users, err
	}

	return users, err
}

func (db *SQLiteDatabase) GetUser(id string) (types.User, error) {
	user := types.User{}
	var listenedTo, favorites, permissions, linkedSources string

	row := db.sqlDB.QueryRow("SELECT * FROM users WHERE id = ?;", id)
	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.Description, &listenedTo, &favorites,
		&user.PublicViewCount, &user.CreationDate, &permissions,
		&user.LinkedArtistID, &linkedSources,
	)
	if err != nil {
		return user, err
	}

	// Parse JSON fields
	if err = json.Unmarshal([]byte(listenedTo), &user.ListenedTo); err != nil {
		return user, errors.New("failed to parse listened_to: " + err.Error())
	}
	if err = json.Unmarshal([]byte(favorites), &user.Favorites); err != nil {
		return user, errors.New("failed to parse favorites: " + err.Error())
	}
	if err = json.Unmarshal([]byte(permissions), &user.Permissions); err != nil {
		return user, errors.New("failed to parse permissions: " + err.Error())
	}
	if err = json.Unmarshal([]byte(linkedSources), &user.LinkedSources); err != nil {
		return user, errors.New("failed to parse linked_sources: " + err.Error())
	}

	return user, nil
}

func (db *SQLiteDatabase) GetUserByUsername(username string) (types.User, error) {
	user := types.User{}
	var listenedTo, favorites, permissions, linkedSources string

	row := db.sqlDB.QueryRow("SELECT * FROM users WHERE username = ?;", username)
	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.Description, &listenedTo, &favorites,
		&user.PublicViewCount, &user.CreationDate, &permissions,
		&user.LinkedArtistID, &linkedSources,
	)
	if err != nil {
		return user, err
	}

	// Parse JSON fields
	if err = json.Unmarshal([]byte(listenedTo), &user.ListenedTo); err != nil {
		return user, errors.New("failed to parse listened_to: " + err.Error())
	}
	if err = json.Unmarshal([]byte(favorites), &user.Favorites); err != nil {
		return user, errors.New("failed to parse favorites: " + err.Error())
	}
	if err = json.Unmarshal([]byte(permissions), &user.Permissions); err != nil {
		return user, errors.New("failed to parse permissions: " + err.Error())
	}
	if err = json.Unmarshal([]byte(linkedSources), &user.LinkedSources); err != nil {
		return user, errors.New("failed to parse linked_sources: " + err.Error())
	}

	return user, nil
}

func (db *SQLiteDatabase) CreateUser(user types.User) error {
	// Convert JSON fields to strings
	listenedTo, err := json.Marshal(user.ListenedTo)
	if err != nil {
		return errors.New("failed to marshal listened_to: " + err.Error())
	}
	favorites, err := json.Marshal(user.Favorites)
	if err != nil {
		return errors.New("failed to marshal favorites: " + err.Error())
	}
	permissions, err := json.Marshal(user.Permissions)
	if err != nil {
		return errors.New("failed to marshal permissions: " + err.Error())
	}
	linkedSources, err := json.Marshal(user.LinkedSources)
	if err != nil {
		return errors.New("failed to marshal linked_sources: " + err.Error())
	}

	_, err = db.sqlDB.Exec(`
	  INSERT INTO users (
	    id, username, email, password_hash, display_name, description, listened_to, 
	    favorites, public_view_count, creation_date, permissions, linked_artist_id, 
	    linked_sources
	  ) VALUES (
	   ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	  );
	`,
		user.ID, user.Username, user.Email, user.PasswordHash, user.DisplayName,
		user.Description, string(listenedTo), string(favorites), user.PublicViewCount,
		user.CreationDate, string(permissions), user.LinkedArtistID, string(linkedSources),
	)

	return err
}

func (db *SQLiteDatabase) UpdateUser(user types.User) error {
	// Convert JSON fields to strings
	listenedTo, err := json.Marshal(user.ListenedTo)
	if err != nil {
		return errors.New("failed to marshal listened_to: " + err.Error())
	}
	favorites, err := json.Marshal(user.Favorites)
	if err != nil {
		return errors.New("failed to marshal favorites: " + err.Error())
	}
	permissions, err := json.Marshal(user.Permissions)
	if err != nil {
		return errors.New("failed to marshal permissions: " + err.Error())
	}
	linkedSources, err := json.Marshal(user.LinkedSources)
	if err != nil {
		return errors.New("failed to marshal linked_sources: " + err.Error())
	}

	_, err = db.sqlDB.Exec(`
	  UPDATE users
	  SET username=?, email=?, password_hash=?, display_name=?, description=?, listened_to=?, 
	      favorites=?, public_view_count=?, creation_date=?, permissions=?, linked_artist_id=?, 
	      linked_sources=?
	  WHERE id=?;
	`,
		user.Username, user.Email, user.PasswordHash, user.DisplayName, user.Description,
		string(listenedTo), string(favorites), user.PublicViewCount, user.CreationDate,
		string(permissions), user.LinkedArtistID, string(linkedSources), user.ID,
	)

	return err
}

func (db *SQLiteDatabase) DeleteUser(id string) error {
	_, err := db.sqlDB.Exec("DELETE FROM users WHERE id = ?;", id)
	return err
}

func (db *SQLiteDatabase) UsernameExists(username string) (bool, error) {
	var exists bool
	err := db.sqlDB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?);", username).Scan(&exists)
	return exists, err
}

func (db *SQLiteDatabase) EmailExists(email string) (bool, error) {
	var exists bool
	err := db.sqlDB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?);", email).Scan(&exists)
	return exists, err
}

func (db *SQLiteDatabase) BlacklistToken(token string, expiration time.Time) error {
	_, err := db.sqlDB.Exec("INSERT INTO blacklisted_tokens (token, expiration) VALUES (?, ?);", token, expiration.Format(time.DateTime))
	return err
}

func (db *SQLiteDatabase) CleanExpiredTokens() error {
	_, err := db.sqlDB.Exec("DELETE FROM blacklisted_tokens WHERE expiration < datetime('now');")
	return err
}

func (db *SQLiteDatabase) IsTokenBlacklisted(token string) (bool, error) {
	var exists bool
	err := db.sqlDB.QueryRow("SELECT EXISTS(SELECT 1 FROM blacklisted_tokens WHERE token = ?);", token).Scan(&exists)
	return exists, err
}
