package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/goccy/go-json"
	_ "github.com/mattn/go-sqlite3"

	"github.com/LibraMusic/LibraCore/config"
	"github.com/LibraMusic/LibraCore/logging"
	"github.com/LibraMusic/LibraCore/types"
)

type SQLiteDatabase struct {
	sqlDB *sql.DB
}

func ConnectSQLite() (*SQLiteDatabase, error) {
	result := &SQLiteDatabase{}
	err := result.Connect()
	return result, err
}

func (db *SQLiteDatabase) Connect() (err error) {
	logging.Info().Msg("Connecting to SQLite...")
	sqlDB, err := sql.Open("sqlite3", config.Conf.Database.SQLite.Path)
	db.sqlDB = sqlDB
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
      addition_date INTEGER, -- BIGINT replaced with INTEGER
      tags TEXT, -- JSON (json array)
      additional_meta BLOB, -- JSONB (json object)
      permissions TEXT, -- JSON (json object)
      linked_item_ids TEXT, -- JSON (json array)
      content_source TEXT,
      metadata_source TEXT,
      lyric_sources TEXT -- JSON (json object)
    );
  `)
	return err
}

func (db *SQLiteDatabase) createAlbumsTable() error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) createVideosTable() error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) createArtistsTable() error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) createPlaylistsTable() error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) createUsersTable() error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) createBlacklistedTokensTable() error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) Close() error {
	logging.Info().Msg("Closing SQLite connection...")
	return db.sqlDB.Close()
}

func (*SQLiteDatabase) EngineName() string {
	return "SQLite"
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
	logging.Error().Msg("unimplemented")
	return nil, nil
}

func (db *SQLiteDatabase) GetAlbums(userID string) ([]types.Album, error) {
	logging.Error().Msg("unimplemented")
	return nil, nil
}

func (db *SQLiteDatabase) GetAlbum(id string) (types.Album, error) {
	logging.Error().Msg("unimplemented")
	return types.Album{}, nil
}

func (db *SQLiteDatabase) AddAlbum(album types.Album) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) UpdateAlbum(album types.Album) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) DeleteAlbum(id string) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) GetAllVideos() ([]types.Video, error) {
	logging.Error().Msg("unimplemented")
	return nil, nil
}

func (db *SQLiteDatabase) GetVideos(userID string) ([]types.Video, error) {
	logging.Error().Msg("unimplemented")
	return nil, nil
}

func (db *SQLiteDatabase) GetVideo(id string) (types.Video, error) {
	logging.Error().Msg("unimplemented")
	return types.Video{}, nil
}

func (db *SQLiteDatabase) AddVideo(video types.Video) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) UpdateVideo(video types.Video) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) DeleteVideo(id string) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) GetAllArtists() ([]types.Artist, error) {
	logging.Error().Msg("unimplemented")
	return nil, nil
}

func (db *SQLiteDatabase) GetArtists(userID string) ([]types.Artist, error) {
	logging.Error().Msg("unimplemented")
	return nil, nil
}

func (db *SQLiteDatabase) GetArtist(id string) (types.Artist, error) {
	logging.Error().Msg("unimplemented")
	return types.Artist{}, nil
}

func (db *SQLiteDatabase) AddArtist(artist types.Artist) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) UpdateArtist(artist types.Artist) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) DeleteArtist(id string) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) GetAllPlaylists() ([]types.Playlist, error) {
	logging.Error().Msg("unimplemented")
	return nil, nil
}

func (db *SQLiteDatabase) GetPlaylists(userID string) ([]types.Playlist, error) {
	logging.Error().Msg("unimplemented")
	return nil, nil
}

func (db *SQLiteDatabase) GetPlaylist(id string) (types.Playlist, error) {
	logging.Error().Msg("unimplemented")
	return types.Playlist{}, nil
}

func (db *SQLiteDatabase) AddPlaylist(playlist types.Playlist) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) UpdatePlaylist(playlist types.Playlist) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) DeletePlaylist(id string) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) GetUsers() ([]types.User, error) {
	logging.Error().Msg("unimplemented")
	return nil, nil
}

func (db *SQLiteDatabase) GetUser(id string) (types.User, error) {
	logging.Error().Msg("unimplemented")
	return types.User{}, nil
}

func (db *SQLiteDatabase) GetUserByUsername(username string) (types.User, error) {
	logging.Error().Msg("unimplemented")
	return types.User{}, nil
}

func (db *SQLiteDatabase) CreateUser(user types.User) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) UpdateUser(user types.User) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) DeleteUser(id string) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) UsernameExists(username string) (bool, error) {
	logging.Error().Msg("unimplemented")
	return false, nil
}

func (db *SQLiteDatabase) EmailExists(email string) (bool, error) {
	logging.Error().Msg("unimplemented")
	return false, nil
}

func (db *SQLiteDatabase) BlacklistToken(token string, expiration time.Time) error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) CleanExpiredTokens() error {
	logging.Error().Msg("unimplemented")
	return nil
}

func (db *SQLiteDatabase) IsTokenBlacklisted(token string) (bool, error) {
	logging.Error().Msg("unimplemented")
	return false, nil
}
