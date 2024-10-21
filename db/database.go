package db

import (
	"strings"
	"time"

	"github.com/DevReaper0/libra/config"
	"github.com/DevReaper0/libra/logging"
	"github.com/DevReaper0/libra/types"
)

var DB Database

type Database interface {
	Connect() error
	Close() error
	EngineName() string

	GetAllTracks() ([]types.Track, error)
	GetTracks(userID string) ([]types.Track, error)
	GetTrack(id string) (types.Track, error)
	AddTrack(track types.Track) error
	UpdateTrack(track types.Track) error

	GetAllAlbums() ([]types.Album, error)
	GetAlbums(userID string) ([]types.Album, error)
	GetAlbum(id string) (types.Album, error)
	AddAlbum(album types.Album) error
	UpdateAlbum(album types.Album) error

	GetAllVideos() ([]types.Video, error)
	GetVideos(userID string) ([]types.Video, error)
	GetVideo(id string) (types.Video, error)
	AddVideo(video types.Video) error
	UpdateVideo(video types.Video) error

	GetAllArtists() ([]types.Artist, error)
	GetArtists(userID string) ([]types.Artist, error)
	GetArtist(id string) (types.Artist, error)
	AddArtist(artist types.Artist) error
	UpdateArtist(artist types.Artist) error

	GetAllPlaylists() ([]types.Playlist, error)
	GetPlaylists(userID string) ([]types.Playlist, error)
	GetPlaylist(id string) (types.Playlist, error)
	AddPlaylist(playlist types.Playlist) error
	UpdatePlaylist(playlist types.Playlist) error

	GetUsers() ([]types.User, error)
	GetUser(id string) (types.User, error)
	GetUserByUsername(username string) (types.User, error)
	CreateUser(user types.User) error
	UpdateUser(user types.User) error
	UsernameExists(username string) (bool, error)
	EmailExists(email string) (bool, error)

	BlacklistToken(token string, expiration time.Time) error
	CleanExpiredTokens() error
	IsTokenBlacklisted(token string) (bool, error)
}

func ConnectDatabase() {
	if DB != nil {
		logging.Warn().Msg("Database already connected")
		return
	}

	var err error

	switch strings.ToLower(config.Conf.Database.Engine) {
	case "sqlite", "sqlite3":
		DB, err = ConnectSQLite()
	case "postgresql", "postgres", "postgre", "pgsql", "psql", "pg":
		DB, err = ConnectPostgreSQL()
	default:
		logging.Fatal().Msgf("Unsupported database engine: %s\n", config.Conf.Database.Engine)
	}
	if err != nil {
		logging.Fatal().Err(err).Msg("Error connecting to database")
	}
}

// TODO: Add a way to filter the types of playables that are returned so we don't perform unnecessary database queries
func GetAllPlayables() ([]types.Playable, error) {
	var playables []types.Playable

	tracks, err := DB.GetAllTracks()
	if err != nil {
		return nil, err
	}
	for _, track := range tracks {
		playables = append(playables, track)
	}

	albums, err := DB.GetAllAlbums()
	if err != nil {
		return nil, err
	}
	for _, album := range albums {
		playables = append(playables, album)
	}

	videos, err := DB.GetAllVideos()
	if err != nil {
		return nil, err
	}
	for _, video := range videos {
		playables = append(playables, video)
	}

	artists, err := DB.GetAllArtists()
	if err != nil {
		return nil, err
	}
	for _, artist := range artists {
		playables = append(playables, artist)
	}

	playlists, err := DB.GetAllPlaylists()
	if err != nil {
		return nil, err
	}
	for _, playlist := range playlists {
		playables = append(playables, playlist)
	}

	return playables, nil
}

// TODO: Add a way to filter the types of playables that are returned so we don't perform unnecessary database queries
func GetPlayables(userID string) ([]types.Playable, error) {
	var playables []types.Playable

	tracks, err := DB.GetTracks(userID)
	if err != nil {
		return nil, err
	}
	for _, track := range tracks {
		playables = append(playables, track)
	}

	albums, err := DB.GetAlbums(userID)
	if err != nil {
		return nil, err
	}
	for _, album := range albums {
		playables = append(playables, album)
	}

	videos, err := DB.GetVideos(userID)
	if err != nil {
		return nil, err
	}
	for _, video := range videos {
		playables = append(playables, video)
	}

	artists, err := DB.GetArtists(userID)
	if err != nil {
		return nil, err
	}
	for _, artist := range artists {
		playables = append(playables, artist)
	}

	playlists, err := DB.GetPlaylists(userID)
	if err != nil {
		return nil, err
	}
	for _, playlist := range playlists {
		playables = append(playables, playlist)
	}

	return playables, nil
}
