package db

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/log"

	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/types"
)

var (
	Registry = map[string]Database{}
	DB       Database
)

//go:embed migrations
var migrationsFS embed.FS

var (
	ErrNotFound = errors.New("not found in database")
	ErrTooMany  = errors.New("too many found in database")
)

type Database interface {
	Satisfies(engine string) bool

	Connect() error
	Close() error
	EngineName() string

	MigrateUp(steps int) error
	MigrateDown(steps int) error

	GetAllTracks(ctx context.Context) ([]types.Track, error)
	GetTracks(ctx context.Context, userID string) ([]types.Track, error)
	GetTrack(ctx context.Context, id string) (types.Track, error)
	AddTrack(ctx context.Context, track types.Track) error
	UpdateTrack(ctx context.Context, track types.Track) error
	DeleteTrack(ctx context.Context, id string) error

	GetAllAlbums(ctx context.Context) ([]types.Album, error)
	GetAlbums(ctx context.Context, userID string) ([]types.Album, error)
	GetAlbum(ctx context.Context, id string) (types.Album, error)
	AddAlbum(ctx context.Context, album types.Album) error
	UpdateAlbum(ctx context.Context, album types.Album) error
	DeleteAlbum(ctx context.Context, id string) error

	GetAllVideos(ctx context.Context) ([]types.Video, error)
	GetVideos(ctx context.Context, userID string) ([]types.Video, error)
	GetVideo(ctx context.Context, id string) (types.Video, error)
	AddVideo(ctx context.Context, video types.Video) error
	UpdateVideo(ctx context.Context, video types.Video) error
	DeleteVideo(ctx context.Context, id string) error

	GetAllArtists(ctx context.Context) ([]types.Artist, error)
	GetArtists(ctx context.Context, userID string) ([]types.Artist, error)
	GetArtist(ctx context.Context, id string) (types.Artist, error)
	AddArtist(ctx context.Context, artist types.Artist) error
	UpdateArtist(ctx context.Context, artist types.Artist) error
	DeleteArtist(ctx context.Context, id string) error

	GetAllPlaylists(ctx context.Context) ([]types.Playlist, error)
	GetPlaylists(ctx context.Context, userID string) ([]types.Playlist, error)
	GetPlaylist(ctx context.Context, id string) (types.Playlist, error)
	AddPlaylist(ctx context.Context, playlist types.Playlist) error
	UpdatePlaylist(ctx context.Context, playlist types.Playlist) error
	DeletePlaylist(ctx context.Context, id string) error

	GetUsers(ctx context.Context) ([]types.DatabaseUser, error)
	GetUser(ctx context.Context, id string) (types.DatabaseUser, error)
	GetUserByUsername(ctx context.Context, username string) (types.DatabaseUser, error)
	CreateUser(ctx context.Context, user types.DatabaseUser) error
	UpdateUser(ctx context.Context, user types.DatabaseUser) error
	UsernameExists(ctx context.Context, username string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	DeleteUser(ctx context.Context, id string) error

	GetProviderUser(ctx context.Context, provider, providerUserID string) (types.DatabaseUser, error)
	IsProviderLinked(ctx context.Context, provider, userID string) (bool, error)
	LinkProviderAccount(ctx context.Context, provider, userID, providerUserID string) error
	DisconnectProviderAccount(ctx context.Context, provider, userID string) error

	BlacklistToken(ctx context.Context, token string, expiration time.Time) error
	CleanExpiredTokens(ctx context.Context) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

func ConnectDatabase() error {
	if DB != nil {
		log.Warn("Database already connected")
		return nil
	}

	for _, db := range Registry {
		if db.Satisfies(config.Conf.Database.Engine) {
			DB = db
			if err := DB.Connect(); err != nil {
				return fmt.Errorf("error connecting to database: %w", err)
			}
			log.Info("Connected to database", "engine", db.EngineName())
			return nil
		}
	}
	return fmt.Errorf("unsupported database engine: %s", config.Conf.Database.Engine)
}

// TODO: Add a way to filter the types of playables that are returned so we don't perform unnecessary database queries.
func GetAllPlayables(ctx context.Context) ([]types.Playable, error) {
	var playables []types.Playable

	tracks, err := DB.GetAllTracks(ctx)
	if err != nil {
		return nil, err
	}
	for _, track := range tracks {
		playables = append(playables, track)
	}

	albums, err := DB.GetAllAlbums(ctx)
	if err != nil {
		return nil, err
	}
	for _, album := range albums {
		playables = append(playables, album)
	}

	videos, err := DB.GetAllVideos(ctx)
	if err != nil {
		return nil, err
	}
	for _, video := range videos {
		playables = append(playables, video)
	}

	artists, err := DB.GetAllArtists(ctx)
	if err != nil {
		return nil, err
	}
	for _, artist := range artists {
		playables = append(playables, artist)
	}

	playlists, err := DB.GetAllPlaylists(ctx)
	if err != nil {
		return nil, err
	}
	for _, playlist := range playlists {
		playables = append(playables, playlist)
	}

	return playables, nil
}

// TODO: Add a way to filter the types of playables that are returned so we don't perform unnecessary database queries.
func GetPlayables(ctx context.Context, userID string) ([]types.Playable, error) {
	var playables []types.Playable

	tracks, err := DB.GetTracks(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, track := range tracks {
		playables = append(playables, track)
	}

	albums, err := DB.GetAlbums(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, album := range albums {
		playables = append(playables, album)
	}

	videos, err := DB.GetVideos(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, video := range videos {
		playables = append(playables, video)
	}

	artists, err := DB.GetArtists(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, artist := range artists {
		playables = append(playables, artist)
	}

	playlists, err := DB.GetPlaylists(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, playlist := range playlists {
		playables = append(playables, playlist)
	}

	return playables, nil
}

func GetOrderedMigrationFiles(entries []fs.DirEntry, up bool) []string {
	var files []string
	for _, entry := range entries {
		name := entry.Name()
		if up && strings.Contains(name, ".down.") {
			continue
		}
		if !up && strings.Contains(name, ".up.") {
			continue
		}
		files = append(files, name)
	}

	slices.Sort(files)
	if !up {
		// Down migrations are applied in reverse order.
		slices.Reverse(files)
	}
	return files
}
