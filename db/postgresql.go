package db

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/LibraMusic/LibraCore/config"
	"github.com/LibraMusic/LibraCore/logging"
	"github.com/LibraMusic/LibraCore/types"
)

type PostgreSQLDatabase struct {
	pool *pgxpool.Pool
}

func ConnectPostgreSQL() (*PostgreSQLDatabase, error) {
	result := &PostgreSQLDatabase{}
	err := result.Connect()
	return result, err
}

func (db *PostgreSQLDatabase) Connect() (err error) {
	logging.Info().Msg("Connecting to PostgreSQL...")
	connStr := "postgresql://" + config.Conf.Database.PostgreSQL.User + ":" + config.Conf.Database.PostgreSQL.Pass + "@" + config.Conf.Database.PostgreSQL.Host + "/" + config.Conf.Database.PostgreSQL.DBName + config.Conf.Database.PostgreSQL.Params
	pool, err := pgxpool.New(context.Background(), connStr)
	db.pool = pool
	if err != nil {
		return
	}

	_, err = db.pool.Exec(context.Background(), "CREATE EXTENSION IF NOT EXISTS hstore;")
	if err != nil {
		return
	}

	if err = db.createTracksTable(); err != nil {
		return
	}
	if err = db.createAlbumsTable(); err != nil {
		return
	}
	if err = db.createVideosTable(); err != nil {
		return
	}
	if err = db.createArtistsTable(); err != nil {
		return
	}
	if err = db.createPlaylistsTable(); err != nil {
		return
	}
	if err = db.createUsersTable(); err != nil {
		return
	}
	if err = db.createBlacklistedTokensTable(); err != nil {
		return
	}

	return
}

func (db *PostgreSQLDatabase) createTracksTable() error {
	_, err := db.pool.Exec(context.Background(), `
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
		  lyrics hstore,
		  listen_count INT,
		  favorite_count INT,
		  addition_date BIGINT,
		  tags TEXT[],
		  additional_meta jsonb,
		  permissions hstore,
		  linked_item_ids TEXT[],
		  content_source TEXT,
		  metadata_source TEXT,
		  lyric_sources hstore
	  );
	`)
	return err
}

func (db *PostgreSQLDatabase) createAlbumsTable() error {
	_, err := db.pool.Exec(context.Background(), `
	  CREATE TABLE IF NOT EXISTS albums (
		  id TEXT PRIMARY KEY,
		  user_id TEXT,
		  upc TEXT,
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
		  permissions hstore,
		  linked_item_ids TEXT[],
		  metadata_source TEXT
	  );
	`)
	return err
}

func (db *PostgreSQLDatabase) createVideosTable() error {
	_, err := db.pool.Exec(context.Background(), `
	  CREATE TABLE IF NOT EXISTS videos (
		  id TEXT PRIMARY KEY,
		  user_id TEXT,
		  title TEXT,
		  artist_ids TEXT[],
		  duration INT,
		  description TEXT,
		  release_date TEXT,
		  subtitles hstore,
		  watch_count INT,
		  favorite_count INT,
		  addition_date BIGINT,
		  tags TEXT[],
		  additional_meta jsonb,
		  permissions hstore,
		  linked_item_ids TEXT[],
		  content_source TEXT,
		  metadata_source TEXT,
		  lyric_sources hstore
	  );
	`)
	return err
}

func (db *PostgreSQLDatabase) createArtistsTable() error {
	_, err := db.pool.Exec(context.Background(), `
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
		  permissions hstore,
		  linked_item_ids TEXT[],
		  metadata_source TEXT
	  );
	`)
	return err
}

func (db *PostgreSQLDatabase) createPlaylistsTable() error {
	_, err := db.pool.Exec(context.Background(), `
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
		  permissions hstore,
		  metadata_source TEXT
	  );
	`)
	return err
}

func (db *PostgreSQLDatabase) createUsersTable() error {
	_, err := db.pool.Exec(context.Background(), `
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
		  creation_date TEXT,
		  permissions hstore,
		  linked_artist_id TEXT,
		  linked_sources hstore
	  );
	`)
	return err
}

func (db *PostgreSQLDatabase) createBlacklistedTokensTable() error {
	_, err := db.pool.Exec(context.Background(), `
	  CREATE TABLE IF NOT EXISTS blacklisted_tokens (
		  token TEXT PRIMARY KEY,
		  expiration TIMESTAMP
	  );
	`)
	return err
}

func (db *PostgreSQLDatabase) Close() error {
	logging.Info().Msg("Closing PostgreSQL connection...")
	db.pool.Close()
	return nil
}

func (*PostgreSQLDatabase) EngineName() string {
	return "PostgreSQL"
}

func (db *PostgreSQLDatabase) GetAllTracks() ([]types.Track, error) {
	var tracks []types.Track
	rows, err := db.pool.Query(context.Background(), "SELECT * FROM tracks;")
	if err != nil {
		return tracks, err
	}
	defer rows.Close()
	for rows.Next() {
		track := types.Track{}
		err = rows.Scan(&track.ID, &track.UserID, &track.ISRC, &track.Title, &track.ArtistIDs, &track.AlbumIDs, &track.PrimaryAlbumID, &track.TrackNumber, &track.Duration, &track.Description, &track.ReleaseDate, &track.Lyrics, &track.ListenCount, &track.FavoriteCount, &track.AdditionDate, &track.Tags, &track.AdditionalMeta, &track.Permissions, &track.LinkedItemIDs, &track.ContentSource, &track.MetadataSource, &track.LyricSources)
		if err != nil {
			return tracks, err
		}
		tracks = append(tracks, track)
	}
	return tracks, err
}

func (db *PostgreSQLDatabase) GetTracks(userID string) ([]types.Track, error) {
	var tracks []types.Track
	rows, err := db.pool.Query(context.Background(), "SELECT * FROM tracks WHERE user_id=$1;", userID)
	if err != nil {
		return tracks, err
	}
	defer rows.Close()
	for rows.Next() {
		track := types.Track{}
		err = rows.Scan(&track.ID, &track.UserID, &track.ISRC, &track.Title, &track.ArtistIDs, &track.AlbumIDs, &track.PrimaryAlbumID, &track.TrackNumber, &track.Duration, &track.Description, &track.ReleaseDate, &track.Lyrics, &track.ListenCount, &track.FavoriteCount, &track.AdditionDate, &track.Tags, &track.AdditionalMeta, &track.Permissions, &track.LinkedItemIDs, &track.ContentSource, &track.MetadataSource, &track.LyricSources)
		if err != nil {
			return tracks, err
		}
		tracks = append(tracks, track)
	}
	return tracks, err
}

func (db *PostgreSQLDatabase) GetTrack(id string) (types.Track, error) {
	track := types.Track{}
	row := db.pool.QueryRow(context.Background(), "SELECT * FROM tracks WHERE id=$1;", id)
	err := row.Scan(&track.ID, &track.UserID, &track.ISRC, &track.Title, &track.ArtistIDs, &track.AlbumIDs, &track.PrimaryAlbumID, &track.TrackNumber, &track.Duration, &track.Description, &track.ReleaseDate, &track.Lyrics, &track.ListenCount, &track.FavoriteCount, &track.AdditionDate, &track.Tags, &track.AdditionalMeta, &track.Permissions, &track.LinkedItemIDs, &track.ContentSource, &track.MetadataSource, &track.LyricSources)

	return track, err
}

func (db *PostgreSQLDatabase) AddTrack(track types.Track) error {
	_, err := db.pool.Exec(context.Background(), `
	  INSERT INTO tracks (
	    id, user_id, isrc, title, artist_ids, album_ids, primary_album_id, track_number, duration, description, release_date, lyrics, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
	  ) VALUES (
	    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
	  );
  `, track.ID, track.UserID, track.ISRC, track.Title, track.ArtistIDs, track.AlbumIDs, track.PrimaryAlbumID, track.TrackNumber, track.Duration, track.Description, track.ReleaseDate, track.Lyrics, track.ListenCount, track.FavoriteCount, track.AdditionDate, track.Tags, track.AdditionalMeta, track.Permissions, track.LinkedItemIDs, track.ContentSource, track.MetadataSource, track.LyricSources)
	return err
}

func (db *PostgreSQLDatabase) UpdateTrack(track types.Track) error {
	_, err := db.pool.Exec(context.Background(), `
	  UPDATE tracks
	  SET user_id=$2, isrc=$3, title=$4, artist_ids=$5, album_ids=$6, primary_album_id=$7, track_number=$8, duration=$9, description=$10, release_date=$11, lyrics=$12, listen_count=$13, favorite_count=$14, addition_date=$15, tags=$16, additional_meta=$17, permissions=$18, linked_item_ids=$19, content_source=$20, metadata_source=$21, lyric_sources=$22
	  WHERE id=$1;
  `, track.ID, track.UserID, track.ISRC, track.Title, track.ArtistIDs, track.AlbumIDs, track.PrimaryAlbumID, track.TrackNumber, track.Duration, track.Description, track.ReleaseDate, track.Lyrics, track.ListenCount, track.FavoriteCount, track.AdditionDate, track.Tags, track.AdditionalMeta, track.Permissions, track.LinkedItemIDs, track.ContentSource, track.MetadataSource, track.LyricSources)
	return err
}

func (db *PostgreSQLDatabase) GetAllAlbums() ([]types.Album, error) {
	var albums []types.Album
	rows, err := db.pool.Query(context.Background(), "SELECT * FROM albums;")
	if err != nil {
		return albums, err
	}
	defer rows.Close()
	for rows.Next() {
		album := types.Album{}
		err = rows.Scan(&album.ID, &album.UserID, &album.UPC, &album.Title, &album.ArtistIDs, &album.TrackIDs, &album.Description, &album.ReleaseDate, &album.ListenCount, &album.FavoriteCount, &album.AdditionDate, &album.Tags, &album.AdditionalMeta, &album.Permissions, &album.LinkedItemIDs, &album.MetadataSource)
		if err != nil {
			return albums, err
		}
		albums = append(albums, album)
	}
	return albums, err
}

func (db *PostgreSQLDatabase) GetAlbums(userID string) ([]types.Album, error) {
	var albums []types.Album
	rows, err := db.pool.Query(context.Background(), "SELECT * FROM albums WHERE user_id=$1;", userID)
	if err != nil {
		return albums, err
	}
	defer rows.Close()
	for rows.Next() {
		album := types.Album{}
		err = rows.Scan(&album.ID, &album.UserID, &album.UPC, &album.Title, &album.ArtistIDs, &album.TrackIDs, &album.Description, &album.ReleaseDate, &album.ListenCount, &album.FavoriteCount, &album.AdditionDate, &album.Tags, &album.AdditionalMeta, &album.Permissions, &album.LinkedItemIDs, &album.MetadataSource)
		if err != nil {
			return albums, err
		}
		albums = append(albums, album)
	}
	return albums, err
}

func (db *PostgreSQLDatabase) GetAlbum(id string) (types.Album, error) {
	album := types.Album{}
	row := db.pool.QueryRow(context.Background(), "SELECT * FROM albums WHERE id=$1;", id)
	err := row.Scan(&album.ID, &album.UserID, &album.UPC, &album.Title, &album.ArtistIDs, &album.TrackIDs, &album.Description, &album.ReleaseDate, &album.ListenCount, &album.FavoriteCount, &album.AdditionDate, &album.Tags, &album.AdditionalMeta, &album.Permissions, &album.LinkedItemIDs, &album.MetadataSource)
	return album, err
}

func (db *PostgreSQLDatabase) AddAlbum(album types.Album) error {
	_, err := db.pool.Exec(context.Background(), `
	  INSERT INTO albums (
	    id, user_id, upc, title, artist_ids, track_ids, description, release_date, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, metadata_source
	  ) VALUES (
	    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
	  );
  `, album.ID, album.UserID, album.UPC, album.Title, album.ArtistIDs, album.TrackIDs, album.Description, album.ReleaseDate, album.ListenCount, album.FavoriteCount, album.AdditionDate, album.Tags, album.AdditionalMeta, album.Permissions, album.LinkedItemIDs, album.MetadataSource)
	return err
}

func (db *PostgreSQLDatabase) UpdateAlbum(album types.Album) error {
	_, err := db.pool.Exec(context.Background(), `
	  UPDATE albums
	  SET user_id=$2, upc=$3, title=$4, artist_ids=$5, track_ids=$6, description=$7, release_date=$8, listen_count=$9, favorite_count=$10, addition_date=$11, tags=$12, additional_meta=$13, permissions=$14, linked_item_ids=$15, metadata_source=$16
	  WHERE id=$1;
  `, album.ID, album.UserID, album.UPC, album.Title, album.ArtistIDs, album.TrackIDs, album.Description, album.ReleaseDate, album.ListenCount, album.FavoriteCount, album.AdditionDate, album.Tags, album.AdditionalMeta, album.Permissions, album.LinkedItemIDs, album.MetadataSource)
	return err
}

func (db *PostgreSQLDatabase) GetAllVideos() ([]types.Video, error) {
	var videos []types.Video
	rows, err := db.pool.Query(context.Background(), "SELECT * FROM videos;")
	if err != nil {
		return videos, err
	}
	defer rows.Close()
	for rows.Next() {
		video := types.Video{}
		err = rows.Scan(&video.ID, &video.UserID, &video.Title, &video.ArtistIDs, &video.Duration, &video.Description, &video.ReleaseDate, &video.Subtitles, &video.WatchCount, &video.FavoriteCount, &video.AdditionDate, &video.Tags, &video.AdditionalMeta, &video.Permissions, &video.LinkedItemIDs, &video.ContentSource, &video.MetadataSource, &video.LyricSources)
		if err != nil {
			return videos, err
		}
		videos = append(videos, video)
	}
	return videos, err
}

func (db *PostgreSQLDatabase) GetVideos(userID string) ([]types.Video, error) {
	var videos []types.Video
	rows, err := db.pool.Query(context.Background(), "SELECT * FROM videos WHERE user_id=$1;", userID)
	if err != nil {
		return videos, err
	}
	defer rows.Close()
	for rows.Next() {
		video := types.Video{}
		err = rows.Scan(&video.ID, &video.UserID, &video.Title, &video.ArtistIDs, &video.Duration, &video.Description, &video.ReleaseDate, &video.Subtitles, &video.WatchCount, &video.FavoriteCount, &video.AdditionDate, &video.Tags, &video.AdditionalMeta, &video.Permissions, &video.LinkedItemIDs, &video.ContentSource, &video.MetadataSource, &video.LyricSources)
		if err != nil {
			return videos, err
		}
		videos = append(videos, video)
	}
	return videos, err
}

func (db *PostgreSQLDatabase) GetVideo(id string) (types.Video, error) {
	video := types.Video{}
	row := db.pool.QueryRow(context.Background(), "SELECT * FROM videos WHERE id=$1;", id)
	err := row.Scan(&video.ID, &video.UserID, &video.Title, &video.ArtistIDs, &video.Duration, &video.Description, &video.ReleaseDate, &video.Subtitles, &video.WatchCount, &video.FavoriteCount, &video.AdditionDate, &video.Tags, &video.AdditionalMeta, &video.Permissions, &video.LinkedItemIDs, &video.ContentSource, &video.MetadataSource, &video.LyricSources)

	return video, err
}

func (db *PostgreSQLDatabase) AddVideo(video types.Video) error {
	_, err := db.pool.Exec(context.Background(), `
	  INSERT INTO videos (
	    id, user_id, title, artist_ids, duration, description, release_date, subtitles, watch_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, content_source, metadata_source, lyric_sources
	  ) VALUES (
	    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
	  );
  `, video.ID, video.UserID, video.Title, video.ArtistIDs, video.Duration, video.Description, video.ReleaseDate, video.Subtitles, video.WatchCount, video.FavoriteCount, video.AdditionDate, video.Tags, video.AdditionalMeta, video.Permissions, video.LinkedItemIDs, video.ContentSource, video.MetadataSource, video.LyricSources)
	return err
}

func (db *PostgreSQLDatabase) UpdateVideo(video types.Video) error {
	_, err := db.pool.Exec(context.Background(), `
	  UPDATE videos
	  SET user_id=$2, title=$3, artist_ids=$4, duration=$5, description=$6, release_date=$7, subtitles=$8, watch_count=$9, favorite_count=$10, addition_date=$11, tags=$12, additional_meta=$13, permissions=$14, linked_item_ids=$15, content_source=$16, metadata_source=$17, lyric_sources=$18
	  WHERE id=$1;
  `, video.ID, video.UserID, video.Title, video.ArtistIDs, video.Duration, video.Description, video.ReleaseDate, video.Subtitles, video.WatchCount, video.FavoriteCount, video.AdditionDate, video.Tags, video.AdditionalMeta, video.Permissions, video.LinkedItemIDs, video.ContentSource, video.MetadataSource, video.LyricSources)
	return err
}

func (db *PostgreSQLDatabase) GetAllArtists() ([]types.Artist, error) {
	var artists []types.Artist
	rows, err := db.pool.Query(context.Background(), "SELECT * FROM artists;")
	if err != nil {
		return artists, err
	}
	defer rows.Close()
	for rows.Next() {
		artist := types.Artist{}
		err = rows.Scan(&artist.ID, &artist.UserID, &artist.Name, &artist.AlbumIDs, &artist.TrackIDs, &artist.Description, &artist.CreationDate, &artist.ListenCount, &artist.FavoriteCount, &artist.AdditionDate, &artist.Tags, &artist.AdditionalMeta, &artist.Permissions, &artist.LinkedItemIDs, &artist.MetadataSource)
		if err != nil {
			return artists, err
		}
		artists = append(artists, artist)
	}
	return artists, err
}

func (db *PostgreSQLDatabase) GetArtists(userID string) ([]types.Artist, error) {
	var artists []types.Artist
	rows, err := db.pool.Query(context.Background(), "SELECT * FROM artists WHERE user_id=$1;", userID)
	if err != nil {
		return artists, err
	}
	defer rows.Close()
	for rows.Next() {
		artist := types.Artist{}
		err = rows.Scan(&artist.ID, &artist.UserID, &artist.Name, &artist.AlbumIDs, &artist.TrackIDs, &artist.Description, &artist.CreationDate, &artist.ListenCount, &artist.FavoriteCount, &artist.AdditionDate, &artist.Tags, &artist.AdditionalMeta, &artist.Permissions, &artist.LinkedItemIDs, &artist.MetadataSource)
		if err != nil {
			return artists, err
		}
		artists = append(artists, artist)
	}
	return artists, err
}

func (db *PostgreSQLDatabase) GetArtist(id string) (types.Artist, error) {
	artist := types.Artist{}
	row := db.pool.QueryRow(context.Background(), "SELECT * FROM artists WHERE id=$1;", id)
	err := row.Scan(&artist.ID, &artist.UserID, &artist.Name, &artist.AlbumIDs, &artist.TrackIDs, &artist.Description, &artist.CreationDate, &artist.ListenCount, &artist.FavoriteCount, &artist.AdditionDate, &artist.Tags, &artist.AdditionalMeta, &artist.Permissions, &artist.LinkedItemIDs, &artist.MetadataSource)
	return artist, err
}

func (db *PostgreSQLDatabase) AddArtist(artist types.Artist) error {
	_, err := db.pool.Exec(context.Background(), `
	  INSERT INTO artists (
	    id, user_id, name, album_ids, track_ids, description, creation_date, listen_count, favorite_count, addition_date, tags, additional_meta, permissions, linked_item_ids, metadata_source
	  ) VALUES (
	    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
	  );
  `, artist.ID, artist.UserID, artist.Name, artist.AlbumIDs, artist.TrackIDs, artist.Description, artist.CreationDate, artist.ListenCount, artist.FavoriteCount, artist.AdditionDate, artist.Tags, artist.AdditionalMeta, artist.Permissions, artist.LinkedItemIDs, artist.MetadataSource)
	return err
}

func (db *PostgreSQLDatabase) UpdateArtist(artist types.Artist) error {
	_, err := db.pool.Exec(context.Background(), `
	  UPDATE artists
	  SET user_id=$2, name=$3, album_ids=$4, track_ids=$5, description=$6, creation_date=$7, listen_count=$8, favorite_count=$9, addition_date=$10, tags=$11, additional_meta=$12, permissions=$13, linked_item_ids=$14, metadata_source=$15
	  WHERE id=$1;
  `, artist.ID, artist.UserID, artist.Name, artist.AlbumIDs, artist.TrackIDs, artist.Description, artist.CreationDate, artist.ListenCount, artist.FavoriteCount, artist.AdditionDate, artist.Tags, artist.AdditionalMeta, artist.Permissions, artist.LinkedItemIDs, artist.MetadataSource)
	return err
}

func (db *PostgreSQLDatabase) GetAllPlaylists() ([]types.Playlist, error) {
	var playlists []types.Playlist
	rows, err := db.pool.Query(context.Background(), "SELECT * FROM playlists;")
	if err != nil {
		return playlists, err
	}
	defer rows.Close()
	for rows.Next() {
		playlist := types.Playlist{}
		err = rows.Scan(&playlist.ID, &playlist.UserID, &playlist.Title, &playlist.TrackIDs, &playlist.ListenCount, &playlist.FavoriteCount, &playlist.Description, &playlist.CreationDate, &playlist.AdditionDate, &playlist.Tags, &playlist.AdditionalMeta, &playlist.Permissions, &playlist.MetadataSource)
		if err != nil {
			return playlists, err
		}
		playlists = append(playlists, playlist)
	}
	return playlists, err
}

func (db *PostgreSQLDatabase) GetPlaylists(userID string) ([]types.Playlist, error) {
	var playlists []types.Playlist
	rows, err := db.pool.Query(context.Background(), "SELECT * FROM playlists WHERE user_id=$1;", userID)
	if err != nil {
		return playlists, err
	}
	defer rows.Close()
	for rows.Next() {
		playlist := types.Playlist{}
		err = rows.Scan(&playlist.ID, &playlist.UserID, &playlist.Title, &playlist.TrackIDs, &playlist.ListenCount, &playlist.FavoriteCount, &playlist.Description, &playlist.CreationDate, &playlist.AdditionDate, &playlist.Tags, &playlist.AdditionalMeta, &playlist.Permissions, &playlist.MetadataSource)
		if err != nil {
			return playlists, err
		}
		playlists = append(playlists, playlist)
	}
	return playlists, err
}

func (db *PostgreSQLDatabase) GetPlaylist(id string) (types.Playlist, error) {
	playlist := types.Playlist{}
	row := db.pool.QueryRow(context.Background(), "SELECT * FROM playlists WHERE id=$1;", id)
	err := row.Scan(&playlist.ID, &playlist.UserID, &playlist.Title, &playlist.TrackIDs, &playlist.ListenCount, &playlist.FavoriteCount, &playlist.Description, &playlist.CreationDate, &playlist.AdditionDate, &playlist.Tags, &playlist.AdditionalMeta, &playlist.Permissions, &playlist.MetadataSource)
	return playlist, err
}

func (db *PostgreSQLDatabase) AddPlaylist(playlist types.Playlist) error {
	_, err := db.pool.Exec(context.Background(), `
	  INSERT INTO playlists (
	    id, user_id, title, track_ids, listen_count, favorite_count, description, creation_date, addition_date, tags, additional_meta, permissions, metadata_source
	  ) VALUES (
	    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
	  );
  `, playlist.ID, playlist.UserID, playlist.Title, playlist.TrackIDs, playlist.ListenCount, playlist.FavoriteCount, playlist.Description, playlist.CreationDate, playlist.AdditionDate, playlist.Tags, playlist.AdditionalMeta, playlist.Permissions, playlist.MetadataSource)
	return err
}

func (db *PostgreSQLDatabase) UpdatePlaylist(playlist types.Playlist) error {
	_, err := db.pool.Exec(context.Background(), `
	  UPDATE playlists
	  SET user_id=$2, title=$3, track_ids=$4, listen_count=$5, favorite_count=$6, description=$7, creation_date=$8, addition_date=$9, tags=$10, additional_meta=$11, permissions=$12, metadata_source=$13
	  WHERE id=$1;
  `, playlist.ID, playlist.UserID, playlist.Title, playlist.TrackIDs, playlist.ListenCount, playlist.FavoriteCount, playlist.Description, playlist.CreationDate, playlist.AdditionDate, playlist.Tags, playlist.AdditionalMeta, playlist.Permissions, playlist.MetadataSource)
	return err
}

func (db *PostgreSQLDatabase) GetUsers() ([]types.User, error) {
	var users []types.User
	rows, err := db.pool.Query(context.Background(), "SELECT * FROM users;")
	if err != nil {
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		user := types.User{}
		err = rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Description, &user.ListenedTo, &user.Favorites, &user.PublicViewCount, &user.CreationDate, &user.Permissions, &user.LinkedArtistID, &user.LinkedSources)
		if err != nil {
			return users, err
		}
		users = append(users, user)
	}
	return users, err
}

func (db *PostgreSQLDatabase) GetUser(id string) (types.User, error) {
	user := types.User{}
	row := db.pool.QueryRow(context.Background(), "SELECT * FROM users WHERE id=$1;", id)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Description, &user.ListenedTo, &user.Favorites, &user.PublicViewCount, &user.CreationDate, &user.Permissions, &user.LinkedArtistID, &user.LinkedSources)
	return user, err
}

func (db *PostgreSQLDatabase) GetUserByUsername(username string) (types.User, error) {
	user := types.User{}
	row := db.pool.QueryRow(context.Background(), "SELECT * FROM users WHERE username=$1 OR email=$1;", strings.ToLower(username))
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Description, &user.ListenedTo, &user.Favorites, &user.PublicViewCount, &user.CreationDate, &user.Permissions, &user.LinkedArtistID, &user.LinkedSources)
	return user, err
}

func (db *PostgreSQLDatabase) CreateUser(user types.User) error {
	_, err := db.pool.Exec(context.Background(), `
	  INSERT INTO users (
	    id, username, email, password_hash, display_name, description, listened_to, favorites, public_view_count, creation_date, permissions, linked_artist_id, linked_sources
	  ) VALUES (
	    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
	  );
  `, user.ID, user.Username, user.Email, user.PasswordHash, user.DisplayName, user.Description, user.ListenedTo, user.Favorites, user.PublicViewCount, user.CreationDate, user.Permissions, user.LinkedArtistID, user.LinkedSources)
	return err
}

func (db *PostgreSQLDatabase) UpdateUser(user types.User) error {
	_, err := db.pool.Exec(context.Background(), `
	  UPDATE users
	  SET username=$2, email=$3, password_hash=$4, display_name=$5, description=$6, listened_to=$7, favorites=$8, public_view_count=$9, creation_date=$10, permissions=$11, linked_artist_id=$12, linked_sources=$13
	  WHERE id=$1;
  `, user.ID, user.Username, user.Email, user.PasswordHash, user.DisplayName, user.Description, user.ListenedTo, user.Favorites, user.PublicViewCount, user.CreationDate, user.Permissions, user.LinkedArtistID, user.LinkedSources)
	return err
}

func (db *PostgreSQLDatabase) UsernameExists(username string) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE username=$1);", strings.ToLower(username)).Scan(&exists)
	return exists, err
}

func (db *PostgreSQLDatabase) EmailExists(email string) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1);", strings.ToLower(email)).Scan(&exists)
	return exists, err
}

func (db *PostgreSQLDatabase) BlacklistToken(token string, expiration time.Time) error {
	_, err := db.pool.Exec(context.Background(), "INSERT INTO blacklisted_tokens (token, expiration) VALUES ($1, $2);", token, expiration)
	return err
}

func (db *PostgreSQLDatabase) CleanExpiredTokens() error {
	_, err := db.pool.Exec(context.Background(), "DELETE FROM blacklisted_tokens WHERE expiration < NOW();")
	return err
}

func (db *PostgreSQLDatabase) IsTokenBlacklisted(token string) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM blacklisted_tokens WHERE token=$1);", token).Scan(&exists)
	return exists, err
}
