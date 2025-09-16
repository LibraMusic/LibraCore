//go:build sqlite_db || !(no_sqlite_db || no_dbs)

package db

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-json"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/media"
)

type SQLiteDatabase struct {
	pool *sqlitex.Pool
}

func (*SQLiteDatabase) Satisfies(engine string) bool {
	return slices.Contains([]string{
		"sqlite",
		"sqlite3",
	}, strings.ToLower(engine))
}

func (db *SQLiteDatabase) Connect() error {
	log.Info("Connecting to SQLite...")
	dbPath := config.Conf.Database.SQLite.Path
	if !filepath.IsAbs(dbPath) && config.DataDir != "" {
		dbPath = filepath.Join(config.DataDir, dbPath)
	}
	pool, err := sqlitex.NewPool(dbPath, sqlitex.PoolOptions{})
	db.pool = pool
	if err != nil {
		return err
	}

	// If the migrations table doesn't exist, create it and run migrations.
	exists, err := db.migrationsTableExists(context.Background())
	if err != nil {
		return err
	}
	if !exists {
		if err = db.createMigrationsTable(context.Background()); err != nil {
			return err
		}
		if err = db.MigrateUp(-1); err != nil {
			return err
		}
	}

	return nil
}

func (db *SQLiteDatabase) Close() error {
	log.Info("Closing SQLite connection...")
	err := db.pool.Close()
	return err
}

func (*SQLiteDatabase) EngineName() string {
	return "SQLite"
}

func (db *SQLiteDatabase) migrationsTableExists(ctx context.Context) (bool, error) {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return false, err
	}
	defer db.pool.Put(conn)

	var exists bool
	err = sqlitex.Execute(conn, `
		SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations';`, &sqlitex.ExecOptions{
		ResultFunc: func(_ *sqlite.Stmt) error {
			exists = true
			return nil
		},
	})
	return exists, err
}

func (db *SQLiteDatabase) createMigrationsTable(ctx context.Context) error {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version BIGINT PRIMARY KEY,
            dirty BOOLEAN
        );`, nil)
	return err
}

func (db *SQLiteDatabase) getCurrentVersion(ctx context.Context) (uint64, bool, error) {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return 0, false, err
	}
	defer db.pool.Put(conn)

	var version uint64
	var dirty bool
	found := false

	err = sqlitex.Execute(conn, `
        SELECT version, dirty FROM schema_migrations 
        ORDER BY version DESC LIMIT 1;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			version = uint64(stmt.ColumnInt64(0))
			dirty = stmt.ColumnBool(1)
			found = true
			return nil
		},
	})
	if err != nil {
		return 0, false, err
	}

	if !found {
		return 0, false, ErrNotFound
	}

	return version, dirty, nil
}

func (db *SQLiteDatabase) setVersion(ctx context.Context, version uint64, dirty bool) error {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `DELETE FROM schema_migrations;`, nil)
	if err != nil {
		return err
	}

	err = sqlitex.Execute(conn, `
        INSERT INTO schema_migrations (version, dirty) VALUES (?, ?);`, &sqlitex.ExecOptions{
		Args: []any{version, dirty},
	})
	return err
}

func (db *SQLiteDatabase) MigrateUp(steps int) error {
	if err := db.createMigrationsTable(context.Background()); err != nil {
		return err
	}

	currentVersion, dirty, err := db.getCurrentVersion(context.Background())
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if dirty {
		return errors.New("database is in dirty state")
	}

	entries, err := migrationsFS.ReadDir("migrations/sqlite")
	if err != nil {
		return err
	}
	files := GetOrderedMigrationFiles(entries, true)

	appliedCount := 0
	for _, file := range files {
		versionStr := strings.Split(file, "_")[0]
		version, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			return err
		}

		if version <= currentVersion {
			continue
		}

		if steps >= 0 && appliedCount >= steps {
			break
		}

		// Set dirty flag before applying migration.
		if err = db.setVersion(context.Background(), version, true); err != nil {
			return err
		}

		// Read and execute migration.
		content, err := migrationsFS.ReadFile(filepath.Join("migrations/sqlite", file))
		if err != nil {
			return err
		}

		conn, err := db.pool.Take(context.Background())
		if err != nil {
			return err
		}
		err = sqlitex.ExecuteScript(conn, string(content), nil)
		db.pool.Put(conn)
		if err != nil {
			return err
		}

		// Clear dirty flag after successful migration.
		if err = db.setVersion(context.Background(), version, false); err != nil {
			return err
		}

		appliedCount++
	}

	return nil
}

func (db *SQLiteDatabase) MigrateDown(steps int) error {
	if err := db.createMigrationsTable(context.Background()); err != nil {
		return err
	}

	currentVersion, dirty, err := db.getCurrentVersion(context.Background())
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if dirty {
		return errors.New("database is in dirty state")
	}

	entries, err := migrationsFS.ReadDir("migrations/sqlite")
	if err != nil {
		return err
	}
	files := GetOrderedMigrationFiles(entries, false)

	appliedCount := 0
	for _, file := range files {
		versionStr := strings.Split(file, "_")[0]
		version, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			return err
		}

		if version > currentVersion {
			continue
		}

		if steps >= 0 && appliedCount >= steps {
			break
		}

		// Set dirty flag before applying migration.
		if err = db.setVersion(context.Background(), version, true); err != nil {
			return err
		}

		// Read and execute migration.
		content, err := migrationsFS.ReadFile(filepath.Join("migrations/sqlite", file))
		if err != nil {
			return err
		}

		conn, err := db.pool.Take(context.Background())
		if err != nil {
			return err
		}
		err = sqlitex.ExecuteScript(conn, string(content), nil)
		db.pool.Put(conn)
		if err != nil {
			return err
		}

		// Set version to previous migration and clear dirty flag.
		prevVersion := uint64(0)
		if appliedCount < len(files)-1 {
			prevVersionStr := strings.Split(files[appliedCount+1], "_")[0]
			prevVersion, err = strconv.ParseUint(prevVersionStr, 10, 64)
			if err != nil {
				return err
			}
		}
		if err = db.setVersion(context.Background(), prevVersion, false); err != nil {
			return err
		}

		appliedCount++
	}

	return nil
}

func (db *SQLiteDatabase) GetAllTracks(ctx context.Context) ([]media.Track, error) {
	var tracks []media.Track

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return tracks, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT * FROM tracks;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			track := media.Track{}

			track.ID = stmt.ColumnText(0)
			track.UserID = stmt.ColumnText(1)
			track.ISRC = stmt.ColumnText(2)
			track.Title = stmt.ColumnText(3)
			if err = json.Unmarshal([]byte(stmt.ColumnText(4)), &track.ArtistIDs); err != nil {
				return fmt.Errorf("failed to parse artist_ids: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(5)), &track.AlbumIDs); err != nil {
				return fmt.Errorf("failed to parse album_ids: %w", err)
			}
			track.PrimaryAlbumID = stmt.ColumnText(6)
			track.TrackNumber = stmt.ColumnInt(7)
			track.Duration = stmt.ColumnInt(8)
			track.Description = stmt.ColumnText(9)
			track.ReleaseDate = stmt.ColumnText(10)
			if err = json.Unmarshal([]byte(stmt.ColumnText(11)), &track.Lyrics); err != nil {
				return fmt.Errorf("failed to parse lyrics: %w", err)
			}
			track.ListenCount = stmt.ColumnInt(12)
			track.FavoriteCount = stmt.ColumnInt(13)
			track.AdditionDate = stmt.ColumnInt64(14)
			if err = json.Unmarshal([]byte(stmt.ColumnText(15)), &track.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(16)), &track.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(17)), &track.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(18)), &track.LinkedItemIDs); err != nil {
				return fmt.Errorf("failed to parse linked_item_ids: %w", err)
			}
			track.ContentSource = stmt.ColumnText(19)
			track.MetadataSource = stmt.ColumnText(20)
			if err = json.Unmarshal([]byte(stmt.ColumnText(21)), &track.LyricSources); err != nil {
				return fmt.Errorf("failed to parse lyric_sources: %w", err)
			}

			tracks = append(tracks, track)

			return nil
		},
	})

	return tracks, err
}

func (db *SQLiteDatabase) GetTracks(ctx context.Context, userID string) ([]media.Track, error) {
	var tracks []media.Track

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return tracks, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT * FROM tracks WHERE user_id = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			track := media.Track{}

			track.ID = stmt.ColumnText(0)
			track.UserID = stmt.ColumnText(1)
			track.ISRC = stmt.ColumnText(2)
			track.Title = stmt.ColumnText(3)
			if err = json.Unmarshal([]byte(stmt.ColumnText(4)), &track.ArtistIDs); err != nil {
				return fmt.Errorf("failed to parse artist_ids: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(5)), &track.AlbumIDs); err != nil {
				return fmt.Errorf("failed to parse album_ids: %w", err)
			}
			track.PrimaryAlbumID = stmt.ColumnText(6)
			track.TrackNumber = stmt.ColumnInt(7)
			track.Duration = stmt.ColumnInt(8)
			track.Description = stmt.ColumnText(9)
			track.ReleaseDate = stmt.ColumnText(10)
			if err = json.Unmarshal([]byte(stmt.ColumnText(11)), &track.Lyrics); err != nil {
				return fmt.Errorf("failed to parse lyrics: %w", err)
			}
			track.ListenCount = stmt.ColumnInt(12)
			track.FavoriteCount = stmt.ColumnInt(13)
			track.AdditionDate = stmt.ColumnInt64(14)
			if err = json.Unmarshal([]byte(stmt.ColumnText(15)), &track.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(16)), &track.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(17)), &track.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(18)), &track.LinkedItemIDs); err != nil {
				return fmt.Errorf("failed to parse linked_item_ids: %w", err)
			}
			track.ContentSource = stmt.ColumnText(19)
			track.MetadataSource = stmt.ColumnText(20)
			if err = json.Unmarshal([]byte(stmt.ColumnText(21)), &track.LyricSources); err != nil {
				return fmt.Errorf("failed to parse lyric_sources: %w", err)
			}

			tracks = append(tracks, track)

			return nil
		},
		Args: []any{userID},
	})

	return tracks, err
}

func (db *SQLiteDatabase) GetTrack(ctx context.Context, id string) (media.Track, error) {
	track := media.Track{}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return track, err
	}
	defer db.pool.Put(conn)

	scanned := false
	err = sqlitex.Execute(conn, `SELECT * FROM tracks WHERE id = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			if scanned {
				return ErrTooMany
			}
			scanned = true

			track.ID = stmt.ColumnText(0)
			track.UserID = stmt.ColumnText(1)
			track.ISRC = stmt.ColumnText(2)
			track.Title = stmt.ColumnText(3)
			if err = json.Unmarshal([]byte(stmt.ColumnText(4)), &track.ArtistIDs); err != nil {
				return fmt.Errorf("failed to parse artist_ids: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(5)), &track.AlbumIDs); err != nil {
				return fmt.Errorf("failed to parse album_ids: %w", err)
			}
			track.PrimaryAlbumID = stmt.ColumnText(6)
			track.TrackNumber = stmt.ColumnInt(7)
			track.Duration = stmt.ColumnInt(8)
			track.Description = stmt.ColumnText(9)
			track.ReleaseDate = stmt.ColumnText(10)
			if err = json.Unmarshal([]byte(stmt.ColumnText(11)), &track.Lyrics); err != nil {
				return fmt.Errorf("failed to parse lyrics: %w", err)
			}
			track.ListenCount = stmt.ColumnInt(12)
			track.FavoriteCount = stmt.ColumnInt(13)
			track.AdditionDate = stmt.ColumnInt64(14)
			if err = json.Unmarshal([]byte(stmt.ColumnText(15)), &track.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(16)), &track.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(17)), &track.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(18)), &track.LinkedItemIDs); err != nil {
				return fmt.Errorf("failed to parse linked_item_ids: %w", err)
			}
			track.ContentSource = stmt.ColumnText(19)
			track.MetadataSource = stmt.ColumnText(20)
			if err = json.Unmarshal([]byte(stmt.ColumnText(21)), &track.LyricSources); err != nil {
				return fmt.Errorf("failed to parse lyric_sources: %w", err)
			}

			return nil
		},
		Args: []any{id},
	})
	if err != nil {
		return track, err
	}
	if !scanned {
		return track, ErrNotFound
	}

	return track, nil
}

func (db *SQLiteDatabase) AddTrack(ctx context.Context, track media.Track) error {
	// Convert JSON fields to strings.
	artistIDs, err := json.Marshal(track.ArtistIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal artist_ids: %w", err)
	}
	albumIDs, err := json.Marshal(track.AlbumIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal album_ids: %w", err)
	}
	lyrics, err := json.Marshal(track.Lyrics)
	if err != nil {
		return fmt.Errorf("failed to marshal lyrics: %w", err)
	}
	tags, err := json.Marshal(track.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	additionalMeta, err := json.Marshal(track.AdditionalMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal additional_meta: %w", err)
	}
	permissions, err := json.Marshal(track.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	linkedItemIDs, err := json.Marshal(track.LinkedItemIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal linked_item_ids: %w", err)
	}
	lyricSources, err := json.Marshal(track.LyricSources)
	if err != nil {
		return fmt.Errorf("failed to marshal lyric_sources: %w", err)
	}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
	  INSERT INTO tracks (
	    id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
	  ) VALUES (
	    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	  );`, &sqlitex.ExecOptions{
		Args: []any{
			track.ID, track.UserID, track.ISRC, track.Title, string(artistIDs), string(albumIDs),
			track.PrimaryAlbumID, track.TrackNumber, track.Duration, track.Description,
			track.ReleaseDate, string(lyrics), track.ListenCount, track.FavoriteCount,
			track.AdditionDate, string(tags), string(additionalMeta), string(permissions),
			string(linkedItemIDs), track.ContentSource, track.MetadataSource, string(lyricSources),
		},
	})

	return err
}

func (db *SQLiteDatabase) UpdateTrack(ctx context.Context, track media.Track) error {
	// Convert JSON fields to strings.
	artistIDs, err := json.Marshal(track.ArtistIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal artist_ids: %w", err)
	}
	albumIDs, err := json.Marshal(track.AlbumIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal album_ids: %w", err)
	}
	lyrics, err := json.Marshal(track.Lyrics)
	if err != nil {
		return fmt.Errorf("failed to marshal lyrics: %w", err)
	}
	tags, err := json.Marshal(track.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	additionalMeta, err := json.Marshal(track.AdditionalMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal additional_meta: %w", err)
	}
	permissions, err := json.Marshal(track.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	linkedItemIDs, err := json.Marshal(track.LinkedItemIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal linked_item_ids: %w", err)
	}
	lyricSources, err := json.Marshal(track.LyricSources)
	if err != nil {
		return fmt.Errorf("failed to marshal lyric_sources: %w", err)
	}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
	  UPDATE tracks
	  SET user_id=?, isrc=?, title=?, artist_ids=?, album_ids=?, primary_album_id=?, 
	      track_number=?, duration=?, description=?, release_date=?, lyrics=?, 
	      listen_count=?, favorite_count=?, addition_date=?, tags=?, additional_meta=?, 
	      permissions=?, linked_item_ids=?, content_source=?, metadata_source=?, 
	      lyric_sources=?
	  WHERE id=?;`, &sqlitex.ExecOptions{
		Args: []any{
			track.UserID, track.ISRC, track.Title, string(artistIDs), string(albumIDs),
			track.PrimaryAlbumID, track.TrackNumber, track.Duration, track.Description,
			track.ReleaseDate, string(lyrics), track.ListenCount, track.FavoriteCount,
			track.AdditionDate, string(tags), string(additionalMeta), string(permissions),
			string(linkedItemIDs), track.ContentSource, track.MetadataSource,
			string(lyricSources), track.ID,
		},
	})

	return err
}

func (db *SQLiteDatabase) DeleteTrack(ctx context.Context, id string) error {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `DELETE FROM tracks WHERE id = ?;`, &sqlitex.ExecOptions{
		Args: []any{id},
	})
	return err
}

func (db *SQLiteDatabase) GetAllAlbums(ctx context.Context) ([]media.Album, error) {
	var albums []media.Album

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return albums, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT * FROM albums;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			album := media.Album{}

			album.ID = stmt.ColumnText(0)
			album.UserID = stmt.ColumnText(1)
			album.UPC = stmt.ColumnText(2)
			album.EAN = stmt.ColumnText(3)
			album.Title = stmt.ColumnText(4)
			if err = json.Unmarshal([]byte(stmt.ColumnText(5)), &album.ArtistIDs); err != nil {
				return fmt.Errorf("failed to parse artist_ids: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(6)), &album.TrackIDs); err != nil {
				return fmt.Errorf("failed to parse track_ids: %w", err)
			}
			album.Description = stmt.ColumnText(7)
			album.ReleaseDate = stmt.ColumnText(8)
			album.ListenCount = stmt.ColumnInt(9)
			album.FavoriteCount = stmt.ColumnInt(10)
			album.AdditionDate = stmt.ColumnInt64(11)
			if err = json.Unmarshal([]byte(stmt.ColumnText(12)), &album.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(13)), &album.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(14)), &album.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(15)), &album.LinkedItemIDs); err != nil {
				return fmt.Errorf("failed to parse linked_item_ids: %w", err)
			}
			album.MetadataSource = stmt.ColumnText(16)

			albums = append(albums, album)

			return nil
		},
	})

	return albums, err
}

func (db *SQLiteDatabase) GetAlbums(ctx context.Context, userID string) ([]media.Album, error) {
	var albums []media.Album

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return albums, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT * FROM albums WHERE user_id = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			album := media.Album{}

			album.ID = stmt.ColumnText(0)
			album.UserID = stmt.ColumnText(1)
			album.UPC = stmt.ColumnText(2)
			album.EAN = stmt.ColumnText(3)
			album.Title = stmt.ColumnText(4)
			if err = json.Unmarshal([]byte(stmt.ColumnText(5)), &album.ArtistIDs); err != nil {
				return fmt.Errorf("failed to parse artist_ids: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(6)), &album.TrackIDs); err != nil {
				return fmt.Errorf("failed to parse track_ids: %w", err)
			}
			album.Description = stmt.ColumnText(7)
			album.ReleaseDate = stmt.ColumnText(8)
			album.ListenCount = stmt.ColumnInt(9)
			album.FavoriteCount = stmt.ColumnInt(10)
			album.AdditionDate = stmt.ColumnInt64(11)
			if err = json.Unmarshal([]byte(stmt.ColumnText(12)), &album.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(13)), &album.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(14)), &album.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(15)), &album.LinkedItemIDs); err != nil {
				return fmt.Errorf("failed to parse linked_item_ids: %w", err)
			}
			album.MetadataSource = stmt.ColumnText(16)

			albums = append(albums, album)

			return nil
		},
		Args: []any{userID},
	})

	return albums, err
}

func (db *SQLiteDatabase) GetAlbum(ctx context.Context, id string) (media.Album, error) {
	album := media.Album{}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return album, err
	}
	defer db.pool.Put(conn)

	scanned := false
	err = sqlitex.Execute(conn, `SELECT * FROM albums WHERE id = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			if scanned {
				return ErrTooMany
			}
			scanned = true

			album.ID = stmt.ColumnText(0)
			album.UserID = stmt.ColumnText(1)
			album.UPC = stmt.ColumnText(2)
			album.EAN = stmt.ColumnText(3)
			album.Title = stmt.ColumnText(4)
			if err = json.Unmarshal([]byte(stmt.ColumnText(5)), &album.ArtistIDs); err != nil {
				return fmt.Errorf("failed to parse artist_ids: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(6)), &album.TrackIDs); err != nil {
				return fmt.Errorf("failed to parse track_ids: %w", err)
			}
			album.Description = stmt.ColumnText(7)
			album.ReleaseDate = stmt.ColumnText(8)
			album.ListenCount = stmt.ColumnInt(9)
			album.FavoriteCount = stmt.ColumnInt(10)
			album.AdditionDate = stmt.ColumnInt64(11)
			if err = json.Unmarshal([]byte(stmt.ColumnText(12)), &album.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(13)), &album.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(14)), &album.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			if err = json.Unmarshal([]byte(stmt.ColumnText(15)), &album.LinkedItemIDs); err != nil {
				return fmt.Errorf("failed to parse linked_item_ids: %w", err)
			}
			album.MetadataSource = stmt.ColumnText(16)

			return nil
		},
		Args: []any{id},
	})
	if err != nil {
		return album, err
	}
	if !scanned {
		return album, ErrNotFound
	}

	return album, nil
}

func (db *SQLiteDatabase) AddAlbum(ctx context.Context, album media.Album) error {
	// Convert JSON fields to strings.
	artistIDs, err := json.Marshal(album.ArtistIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal artist_ids: %w", err)
	}
	trackIDs, err := json.Marshal(album.TrackIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal track_ids: %w", err)
	}
	tags, err := json.Marshal(album.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	additionalMeta, err := json.Marshal(album.AdditionalMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal additional_meta: %w", err)
	}
	permissions, err := json.Marshal(album.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	linkedItemIDs, err := json.Marshal(album.LinkedItemIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal linked_item_ids: %w", err)
	}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
	  INSERT INTO albums (
	    id, user_id, upc, ean, title, artist_ids, track_ids, description, release_date, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, metadata_source
	  ) VALUES (
	    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	  );`, &sqlitex.ExecOptions{
		Args: []any{
			album.ID, album.UserID, album.UPC, album.EAN, album.Title, string(artistIDs), string(trackIDs),
			album.Description, album.ReleaseDate, album.ListenCount, album.FavoriteCount,
			album.AdditionDate, string(tags), string(additionalMeta), string(permissions),
			string(linkedItemIDs), album.MetadataSource,
		},
	})

	return err
}

func (db *SQLiteDatabase) UpdateAlbum(ctx context.Context, album media.Album) error {
	// Convert JSON fields to strings.
	artistIDs, err := json.Marshal(album.ArtistIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal artist_ids: %w", err)
	}
	trackIDs, err := json.Marshal(album.TrackIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal track_ids: %w", err)
	}
	tags, err := json.Marshal(album.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	additionalMeta, err := json.Marshal(album.AdditionalMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal additional_meta: %w", err)
	}
	permissions, err := json.Marshal(album.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	linkedItemIDs, err := json.Marshal(album.LinkedItemIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal linked_item_ids: %w", err)
	}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
	  UPDATE albums
	  SET user_id=?, upc=?, ean=?, title=?, artist_ids=?, track_ids=?, description=?, 
	      release_date=?, listen_count=?, favorite_count=?, addition_date=?, tags=?, 
	      additional_meta=?, permissions=?, linked_item_ids=?, metadata_source=?
	  WHERE id=?;`, &sqlitex.ExecOptions{
		Args: []any{
			album.UserID, album.UPC, album.EAN, album.Title, string(artistIDs), string(trackIDs),
			album.Description, album.ReleaseDate, album.ListenCount, album.FavoriteCount,
			album.AdditionDate, string(tags), string(additionalMeta), string(permissions),
			string(linkedItemIDs), album.MetadataSource, album.ID,
		},
	})

	return err
}

func (db *SQLiteDatabase) DeleteAlbum(ctx context.Context, id string) error {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `DELETE FROM albums WHERE id = ?;`, &sqlitex.ExecOptions{
		Args: []any{id},
	})
	return err
}

func (db *SQLiteDatabase) GetAllVideos(ctx context.Context) ([]media.Video, error) {
	var videos []media.Video

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return videos, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT * FROM videos;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			video := media.Video{}

			video.ID = stmt.ColumnText(0)
			video.UserID = stmt.ColumnText(1)
			video.Title = stmt.ColumnText(2)
			if err := json.Unmarshal([]byte(stmt.ColumnText(3)), &video.ArtistIDs); err != nil {
				return fmt.Errorf("failed to parse artist_ids: %w", err)
			}
			video.Duration = stmt.ColumnInt(4)
			video.Description = stmt.ColumnText(5)
			video.ReleaseDate = stmt.ColumnText(6)
			if err := json.Unmarshal([]byte(stmt.ColumnText(7)), &video.Subtitles); err != nil {
				return fmt.Errorf("failed to parse subtitles: %w", err)
			}
			video.WatchCount = stmt.ColumnInt(8)
			video.FavoriteCount = stmt.ColumnInt(9)
			video.AdditionDate = stmt.ColumnInt64(10)
			if err := json.Unmarshal([]byte(stmt.ColumnText(11)), &video.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(12)), &video.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(13)), &video.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(14)), &video.LinkedItemIDs); err != nil {
				return fmt.Errorf("failed to parse linked_item_ids: %w", err)
			}
			video.ContentSource = stmt.ColumnText(15)
			video.MetadataSource = stmt.ColumnText(16)
			if err := json.Unmarshal([]byte(stmt.ColumnText(17)), &video.LyricSources); err != nil {
				return fmt.Errorf("failed to parse lyric_sources: %w", err)
			}

			videos = append(videos, video)

			return nil
		},
	})

	return videos, err
}

func (db *SQLiteDatabase) GetVideos(ctx context.Context, userID string) ([]media.Video, error) {
	var videos []media.Video

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return videos, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT * FROM videos WHERE user_id = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			video := media.Video{}

			video.ID = stmt.ColumnText(0)
			video.UserID = stmt.ColumnText(1)
			video.Title = stmt.ColumnText(2)
			if err := json.Unmarshal([]byte(stmt.ColumnText(3)), &video.ArtistIDs); err != nil {
				return fmt.Errorf("failed to parse artist_ids: %w", err)
			}
			video.Duration = stmt.ColumnInt(4)
			video.Description = stmt.ColumnText(5)
			video.ReleaseDate = stmt.ColumnText(6)
			if err := json.Unmarshal([]byte(stmt.ColumnText(7)), &video.Subtitles); err != nil {
				return fmt.Errorf("failed to parse subtitles: %w", err)
			}
			video.WatchCount = stmt.ColumnInt(8)
			video.FavoriteCount = stmt.ColumnInt(9)
			video.AdditionDate = stmt.ColumnInt64(10)
			if err := json.Unmarshal([]byte(stmt.ColumnText(11)), &video.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(12)), &video.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(13)), &video.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(14)), &video.LinkedItemIDs); err != nil {
				return fmt.Errorf("failed to parse linked_item_ids: %w", err)
			}
			video.ContentSource = stmt.ColumnText(15)
			video.MetadataSource = stmt.ColumnText(16)
			if err := json.Unmarshal([]byte(stmt.ColumnText(17)), &video.LyricSources); err != nil {
				return fmt.Errorf("failed to parse lyric_sources: %w", err)
			}

			videos = append(videos, video)

			return nil
		},
		Args: []any{userID},
	})

	return videos, err
}

func (db *SQLiteDatabase) GetVideo(ctx context.Context, id string) (media.Video, error) {
	video := media.Video{}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return video, err
	}
	defer db.pool.Put(conn)

	scanned := false
	err = sqlitex.Execute(conn, `SELECT * FROM videos WHERE id = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			if scanned {
				return ErrTooMany
			}
			scanned = true

			video.ID = stmt.ColumnText(0)
			video.UserID = stmt.ColumnText(1)
			video.Title = stmt.ColumnText(2)
			if err := json.Unmarshal([]byte(stmt.ColumnText(3)), &video.ArtistIDs); err != nil {
				return fmt.Errorf("failed to parse artist_ids: %w", err)
			}
			video.Duration = stmt.ColumnInt(4)
			video.Description = stmt.ColumnText(5)
			video.ReleaseDate = stmt.ColumnText(6)
			if err := json.Unmarshal([]byte(stmt.ColumnText(7)), &video.Subtitles); err != nil {
				return fmt.Errorf("failed to parse subtitles: %w", err)
			}
			video.WatchCount = stmt.ColumnInt(8)
			video.FavoriteCount = stmt.ColumnInt(9)
			video.AdditionDate = stmt.ColumnInt64(10)
			if err := json.Unmarshal([]byte(stmt.ColumnText(11)), &video.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(12)), &video.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(13)), &video.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(14)), &video.LinkedItemIDs); err != nil {
				return fmt.Errorf("failed to parse linked_item_ids: %w", err)
			}
			video.ContentSource = stmt.ColumnText(15)
			video.MetadataSource = stmt.ColumnText(16)
			if err := json.Unmarshal([]byte(stmt.ColumnText(17)), &video.LyricSources); err != nil {
				return fmt.Errorf("failed to parse lyric_sources: %w", err)
			}

			return nil
		},
		Args: []any{id},
	})
	if err != nil {
		return video, err
	}
	if !scanned {
		return video, ErrNotFound
	}

	return video, nil
}

func (db *SQLiteDatabase) AddVideo(ctx context.Context, video media.Video) error {
	// Convert JSON fields to strings.
	artistIDs, err := json.Marshal(video.ArtistIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal artist_ids: %w", err)
	}
	subtitles, err := json.Marshal(video.Subtitles)
	if err != nil {
		return fmt.Errorf("failed to marshal subtitles: %w", err)
	}
	tags, err := json.Marshal(video.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	additionalMeta, err := json.Marshal(video.AdditionalMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal additional_meta: %w", err)
	}
	permissions, err := json.Marshal(video.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	linkedItemIDs, err := json.Marshal(video.LinkedItemIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal linked_item_ids: %w", err)
	}
	lyricSources, err := json.Marshal(video.LyricSources)
	if err != nil {
		return fmt.Errorf("failed to marshal lyric_sources: %w", err)
	}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
	  INSERT INTO videos (
	    id, user_id, title, artist_ids, duration, description, release_date, subtitles, watch_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
	  ) VALUES (
	   	?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	  );`, &sqlitex.ExecOptions{
		Args: []any{
			video.ID, video.UserID, video.Title, string(artistIDs), video.Duration,
			video.Description, video.ReleaseDate, string(subtitles), video.WatchCount,
			video.FavoriteCount, video.AdditionDate, string(tags), string(additionalMeta),
			string(permissions), string(linkedItemIDs), video.ContentSource,
			video.MetadataSource, string(lyricSources),
		},
	})

	return err
}

func (db *SQLiteDatabase) UpdateVideo(ctx context.Context, video media.Video) error {
	// Convert JSON fields to strings.
	artistIDs, err := json.Marshal(video.ArtistIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal artist_ids: %w", err)
	}
	subtitles, err := json.Marshal(video.Subtitles)
	if err != nil {
		return fmt.Errorf("failed to marshal subtitles: %w", err)
	}
	tags, err := json.Marshal(video.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	additionalMeta, err := json.Marshal(video.AdditionalMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal additional_meta: %w", err)
	}
	permissions, err := json.Marshal(video.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	linkedItemIDs, err := json.Marshal(video.LinkedItemIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal linked_item_ids: %w", err)
	}
	lyricSources, err := json.Marshal(video.LyricSources)
	if err != nil {
		return fmt.Errorf("failed to marshal lyric_sources: %w", err)
	}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
	  UPDATE videos
	  SET user_id=?, title=?, artist_ids=?, duration=?, description=?, release_date=?,
	      subtitles=?, watch_count=?, favorite_count=?, addition_date=?, tags=?,
	      additional_meta=?, permissions=?, linked_item_ids=?, content_source=?,
	      metadata_source=?, lyric_sources=?
	  WHERE id=?;`, &sqlitex.ExecOptions{
		Args: []any{
			video.UserID, video.Title, string(artistIDs), video.Duration, video.Description,
			video.ReleaseDate, string(subtitles), video.WatchCount, video.FavoriteCount,
			video.AdditionDate, string(tags), string(additionalMeta), string(permissions),
			string(linkedItemIDs), video.ContentSource, video.MetadataSource,
			string(lyricSources), video.ID,
		},
	})

	return err
}

func (db *SQLiteDatabase) DeleteVideo(ctx context.Context, id string) error {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `DELETE FROM videos WHERE id = ?;`, &sqlitex.ExecOptions{
		Args: []any{id},
	})
	return err
}

func (db *SQLiteDatabase) GetAllArtists(ctx context.Context) ([]media.Artist, error) {
	var artists []media.Artist

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return artists, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT * FROM artists;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			artist := media.Artist{}

			artist.ID = stmt.ColumnText(0)
			artist.UserID = stmt.ColumnText(1)
			artist.Name = stmt.ColumnText(2)
			if err := json.Unmarshal([]byte(stmt.ColumnText(3)), &artist.AlbumIDs); err != nil {
				return fmt.Errorf("failed to parse album_ids: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(4)), &artist.TrackIDs); err != nil {
				return fmt.Errorf("failed to parse track_ids: %w", err)
			}
			artist.Description = stmt.ColumnText(5)
			artist.CreationDate = stmt.ColumnText(6)
			artist.ListenCount = stmt.ColumnInt(7)
			artist.FavoriteCount = stmt.ColumnInt(8)
			artist.AdditionDate = stmt.ColumnInt64(9)
			if err := json.Unmarshal([]byte(stmt.ColumnText(10)), &artist.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(11)), &artist.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(12)), &artist.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(13)), &artist.LinkedItemIDs); err != nil {
				return fmt.Errorf("failed to parse linked_item_ids: %w", err)
			}
			artist.MetadataSource = stmt.ColumnText(14)

			artists = append(artists, artist)

			return nil
		},
	})

	return artists, err
}

func (db *SQLiteDatabase) GetArtists(ctx context.Context, userID string) ([]media.Artist, error) {
	var artists []media.Artist

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return artists, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT * FROM artists WHERE user_id = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			artist := media.Artist{}

			artist.ID = stmt.ColumnText(0)
			artist.UserID = stmt.ColumnText(1)
			artist.Name = stmt.ColumnText(2)
			if err := json.Unmarshal([]byte(stmt.ColumnText(3)), &artist.AlbumIDs); err != nil {
				return fmt.Errorf("failed to parse album_ids: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(4)), &artist.TrackIDs); err != nil {
				return fmt.Errorf("failed to parse track_ids: %w", err)
			}
			artist.Description = stmt.ColumnText(5)
			artist.CreationDate = stmt.ColumnText(6)
			artist.ListenCount = stmt.ColumnInt(7)
			artist.FavoriteCount = stmt.ColumnInt(8)
			artist.AdditionDate = stmt.ColumnInt64(9)
			if err := json.Unmarshal([]byte(stmt.ColumnText(10)), &artist.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(11)), &artist.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(12)), &artist.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(13)), &artist.LinkedItemIDs); err != nil {
				return fmt.Errorf("failed to parse linked_item_ids: %w", err)
			}
			artist.MetadataSource = stmt.ColumnText(14)

			artists = append(artists, artist)

			return nil
		},
		Args: []any{userID},
	})

	return artists, err
}

func (db *SQLiteDatabase) GetArtist(ctx context.Context, id string) (media.Artist, error) {
	artist := media.Artist{}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return artist, err
	}
	defer db.pool.Put(conn)

	scanned := false
	err = sqlitex.Execute(conn, `SELECT * FROM artists WHERE id = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			if scanned {
				return ErrTooMany
			}
			scanned = true

			artist.ID = stmt.ColumnText(0)
			artist.UserID = stmt.ColumnText(1)
			artist.Name = stmt.ColumnText(2)
			if err := json.Unmarshal([]byte(stmt.ColumnText(3)), &artist.AlbumIDs); err != nil {
				return fmt.Errorf("failed to parse album_ids: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(4)), &artist.TrackIDs); err != nil {
				return fmt.Errorf("failed to parse track_ids: %w", err)
			}
			artist.Description = stmt.ColumnText(5)
			artist.CreationDate = stmt.ColumnText(6)
			artist.ListenCount = stmt.ColumnInt(7)
			artist.FavoriteCount = stmt.ColumnInt(8)
			artist.AdditionDate = stmt.ColumnInt64(9)
			if err := json.Unmarshal([]byte(stmt.ColumnText(10)), &artist.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(11)), &artist.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(12)), &artist.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(13)), &artist.LinkedItemIDs); err != nil {
				return fmt.Errorf("failed to parse linked_item_ids: %w", err)
			}
			artist.MetadataSource = stmt.ColumnText(14)

			return nil
		},
		Args: []any{id},
	})
	if err != nil {
		return artist, err
	}
	if !scanned {
		return artist, ErrNotFound
	}

	return artist, nil
}

func (db *SQLiteDatabase) AddArtist(ctx context.Context, artist media.Artist) error {
	// Convert JSON fields to strings.
	albumIDs, err := json.Marshal(artist.AlbumIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal album_ids: %w", err)
	}
	trackIDs, err := json.Marshal(artist.TrackIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal track_ids: %w", err)
	}
	tags, err := json.Marshal(artist.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	additionalMeta, err := json.Marshal(artist.AdditionalMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal additional_meta: %w", err)
	}
	permissions, err := json.Marshal(artist.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	linkedItemIDs, err := json.Marshal(artist.LinkedItemIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal linked_item_ids: %w", err)
	}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
	  INSERT INTO artists (
	    id, user_id, name, album_ids, track_ids, description, creation_date, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, metadata_source
	  ) VALUES (
	   	?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	  );`, &sqlitex.ExecOptions{
		Args: []any{
			artist.ID, artist.UserID, artist.Name, string(albumIDs), string(trackIDs),
			artist.Description, artist.CreationDate, artist.ListenCount, artist.FavoriteCount,
			artist.AdditionDate, string(tags), string(additionalMeta), string(permissions),
			string(linkedItemIDs), artist.MetadataSource,
		},
	})

	return err
}

func (db *SQLiteDatabase) UpdateArtist(ctx context.Context, artist media.Artist) error {
	// Convert JSON fields to strings.
	albumIDs, err := json.Marshal(artist.AlbumIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal album_ids: %w", err)
	}
	trackIDs, err := json.Marshal(artist.TrackIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal track_ids: %w", err)
	}
	tags, err := json.Marshal(artist.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	additionalMeta, err := json.Marshal(artist.AdditionalMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal additional_meta: %w", err)
	}
	permissions, err := json.Marshal(artist.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	linkedItemIDs, err := json.Marshal(artist.LinkedItemIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal linked_item_ids: %w", err)
	}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
	  UPDATE artists
	  SET user_id=?, name=?, album_ids=?, track_ids=?, description=?, creation_date=?,
	      listen_count=?, favorite_count=?, addition_date=?, tags=?, additional_meta=?,
		  permissions=?, linked_item_ids=?, metadata_source=?
	  WHERE id=?;`, &sqlitex.ExecOptions{
		Args: []any{
			artist.UserID, artist.Name, string(albumIDs), string(trackIDs),
			artist.Description, artist.CreationDate, artist.ListenCount, artist.FavoriteCount,
			artist.AdditionDate, string(tags), string(additionalMeta), string(permissions),
			string(linkedItemIDs), artist.MetadataSource, artist.ID,
		},
	})

	return err
}

func (db *SQLiteDatabase) DeleteArtist(ctx context.Context, id string) error {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `DELETE FROM artists WHERE id = ?;`, &sqlitex.ExecOptions{
		Args: []any{id},
	})
	return err
}

func (db *SQLiteDatabase) GetAllPlaylists(ctx context.Context) ([]media.Playlist, error) {
	var playlists []media.Playlist

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return playlists, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT * FROM playlists;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			playlist := media.Playlist{}

			playlist.ID = stmt.ColumnText(0)
			playlist.UserID = stmt.ColumnText(1)
			playlist.Title = stmt.ColumnText(2)
			if err := json.Unmarshal([]byte(stmt.ColumnText(3)), &playlist.TrackIDs); err != nil {
				return fmt.Errorf("failed to parse track_ids: %w", err)
			}
			playlist.ListenCount = stmt.ColumnInt(4)
			playlist.FavoriteCount = stmt.ColumnInt(5)
			playlist.Description = stmt.ColumnText(6)
			playlist.CreationDate = stmt.ColumnText(7)
			playlist.AdditionDate = stmt.ColumnInt64(8)
			if err := json.Unmarshal([]byte(stmt.ColumnText(9)), &playlist.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(10)), &playlist.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(11)), &playlist.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			playlist.MetadataSource = stmt.ColumnText(12)

			playlists = append(playlists, playlist)

			return nil
		},
	})

	return playlists, err
}

func (db *SQLiteDatabase) GetPlaylists(ctx context.Context, userID string) ([]media.Playlist, error) {
	var playlists []media.Playlist

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return playlists, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT * FROM playlists WHERE user_id = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			playlist := media.Playlist{}

			playlist.ID = stmt.ColumnText(0)
			playlist.UserID = stmt.ColumnText(1)
			playlist.Title = stmt.ColumnText(2)
			if err := json.Unmarshal([]byte(stmt.ColumnText(3)), &playlist.TrackIDs); err != nil {
				return fmt.Errorf("failed to parse track_ids: %w", err)
			}
			playlist.ListenCount = stmt.ColumnInt(4)
			playlist.FavoriteCount = stmt.ColumnInt(5)
			playlist.Description = stmt.ColumnText(6)
			playlist.CreationDate = stmt.ColumnText(7)
			playlist.AdditionDate = stmt.ColumnInt64(8)
			if err := json.Unmarshal([]byte(stmt.ColumnText(9)), &playlist.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(10)), &playlist.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(11)), &playlist.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			playlist.MetadataSource = stmt.ColumnText(12)

			playlists = append(playlists, playlist)

			return nil
		},
		Args: []any{userID},
	})

	return playlists, err
}

func (db *SQLiteDatabase) GetPlaylist(ctx context.Context, id string) (media.Playlist, error) {
	playlist := media.Playlist{}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return playlist, err
	}
	defer db.pool.Put(conn)

	scanned := false
	err = sqlitex.Execute(conn, `SELECT * FROM playlists WHERE id = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			if scanned {
				return ErrTooMany
			}
			scanned = true

			playlist.ID = stmt.ColumnText(0)
			playlist.UserID = stmt.ColumnText(1)
			playlist.Title = stmt.ColumnText(2)
			if err := json.Unmarshal([]byte(stmt.ColumnText(3)), &playlist.TrackIDs); err != nil {
				return fmt.Errorf("failed to parse track_ids: %w", err)
			}
			playlist.ListenCount = stmt.ColumnInt(4)
			playlist.FavoriteCount = stmt.ColumnInt(5)
			playlist.Description = stmt.ColumnText(6)
			playlist.CreationDate = stmt.ColumnText(7)
			playlist.AdditionDate = stmt.ColumnInt64(8)
			if err := json.Unmarshal([]byte(stmt.ColumnText(9)), &playlist.Tags); err != nil {
				return fmt.Errorf("failed to parse tags: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(10)), &playlist.AdditionalMeta); err != nil {
				return fmt.Errorf("failed to parse additional_meta: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(11)), &playlist.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			playlist.MetadataSource = stmt.ColumnText(12)

			return nil
		},
		Args: []any{id},
	})
	if err != nil {
		return playlist, err
	}
	if !scanned {
		return playlist, ErrNotFound
	}

	return playlist, nil
}

func (db *SQLiteDatabase) AddPlaylist(ctx context.Context, playlist media.Playlist) error {
	// Convert JSON fields to strings.
	trackIDs, err := json.Marshal(playlist.TrackIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal track_ids: %w", err)
	}
	tags, err := json.Marshal(playlist.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	additionalMeta, err := json.Marshal(playlist.AdditionalMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal additional_meta: %w", err)
	}
	permissions, err := json.Marshal(playlist.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
	  INSERT INTO playlists (
	    id, user_id, title, track_ids, listen_count, favorite_count, description, 
	    creation_date, addition_date, tags, additional_meta, permissions, metadata_source
	  ) VALUES (
	   	?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	  );`, &sqlitex.ExecOptions{
		Args: []any{
			playlist.ID, playlist.UserID, playlist.Title, string(trackIDs),
			playlist.ListenCount, playlist.FavoriteCount, playlist.Description,
			playlist.CreationDate, playlist.AdditionDate, string(tags),
			string(additionalMeta), string(permissions), playlist.MetadataSource,
		},
	})

	return err
}

func (db *SQLiteDatabase) UpdatePlaylist(ctx context.Context, playlist media.Playlist) error {
	// Convert JSON fields to strings.
	trackIDs, err := json.Marshal(playlist.TrackIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal track_ids: %w", err)
	}
	tags, err := json.Marshal(playlist.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	additionalMeta, err := json.Marshal(playlist.AdditionalMeta)
	if err != nil {
		return fmt.Errorf("failed to marshal additional_meta: %w", err)
	}
	permissions, err := json.Marshal(playlist.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
	  UPDATE playlists
	  SET user_id=?, title=?, track_ids=?, listen_count=?, favorite_count=?, description=?, 
	      creation_date=?, addition_date=?, tags=?, additional_meta=?, permissions=?, 
	      metadata_source=?
	  WHERE id=?;`, &sqlitex.ExecOptions{
		Args: []any{
			playlist.UserID, playlist.Title, string(trackIDs),
			playlist.ListenCount, playlist.FavoriteCount, playlist.Description,
			playlist.CreationDate, playlist.AdditionDate, string(tags),
			string(additionalMeta), string(permissions), playlist.MetadataSource,
			playlist.ID,
		},
	})

	return err
}

func (db *SQLiteDatabase) DeletePlaylist(ctx context.Context, id string) error {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `DELETE FROM playlists WHERE id = ?;`, &sqlitex.ExecOptions{
		Args: []any{id},
	})
	return err
}

func (db *SQLiteDatabase) GetUsers(ctx context.Context) ([]media.DatabaseUser, error) {
	var users []media.DatabaseUser

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return users, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT * FROM users;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			user := media.DatabaseUser{}

			user.ID = stmt.ColumnText(0)
			user.Username = stmt.ColumnText(1)
			user.Email = stmt.ColumnText(2)
			user.PasswordHash = stmt.ColumnText(3)
			user.DisplayName = stmt.ColumnText(4)
			user.Description = stmt.ColumnText(5)
			if err := json.Unmarshal([]byte(stmt.ColumnText(6)), &user.ListenedTo); err != nil {
				return fmt.Errorf("failed to parse listened_to: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(7)), &user.Favorites); err != nil {
				return fmt.Errorf("failed to parse favorites: %w", err)
			}
			user.PublicViewCount = stmt.ColumnInt(8)
			user.CreationDate = stmt.ColumnInt64(9)
			if err := json.Unmarshal([]byte(stmt.ColumnText(10)), &user.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			user.LinkedArtistID = stmt.ColumnText(11)
			if err := json.Unmarshal([]byte(stmt.ColumnText(12)), &user.LinkedSources); err != nil {
				return fmt.Errorf("failed to parse linked_sources: %w", err)
			}

			users = append(users, user)

			return nil
		},
	})

	return users, err
}

func (db *SQLiteDatabase) GetUser(ctx context.Context, id string) (media.DatabaseUser, error) {
	user := media.DatabaseUser{}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return user, err
	}
	defer db.pool.Put(conn)

	scanned := false
	err = sqlitex.Execute(conn, `SELECT * FROM users WHERE id = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			if scanned {
				return ErrTooMany
			}
			scanned = true

			user.ID = stmt.ColumnText(0)
			user.Username = stmt.ColumnText(1)
			user.Email = stmt.ColumnText(2)
			user.PasswordHash = stmt.ColumnText(3)
			user.DisplayName = stmt.ColumnText(4)
			user.Description = stmt.ColumnText(5)
			if err := json.Unmarshal([]byte(stmt.ColumnText(6)), &user.ListenedTo); err != nil {
				return fmt.Errorf("failed to parse listened_to: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(7)), &user.Favorites); err != nil {
				return fmt.Errorf("failed to parse favorites: %w", err)
			}
			user.PublicViewCount = stmt.ColumnInt(8)
			user.CreationDate = stmt.ColumnInt64(9)
			if err := json.Unmarshal([]byte(stmt.ColumnText(10)), &user.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			user.LinkedArtistID = stmt.ColumnText(11)
			if err := json.Unmarshal([]byte(stmt.ColumnText(12)), &user.LinkedSources); err != nil {
				return fmt.Errorf("failed to parse linked_sources: %w", err)
			}

			return nil
		},
		Args: []any{id},
	})
	if err != nil {
		return user, err
	}
	if !scanned {
		return user, ErrNotFound
	}

	return user, nil
}

func (db *SQLiteDatabase) GetUserByUsername(ctx context.Context, username string) (media.DatabaseUser, error) {
	user := media.DatabaseUser{}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return user, err
	}
	defer db.pool.Put(conn)

	scanned := false
	err = sqlitex.Execute(conn, `SELECT * FROM users WHERE username = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			if scanned {
				return ErrTooMany
			}
			scanned = true

			user.ID = stmt.ColumnText(0)
			user.Username = stmt.ColumnText(1)
			user.Email = stmt.ColumnText(2)
			user.PasswordHash = stmt.ColumnText(3)
			user.DisplayName = stmt.ColumnText(4)
			user.Description = stmt.ColumnText(5)
			if err := json.Unmarshal([]byte(stmt.ColumnText(6)), &user.ListenedTo); err != nil {
				return fmt.Errorf("failed to parse listened_to: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(7)), &user.Favorites); err != nil {
				return fmt.Errorf("failed to parse favorites: %w", err)
			}
			user.PublicViewCount = stmt.ColumnInt(8)
			user.CreationDate = stmt.ColumnInt64(9)
			if err := json.Unmarshal([]byte(stmt.ColumnText(10)), &user.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			user.LinkedArtistID = stmt.ColumnText(11)
			if err := json.Unmarshal([]byte(stmt.ColumnText(12)), &user.LinkedSources); err != nil {
				return fmt.Errorf("failed to parse linked_sources: %w", err)
			}

			return nil
		},
		Args: []any{username},
	})
	if err != nil {
		return user, err
	}
	if !scanned {
		return user, ErrNotFound
	}

	return user, nil
}

func (db *SQLiteDatabase) CreateUser(ctx context.Context, user media.DatabaseUser) error {
	// Convert JSON fields to strings.
	listenedTo, err := json.Marshal(user.ListenedTo)
	if err != nil {
		return fmt.Errorf("failed to marshal listened_to: %w", err)
	}
	favorites, err := json.Marshal(user.Favorites)
	if err != nil {
		return fmt.Errorf("failed to marshal favorites: %w", err)
	}
	permissions, err := json.Marshal(user.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	linkedSources, err := json.Marshal(user.LinkedSources)
	if err != nil {
		return fmt.Errorf("failed to marshal linked_sources: %w", err)
	}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
	  INSERT INTO users (
	    id, username, email, password_hash, display_name, description, listened_to,
	    favorites, public_view_count, creation_date, permissions, linked_artist_id,
	    linked_sources
	  ) VALUES (
	   ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	  );`, &sqlitex.ExecOptions{
		Args: []any{
			user.ID, user.Username, user.Email, user.PasswordHash, user.DisplayName,
			user.Description, string(listenedTo), string(favorites), user.PublicViewCount,
			user.CreationDate, string(permissions), user.LinkedArtistID, string(linkedSources),
		},
	})

	return err
}

func (db *SQLiteDatabase) UpdateUser(ctx context.Context, user media.DatabaseUser) error {
	// Convert JSON fields to strings.
	listenedTo, err := json.Marshal(user.ListenedTo)
	if err != nil {
		return fmt.Errorf("failed to marshal listened_to: %w", err)
	}
	favorites, err := json.Marshal(user.Favorites)
	if err != nil {
		return fmt.Errorf("failed to marshal favorites: %w", err)
	}
	permissions, err := json.Marshal(user.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	linkedSources, err := json.Marshal(user.LinkedSources)
	if err != nil {
		return fmt.Errorf("failed to marshal linked_sources: %w", err)
	}

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
	  UPDATE users
	  SET username=?, email=?, password_hash=?, display_name=?, description=?, listened_to=?, 
	      favorites=?, public_view_count=?, creation_date=?, permissions=?, linked_artist_id=?, 
	      linked_sources=?
	  WHERE id=?;`, &sqlitex.ExecOptions{
		Args: []any{
			user.Username, user.Email, user.PasswordHash, user.DisplayName, user.Description,
			string(listenedTo), string(favorites), user.PublicViewCount, user.CreationDate,
			string(permissions), user.LinkedArtistID, string(linkedSources), user.ID,
		},
	})

	return err
}

func (db *SQLiteDatabase) DeleteUser(ctx context.Context, id string) error {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `DELETE FROM users WHERE id = ?;`, &sqlitex.ExecOptions{
		Args: []any{id},
	})
	return err
}

func (db *SQLiteDatabase) GetProviderUser(
	ctx context.Context,
	provider, providerUserID string,
) (media.DatabaseUser, error) {
	var user media.DatabaseUser

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return user, err
	}
	defer db.pool.Put(conn)

	scanned := false
	err = sqlitex.Execute(conn, `
        SELECT u.* FROM users u
        JOIN auth_providers p ON u.id = p.user_id
        WHERE p.provider = ? AND p.provider_user_id = ?;`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			if scanned {
				return ErrTooMany
			}
			scanned = true

			user.ID = stmt.ColumnText(0)
			user.Username = stmt.ColumnText(1)
			user.Email = stmt.ColumnText(2)
			user.PasswordHash = stmt.ColumnText(3)
			user.DisplayName = stmt.ColumnText(4)
			user.Description = stmt.ColumnText(5)
			if err := json.Unmarshal([]byte(stmt.ColumnText(6)), &user.ListenedTo); err != nil {
				return fmt.Errorf("failed to parse listened_to: %w", err)
			}
			if err := json.Unmarshal([]byte(stmt.ColumnText(7)), &user.Favorites); err != nil {
				return fmt.Errorf("failed to parse favorites: %w", err)
			}
			user.PublicViewCount = stmt.ColumnInt(8)
			user.CreationDate = stmt.ColumnInt64(9)
			if err := json.Unmarshal([]byte(stmt.ColumnText(10)), &user.Permissions); err != nil {
				return fmt.Errorf("failed to parse permissions: %w", err)
			}
			user.LinkedArtistID = stmt.ColumnText(11)
			if err := json.Unmarshal([]byte(stmt.ColumnText(12)), &user.LinkedSources); err != nil {
				return fmt.Errorf("failed to parse linked_sources: %w", err)
			}

			return nil
		},
		Args: []any{provider, providerUserID},
	})
	if err != nil {
		return user, err
	}
	if !scanned {
		return user, ErrNotFound
	}

	return user, nil
}

func (db *SQLiteDatabase) IsProviderLinked(ctx context.Context, provider, userID string) (bool, error) {
	var exists bool

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return exists, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(
		conn,
		`SELECT EXISTS(SELECT 1 FROM auth_providers WHERE provider = ? AND user_id = ?);`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				exists = stmt.ColumnBool(0)

				return nil
			},
			Args: []any{provider, userID},
		},
	)

	return exists, err
}

func (db *SQLiteDatabase) LinkProviderAccount(ctx context.Context, provider, userID, providerUserID string) error {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
        INSERT INTO auth_providers (user_id, provider, provider_user_id)
        VALUES (?, ?, ?);`, &sqlitex.ExecOptions{
		Args: []any{userID, provider, providerUserID},
	})

	return err
}

func (db *SQLiteDatabase) DisconnectProviderAccount(ctx context.Context, provider, userID string) error {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `
        DELETE FROM auth_providers WHERE user_id = ? AND provider = ?;`, &sqlitex.ExecOptions{
		Args: []any{userID, provider},
	})

	return err
}

func (db *SQLiteDatabase) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return exists, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?);`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			exists = stmt.ColumnBool(0)

			return nil
		},
		Args: []any{username},
	})

	return exists, err
}

func (db *SQLiteDatabase) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return exists, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?);`, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			exists = stmt.ColumnBool(0)

			return nil
		},
		Args: []any{email},
	})

	return exists, err
}

func (db *SQLiteDatabase) BlacklistToken(ctx context.Context, token string, expiration time.Time) error {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(
		conn,
		`INSERT INTO blacklisted_tokens (token, expiration) VALUES (?, ?);`,
		&sqlitex.ExecOptions{
			Args: []any{token, expiration.Format(time.DateTime)},
		},
	)

	return err
}

func (db *SQLiteDatabase) CleanExpiredTokens(ctx context.Context) error {
	conn, err := db.pool.Take(ctx)
	if err != nil {
		return err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(conn, `DELETE FROM blacklisted_tokens WHERE expiration < datetime('now');`, nil)

	return err
}

func (db *SQLiteDatabase) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	var exists bool

	conn, err := db.pool.Take(ctx)
	if err != nil {
		return exists, err
	}
	defer db.pool.Put(conn)

	err = sqlitex.Execute(
		conn,
		`SELECT EXISTS(SELECT 1 FROM blacklisted_tokens WHERE token = ?);`,
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				exists = stmt.ColumnBool(0)

				return nil
			},
			Args: []any{token},
		},
	)

	return exists, err
}

func init() {
	db := &SQLiteDatabase{}
	Registry["sqlite"] = db
}
