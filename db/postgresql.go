//go:build postgresql_db || !(no_postgresql_db || no_dbs)

package db

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/media"
)

type PostgreSQLDatabase struct {
	pool      *pgxpool.Pool
	closeOnce sync.Once
}

func (*PostgreSQLDatabase) EngineName() string {
	return "PostgreSQL"
}

func (*PostgreSQLDatabase) Satisfies(engine string) bool {
	return slices.Contains([]string{
		"postgresql",
		"postgres",
		"postgre",
		"pgsql",
		"pg",
	}, strings.ToLower(engine))
}

func (db *PostgreSQLDatabase) Connect() error {
	log.Info("Connecting to PostgreSQL...")
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s %s",
		config.Conf.Database.PostgreSQL.Host,
		config.Conf.Database.PostgreSQL.Port,
		config.Conf.Database.PostgreSQL.User,
		config.Conf.Database.PostgreSQL.Pass,
		config.Conf.Database.PostgreSQL.DBName,
		config.Conf.Database.PostgreSQL.Params,
	)
	pool, err := pgxpool.New(context.Background(), connStr)
	db.pool = pool
	if err != nil {
		return normalizePostgreSQLError(err)
	}

	// If the migrations table doesn't exist, create it and run migrations.
	exists, err := db.migrationsTableExists(context.Background())
	if err != nil {
		return err
	}
	if !exists {
		if err := db.createMigrationsTable(context.Background()); err != nil {
			return err
		}
		if err := db.MigrateUp(-1); err != nil {
			return err
		}
	}

	return nil
}

func (db *PostgreSQLDatabase) Close() error {
	db.closeOnce.Do(func() {
		log.Info("Closing PostgreSQL connection...")
		db.pool.Close()
	})
	return nil
}

func (db *PostgreSQLDatabase) migrationsTableExists(ctx context.Context) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_name = 'schema_migrations'
        );
    `).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return exists, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) createMigrationsTable(ctx context.Context) error {
	_, err := db.pool.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version BIGINT PRIMARY KEY,
            dirty BOOLEAN
        );
    `)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) getCurrentVersion(ctx context.Context) (uint64, bool, error) {
	var version uint64
	var dirty bool
	err := db.pool.QueryRow(ctx, `
        SELECT version, dirty FROM schema_migrations 
        ORDER BY version DESC LIMIT 1;
    `).Scan(&version, &dirty)
	if err != nil {
		return 0, false, normalizePostgreSQLError(err)
	}

	return version, dirty, nil
}

func (db *PostgreSQLDatabase) setVersion(ctx context.Context, version uint64, dirty bool) error {
	_, err := db.pool.Exec(ctx, `
        DELETE FROM schema_migrations;
        INSERT INTO schema_migrations (version, dirty) VALUES ($1, $2);
    `, version, dirty)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) MigrateUp(steps int) error {
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

	entries, err := migrationsFS.ReadDir("migrations/postgresql")
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
		if err := db.setVersion(context.Background(), version, true); err != nil {
			return err
		}

		// Read and execute migration.
		content, err := migrationsFS.ReadFile(filepath.Join("migrations/postgresql", file))
		if err != nil {
			return err
		}

		_, err = db.pool.Exec(context.Background(), string(content))
		if err != nil {
			return normalizePostgreSQLError(err)
		}

		// Clear dirty flag after successful migration.
		if err := db.setVersion(context.Background(), version, false); err != nil {
			return err
		}

		appliedCount++
	}

	return nil
}

func (db *PostgreSQLDatabase) MigrateDown(steps int) error {
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

	entries, err := migrationsFS.ReadDir("migrations/postgresql")
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
		if err := db.setVersion(context.Background(), version, true); err != nil {
			return err
		}

		// Read and execute migration.
		content, err := migrationsFS.ReadFile(filepath.Join("migrations/postgresql", file))
		if err != nil {
			return err
		}

		_, err = db.pool.Exec(context.Background(), string(content))
		if err != nil {
			return normalizePostgreSQLError(err)
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
		if err := db.setVersion(context.Background(), prevVersion, false); err != nil {
			return err
		}

		appliedCount++
	}

	return nil
}

func (db *PostgreSQLDatabase) GetAllTracks(ctx context.Context) ([]media.Track, error) {
	var tracks []media.Track
	rows, err := db.pool.Query(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM tracks;
    `)
	if err != nil {
		return tracks, normalizePostgreSQLError(err)
	}
	defer rows.Close()
	for rows.Next() {
		track := media.Track{}
		err = rows.Scan(
			&track.ID,
			&track.UserID,
			&track.ISRC,
			&track.Title,
			&track.ArtistIDs,
			&track.AlbumIDs,
			&track.PrimaryAlbumID,
			&track.TrackNumber,
			&track.Duration,
			&track.Description,
			&track.ReleaseDate,
			&track.Lyrics,
			&track.ListenCount,
			&track.FavoriteCount,
			&track.AdditionDate,
			&track.Tags,
			&track.AdditionalMeta,
			&track.Permissions,
			&track.LinkedItemIDs,
			&track.ContentSource,
			&track.MetadataSource,
			&track.LyricSources,
		)
		if err != nil {
			return tracks, normalizePostgreSQLError(err)
		}
		tracks = append(tracks, track)
	}
	return tracks, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetTracks(ctx context.Context, userID string) ([]media.Track, error) {
	var tracks []media.Track
	rows, err := db.pool.Query(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM tracks WHERE user_id=$1;
    `, userID)
	if err != nil {
		return tracks, normalizePostgreSQLError(err)
	}
	defer rows.Close()
	for rows.Next() {
		track := media.Track{}
		err = rows.Scan(
			&track.ID,
			&track.UserID,
			&track.ISRC,
			&track.Title,
			&track.ArtistIDs,
			&track.AlbumIDs,
			&track.PrimaryAlbumID,
			&track.TrackNumber,
			&track.Duration,
			&track.Description,
			&track.ReleaseDate,
			&track.Lyrics,
			&track.ListenCount,
			&track.FavoriteCount,
			&track.AdditionDate,
			&track.Tags,
			&track.AdditionalMeta,
			&track.Permissions,
			&track.LinkedItemIDs,
			&track.ContentSource,
			&track.MetadataSource,
			&track.LyricSources,
		)
		if err != nil {
			return tracks, normalizePostgreSQLError(err)
		}
		tracks = append(tracks, track)
	}
	return tracks, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetTrack(ctx context.Context, id string) (media.Track, error) {
	track := media.Track{}
	row := db.pool.QueryRow(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM tracks WHERE id=$1;
    `, id)
	err := row.Scan(
		&track.ID,
		&track.UserID,
		&track.ISRC,
		&track.Title,
		&track.ArtistIDs,
		&track.AlbumIDs,
		&track.PrimaryAlbumID,
		&track.TrackNumber,
		&track.Duration,
		&track.Description,
		&track.ReleaseDate,
		&track.Lyrics,
		&track.ListenCount,
		&track.FavoriteCount,
		&track.AdditionDate,
		&track.Tags,
		&track.AdditionalMeta,
		&track.Permissions,
		&track.LinkedItemIDs,
		&track.ContentSource,
		&track.MetadataSource,
		&track.LyricSources,
	)

	return track, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) AddTrack(ctx context.Context, track media.Track) error {
	_, err := db.pool.Exec(ctx, `
        INSERT INTO tracks (
            id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
        );
    `, track.ID, track.UserID, track.ISRC, track.Title, track.ArtistIDs, track.AlbumIDs, track.PrimaryAlbumID, track.TrackNumber, track.Duration, track.Description, track.ReleaseDate, track.Lyrics, track.ListenCount, track.FavoriteCount, track.AdditionDate, track.Tags, track.AdditionalMeta, track.Permissions, track.LinkedItemIDs, track.ContentSource, track.MetadataSource, track.LyricSources)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) UpdateTrack(ctx context.Context, track media.Track) error {
	_, err := db.pool.Exec(ctx, `
        UPDATE tracks
        SET user_id=$2, isrc=$3, title=$4, artist_ids=$5, album_ids=$6, primary_album_id=$7, track_number=$8, duration=$9, description=$10, release_date=$11, lyrics=$12, listen_count=$13, favorite_count=$14, addition_date=$15, tags=$16, additional_meta=$17, permissions=$18, linked_item_ids=$19, content_source=$20, metadata_source=$21, lyric_sources=$22
        WHERE id=$1;
    `, track.ID, track.UserID, track.ISRC, track.Title, track.ArtistIDs, track.AlbumIDs, track.PrimaryAlbumID, track.TrackNumber, track.Duration, track.Description, track.ReleaseDate, track.Lyrics, track.ListenCount, track.FavoriteCount, track.AdditionDate, track.Tags, track.AdditionalMeta, track.Permissions, track.LinkedItemIDs, track.ContentSource, track.MetadataSource, track.LyricSources)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) DeleteTrack(ctx context.Context, id string) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM tracks WHERE id=$1;`, id)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetAllAlbums(ctx context.Context) ([]media.Album, error) {
	var albums []media.Album
	rows, err := db.pool.Query(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM albums;
    `)
	if err != nil {
		return albums, normalizePostgreSQLError(err)
	}
	defer rows.Close()
	for rows.Next() {
		album := media.Album{}
		err = rows.Scan(
			&album.ID,
			&album.UserID,
			&album.UPC,
			&album.EAN,
			&album.Title,
			&album.ArtistIDs,
			&album.TrackIDs,
			&album.Description,
			&album.ReleaseDate,
			&album.ListenCount,
			&album.FavoriteCount,
			&album.AdditionDate,
			&album.Tags,
			&album.AdditionalMeta,
			&album.Permissions,
			&album.LinkedItemIDs,
			&album.MetadataSource,
		)
		if err != nil {
			return albums, normalizePostgreSQLError(err)
		}
		albums = append(albums, album)
	}
	return albums, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetAlbums(ctx context.Context, userID string) ([]media.Album, error) {
	var albums []media.Album
	rows, err := db.pool.Query(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM albums WHERE user_id=$1;
    `, userID)
	if err != nil {
		return albums, normalizePostgreSQLError(err)
	}
	defer rows.Close()
	for rows.Next() {
		album := media.Album{}
		err = rows.Scan(
			&album.ID,
			&album.UserID,
			&album.UPC,
			&album.EAN,
			&album.Title,
			&album.ArtistIDs,
			&album.TrackIDs,
			&album.Description,
			&album.ReleaseDate,
			&album.ListenCount,
			&album.FavoriteCount,
			&album.AdditionDate,
			&album.Tags,
			&album.AdditionalMeta,
			&album.Permissions,
			&album.LinkedItemIDs,
			&album.MetadataSource,
		)
		if err != nil {
			return albums, normalizePostgreSQLError(err)
		}
		albums = append(albums, album)
	}
	return albums, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetAlbum(ctx context.Context, id string) (media.Album, error) {
	album := media.Album{}
	row := db.pool.QueryRow(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM albums WHERE id=$1;
    `, id)
	err := row.Scan(
		&album.ID,
		&album.UserID,
		&album.UPC,
		&album.EAN,
		&album.Title,
		&album.ArtistIDs,
		&album.TrackIDs,
		&album.Description,
		&album.ReleaseDate,
		&album.ListenCount,
		&album.FavoriteCount,
		&album.AdditionDate,
		&album.Tags,
		&album.AdditionalMeta,
		&album.Permissions,
		&album.LinkedItemIDs,
		&album.MetadataSource,
	)
	return album, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) AddAlbum(ctx context.Context, album media.Album) error {
	_, err := db.pool.Exec(ctx, `
        INSERT INTO albums (
            id, user_id, upc, ean, title, artist_ids, track_ids, description, release_date, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, metadata_source
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
        );
    `, album.ID, album.UserID, album.UPC, album.EAN, album.Title, album.ArtistIDs, album.TrackIDs, album.Description, album.ReleaseDate, album.ListenCount, album.FavoriteCount, album.AdditionDate, album.Tags, album.AdditionalMeta, album.Permissions, album.LinkedItemIDs, album.MetadataSource)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) UpdateAlbum(ctx context.Context, album media.Album) error {
	_, err := db.pool.Exec(ctx, `
        UPDATE albums
        SET user_id=$2, upc=$3, ean=$4, title=$5, artist_ids=$6, track_ids=$7, description=$8, release_date=$9, listen_count=$10, favorite_count=$11, addition_date=$12, tags=$13, additional_meta=$14, permissions=$15, linked_item_ids=$16, metadata_source=$17
        WHERE id=$1;
    `, album.ID, album.UserID, album.UPC, album.EAN, album.Title, album.ArtistIDs, album.TrackIDs, album.Description, album.ReleaseDate, album.ListenCount, album.FavoriteCount, album.AdditionDate, album.Tags, album.AdditionalMeta, album.Permissions, album.LinkedItemIDs, album.MetadataSource)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) DeleteAlbum(ctx context.Context, id string) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM albums WHERE id=$1;`, id)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetAllVideos(ctx context.Context) ([]media.Video, error) {
	var videos []media.Video
	rows, err := db.pool.Query(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM videos;
    `)
	if err != nil {
		return videos, normalizePostgreSQLError(err)
	}
	defer rows.Close()
	for rows.Next() {
		video := media.Video{}
		err = rows.Scan(
			&video.ID,
			&video.UserID,
			&video.Title,
			&video.ArtistIDs,
			&video.Duration,
			&video.Description,
			&video.ReleaseDate,
			&video.Subtitles,
			&video.WatchCount,
			&video.FavoriteCount,
			&video.AdditionDate,
			&video.Tags,
			&video.AdditionalMeta,
			&video.Permissions,
			&video.LinkedItemIDs,
			&video.ContentSource,
			&video.MetadataSource,
			&video.LyricSources,
		)
		if err != nil {
			return videos, normalizePostgreSQLError(err)
		}
		videos = append(videos, video)
	}
	return videos, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetVideos(ctx context.Context, userID string) ([]media.Video, error) {
	var videos []media.Video
	rows, err := db.pool.Query(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM videos WHERE user_id=$1;
    `, userID)
	if err != nil {
		return videos, normalizePostgreSQLError(err)
	}
	defer rows.Close()
	for rows.Next() {
		video := media.Video{}
		err = rows.Scan(
			&video.ID,
			&video.UserID,
			&video.Title,
			&video.ArtistIDs,
			&video.Duration,
			&video.Description,
			&video.ReleaseDate,
			&video.Subtitles,
			&video.WatchCount,
			&video.FavoriteCount,
			&video.AdditionDate,
			&video.Tags,
			&video.AdditionalMeta,
			&video.Permissions,
			&video.LinkedItemIDs,
			&video.ContentSource,
			&video.MetadataSource,
			&video.LyricSources,
		)
		if err != nil {
			return videos, normalizePostgreSQLError(err)
		}
		videos = append(videos, video)
	}
	return videos, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetVideo(ctx context.Context, id string) (media.Video, error) {
	video := media.Video{}
	row := db.pool.QueryRow(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM videos WHERE id=$1;
    `, id)
	err := row.Scan(
		&video.ID,
		&video.UserID,
		&video.Title,
		&video.ArtistIDs,
		&video.Duration,
		&video.Description,
		&video.ReleaseDate,
		&video.Subtitles,
		&video.WatchCount,
		&video.FavoriteCount,
		&video.AdditionDate,
		&video.Tags,
		&video.AdditionalMeta,
		&video.Permissions,
		&video.LinkedItemIDs,
		&video.ContentSource,
		&video.MetadataSource,
		&video.LyricSources,
	)

	return video, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) AddVideo(ctx context.Context, video media.Video) error {
	_, err := db.pool.Exec(ctx, `
        INSERT INTO videos (
            id, user_id, title, artist_ids, duration, description, release_date, subtitles, watch_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
        );
    `, video.ID, video.UserID, video.Title, video.ArtistIDs, video.Duration, video.Description, video.ReleaseDate, video.Subtitles, video.WatchCount, video.FavoriteCount, video.AdditionDate, video.Tags, video.AdditionalMeta, video.Permissions, video.LinkedItemIDs, video.ContentSource, video.MetadataSource, video.LyricSources)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) UpdateVideo(ctx context.Context, video media.Video) error {
	_, err := db.pool.Exec(ctx, `
        UPDATE videos
        SET user_id=$2, title=$3, artist_ids=$4, duration=$5, description=$6, release_date=$7, subtitles=$8, watch_count=$9, favorite_count=$10, addition_date=$11, tags=$12, additional_meta=$13, permissions=$14, linked_item_ids=$15, content_source=$16, metadata_source=$17, lyric_sources=$18
        WHERE id=$1;
    `, video.ID, video.UserID, video.Title, video.ArtistIDs, video.Duration, video.Description, video.ReleaseDate, video.Subtitles, video.WatchCount, video.FavoriteCount, video.AdditionDate, video.Tags, video.AdditionalMeta, video.Permissions, video.LinkedItemIDs, video.ContentSource, video.MetadataSource, video.LyricSources)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) DeleteVideo(ctx context.Context, id string) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM videos WHERE id=$1;`, id)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetAllArtists(ctx context.Context) ([]media.Artist, error) {
	var artists []media.Artist
	rows, err := db.pool.Query(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM artists;
    `)
	if err != nil {
		return artists, normalizePostgreSQLError(err)
	}
	defer rows.Close()
	for rows.Next() {
		artist := media.Artist{}
		err = rows.Scan(
			&artist.ID,
			&artist.UserID,
			&artist.Name,
			&artist.AlbumIDs,
			&artist.TrackIDs,
			&artist.Description,
			&artist.CreationDate,
			&artist.ListenCount,
			&artist.FavoriteCount,
			&artist.AdditionDate,
			&artist.Tags,
			&artist.AdditionalMeta,
			&artist.Permissions,
			&artist.LinkedItemIDs,
			&artist.MetadataSource,
		)
		if err != nil {
			return artists, normalizePostgreSQLError(err)
		}
		artists = append(artists, artist)
	}
	return artists, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetArtists(ctx context.Context, userID string) ([]media.Artist, error) {
	var artists []media.Artist
	rows, err := db.pool.Query(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM artists WHERE user_id=$1;
    `, userID)
	if err != nil {
		return artists, normalizePostgreSQLError(err)
	}
	defer rows.Close()
	for rows.Next() {
		artist := media.Artist{}
		err = rows.Scan(
			&artist.ID,
			&artist.UserID,
			&artist.Name,
			&artist.AlbumIDs,
			&artist.TrackIDs,
			&artist.Description,
			&artist.CreationDate,
			&artist.ListenCount,
			&artist.FavoriteCount,
			&artist.AdditionDate,
			&artist.Tags,
			&artist.AdditionalMeta,
			&artist.Permissions,
			&artist.LinkedItemIDs,
			&artist.MetadataSource,
		)
		if err != nil {
			return artists, normalizePostgreSQLError(err)
		}
		artists = append(artists, artist)
	}
	return artists, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetArtist(ctx context.Context, id string) (media.Artist, error) {
	artist := media.Artist{}
	row := db.pool.QueryRow(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM artists WHERE id=$1;
    `, id)
	err := row.Scan(
		&artist.ID,
		&artist.UserID,
		&artist.Name,
		&artist.AlbumIDs,
		&artist.TrackIDs,
		&artist.Description,
		&artist.CreationDate,
		&artist.ListenCount,
		&artist.FavoriteCount,
		&artist.AdditionDate,
		&artist.Tags,
		&artist.AdditionalMeta,
		&artist.Permissions,
		&artist.LinkedItemIDs,
		&artist.MetadataSource,
	)
	return artist, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) AddArtist(ctx context.Context, artist media.Artist) error {
	_, err := db.pool.Exec(ctx, `
        INSERT INTO artists (
            id, user_id, name, album_ids, track_ids, description, creation_date, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, metadata_source
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
        );
    `, artist.ID, artist.UserID, artist.Name, artist.AlbumIDs, artist.TrackIDs, artist.Description, artist.CreationDate, artist.ListenCount, artist.FavoriteCount, artist.AdditionDate, artist.Tags, artist.AdditionalMeta, artist.Permissions, artist.LinkedItemIDs, artist.MetadataSource)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) UpdateArtist(ctx context.Context, artist media.Artist) error {
	_, err := db.pool.Exec(ctx, `
        UPDATE artists
        SET user_id=$2, name=$3, album_ids=$4, track_ids=$5, description=$6, creation_date=$7, listen_count=$8, favorite_count=$9, addition_date=$10, tags=$11, additional_meta=$12, permissions=$13, linked_item_ids=$14, metadata_source=$15
        WHERE id=$1;
    `, artist.ID, artist.UserID, artist.Name, artist.AlbumIDs, artist.TrackIDs, artist.Description, artist.CreationDate, artist.ListenCount, artist.FavoriteCount, artist.AdditionDate, artist.Tags, artist.AdditionalMeta, artist.Permissions, artist.LinkedItemIDs, artist.MetadataSource)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) DeleteArtist(ctx context.Context, id string) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM artists WHERE id=$1;`, id)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetAllPlaylists(ctx context.Context) ([]media.Playlist, error) {
	var playlists []media.Playlist
	rows, err := db.pool.Query(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM playlists;
    `)
	if err != nil {
		return playlists, normalizePostgreSQLError(err)
	}
	defer rows.Close()
	for rows.Next() {
		playlist := media.Playlist{}
		err = rows.Scan(
			&playlist.ID,
			&playlist.UserID,
			&playlist.Title,
			&playlist.TrackIDs,
			&playlist.ListenCount,
			&playlist.FavoriteCount,
			&playlist.Description,
			&playlist.CreationDate,
			&playlist.AdditionDate,
			&playlist.Tags,
			&playlist.AdditionalMeta,
			&playlist.Permissions,
			&playlist.MetadataSource,
		)
		if err != nil {
			return playlists, normalizePostgreSQLError(err)
		}
		playlists = append(playlists, playlist)
	}
	return playlists, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetPlaylists(ctx context.Context, userID string) ([]media.Playlist, error) {
	var playlists []media.Playlist
	rows, err := db.pool.Query(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM playlists WHERE user_id=$1;
    `, userID)
	if err != nil {
		return playlists, normalizePostgreSQLError(err)
	}
	defer rows.Close()
	for rows.Next() {
		playlist := media.Playlist{}
		err = rows.Scan(
			&playlist.ID,
			&playlist.UserID,
			&playlist.Title,
			&playlist.TrackIDs,
			&playlist.ListenCount,
			&playlist.FavoriteCount,
			&playlist.Description,
			&playlist.CreationDate,
			&playlist.AdditionDate,
			&playlist.Tags,
			&playlist.AdditionalMeta,
			&playlist.Permissions,
			&playlist.MetadataSource,
		)
		if err != nil {
			return playlists, normalizePostgreSQLError(err)
		}
		playlists = append(playlists, playlist)
	}
	return playlists, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetPlaylist(ctx context.Context, id string) (media.Playlist, error) {
	playlist := media.Playlist{}
	row := db.pool.QueryRow(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM playlists WHERE id=$1;
    `, id)
	err := row.Scan(
		&playlist.ID,
		&playlist.UserID,
		&playlist.Title,
		&playlist.TrackIDs,
		&playlist.ListenCount,
		&playlist.FavoriteCount,
		&playlist.Description,
		&playlist.CreationDate,
		&playlist.AdditionDate,
		&playlist.Tags,
		&playlist.AdditionalMeta,
		&playlist.Permissions,
		&playlist.MetadataSource,
	)
	return playlist, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) AddPlaylist(ctx context.Context, playlist media.Playlist) error {
	_, err := db.pool.Exec(ctx, `
        INSERT INTO playlists (
            id, user_id, title, track_ids, listen_count, favorite_count, description, creation_date, addition_date, tags, additional_meta, permissions, metadata_source
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
        );
    `, playlist.ID, playlist.UserID, playlist.Title, playlist.TrackIDs, playlist.ListenCount, playlist.FavoriteCount, playlist.Description, playlist.CreationDate, playlist.AdditionDate, playlist.Tags, playlist.AdditionalMeta, playlist.Permissions, playlist.MetadataSource)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) UpdatePlaylist(ctx context.Context, playlist media.Playlist) error {
	_, err := db.pool.Exec(ctx, `
        UPDATE playlists
        SET user_id=$2, title=$3, track_ids=$4, listen_count=$5, favorite_count=$6, description=$7, creation_date=$8, addition_date=$9, tags=$10, additional_meta=$11, permissions=$12, metadata_source=$13
        WHERE id=$1;
    `, playlist.ID, playlist.UserID, playlist.Title, playlist.TrackIDs, playlist.ListenCount, playlist.FavoriteCount, playlist.Description, playlist.CreationDate, playlist.AdditionDate, playlist.Tags, playlist.AdditionalMeta, playlist.Permissions, playlist.MetadataSource)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) DeletePlaylist(ctx context.Context, id string) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM playlists WHERE id=$1;`, id)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetUsers(ctx context.Context) ([]media.DatabaseUser, error) {
	var users []media.DatabaseUser
	rows, err := db.pool.Query(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM users;
    `)
	if err != nil {
		return users, normalizePostgreSQLError(err)
	}
	defer rows.Close()
	for rows.Next() {
		user := media.DatabaseUser{}
		err = rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.DisplayName,
			&user.Description,
			&user.ListenedTo,
			&user.Favorites,
			&user.PublicViewCount,
			&user.CreationDate,
			&user.Permissions,
			&user.LinkedArtistID,
			&user.LinkedSources,
		)
		if err != nil {
			return users, normalizePostgreSQLError(err)
		}
		users = append(users, user)
	}
	return users, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetUser(ctx context.Context, id string) (media.DatabaseUser, error) {
	user := media.DatabaseUser{}
	row := db.pool.QueryRow(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM users WHERE id=$1;
    `, id)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.Description,
		&user.ListenedTo,
		&user.Favorites,
		&user.PublicViewCount,
		&user.CreationDate,
		&user.Permissions,
		&user.LinkedArtistID,
		&user.LinkedSources,
	)
	return user, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetUserByUsername(ctx context.Context, username string) (media.DatabaseUser, error) {
	user := media.DatabaseUser{}
	row := db.pool.QueryRow(ctx, `
        SELECT id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
        FROM users WHERE username=$1 OR email=$1;
    `, strings.ToLower(username))
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.Description,
		&user.ListenedTo,
		&user.Favorites,
		&user.PublicViewCount,
		&user.CreationDate,
		&user.Permissions,
		&user.LinkedArtistID,
		&user.LinkedSources,
	)
	return user, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) CreateUser(ctx context.Context, user media.DatabaseUser) error {
	_, err := db.pool.Exec(ctx, `
        INSERT INTO users (
            id, username, email, password_hash, display_name, description, listened_to, favorites, public_view_count, creation_date, permissions, linked_artist_id, linked_sources
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
        );
    `, user.ID, user.Username, user.Email, user.PasswordHash, user.DisplayName, user.Description, user.ListenedTo, user.Favorites, user.PublicViewCount, user.CreationDate, user.Permissions, user.LinkedArtistID, user.LinkedSources)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) UpdateUser(ctx context.Context, user media.DatabaseUser) error {
	_, err := db.pool.Exec(ctx, `
        UPDATE users
        SET username=$2, email=$3, password_hash=$4, display_name=$5, description=$6, listened_to=$7, favorites=$8, public_view_count=$9, creation_date=$10, permissions=$11, linked_artist_id=$12, linked_sources=$13
        WHERE id=$1;
    `, user.ID, user.Username, user.Email, user.PasswordHash, user.DisplayName, user.Description, user.ListenedTo, user.Favorites, user.PublicViewCount, user.CreationDate, user.Permissions, user.LinkedArtistID, user.LinkedSources)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) DeleteUser(ctx context.Context, id string) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM users WHERE id=$1;`, id)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) GetProviderUser(
	ctx context.Context,
	provider, providerUserID string,
) (media.DatabaseUser, error) {
	var user media.DatabaseUser
	row := db.pool.QueryRow(ctx, `
        SELECT u.id, u.user_id, u.isrc, u.title, u.artist_ids, u.album_ids, u.primary_album_id, u.track_number, u.duration, u.description, u.release_date, u.lyrics, u.listen_count, u.favorite_count, u.addition_date, u.tags, u.additional_meta, u.permissions, u.linked_item_ids, u.content_source, u.metadata_source, u.lyric_sources
        FROM users u
        JOIN auth_providers p ON u.id = p.user_id
        WHERE p.provider = $1 AND p.provider_user_id = $2;
    `, provider, providerUserID)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.Description,
		&user.ListenedTo,
		&user.Favorites,
		&user.PublicViewCount,
		&user.CreationDate,
		&user.Permissions,
		&user.LinkedArtistID,
		&user.LinkedSources,
	)
	return user, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) IsProviderLinked(ctx context.Context, provider, userID string) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM auth_providers WHERE user_id = $1 AND provider = $2
        );
    `, userID, provider).Scan(&exists)
	return exists, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) LinkProviderAccount(ctx context.Context, provider, userID, providerUserID string) error {
	_, err := db.pool.Exec(ctx, `
        INSERT INTO auth_providers (user_id, provider, provider_user_id)
        VALUES ($1, $2, $3);
    `, userID, provider, providerUserID)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) DisconnectProviderAccount(ctx context.Context, provider, userID string) error {
	_, err := db.pool.Exec(ctx, `
        DELETE FROM auth_providers WHERE user_id = $1 AND provider = $2;
    `, userID, provider)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username=$1);`, strings.ToLower(username)).
		Scan(&exists)
	return exists, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1);`, strings.ToLower(email)).
		Scan(&exists)
	return exists, normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) BlacklistToken(ctx context.Context, token string, expiration time.Time) error {
	_, err := db.pool.Exec(
		ctx,
		`INSERT INTO blacklisted_tokens (token, expiration) VALUES ($1, $2);`,
		token,
		expiration,
	)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) CleanExpiredTokens(ctx context.Context) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM blacklisted_tokens WHERE expiration < NOW();`)
	return normalizePostgreSQLError(err)
}

func (db *PostgreSQLDatabase) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM blacklisted_tokens WHERE token=$1);`, token).Scan(&exists)
	return exists, normalizePostgreSQLError(err)
}

func normalizePostgreSQLError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func init() {
	db := &PostgreSQLDatabase{}
	Registry["postgresql"] = db
}
