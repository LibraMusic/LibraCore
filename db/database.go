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
	"github.com/libramusic/libracore/media"
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
	EngineName() string
	Satisfies(engine string) bool

	Connect() error
	// Closes the database connection.
	// Subsequent calls will always return nil.
	Close() error

	MigrateUp(steps int) error
	MigrateDown(steps int) error

	GetAllTracks(ctx context.Context) ([]media.Track, error)
	GetTracks(ctx context.Context, userID string) ([]media.Track, error)
	GetTrack(ctx context.Context, id string) (media.Track, error)
	AddTrack(ctx context.Context, track media.Track) error
	UpdateTrack(ctx context.Context, track media.Track) error
	DeleteTrack(ctx context.Context, id string) error

	GetAllAlbums(ctx context.Context) ([]media.Album, error)
	GetAlbums(ctx context.Context, userID string) ([]media.Album, error)
	GetAlbum(ctx context.Context, id string) (media.Album, error)
	AddAlbum(ctx context.Context, album media.Album) error
	UpdateAlbum(ctx context.Context, album media.Album) error
	DeleteAlbum(ctx context.Context, id string) error

	GetAllVideos(ctx context.Context) ([]media.Video, error)
	GetVideos(ctx context.Context, userID string) ([]media.Video, error)
	GetVideo(ctx context.Context, id string) (media.Video, error)
	AddVideo(ctx context.Context, video media.Video) error
	UpdateVideo(ctx context.Context, video media.Video) error
	DeleteVideo(ctx context.Context, id string) error

	GetAllArtists(ctx context.Context) ([]media.Artist, error)
	GetArtists(ctx context.Context, userID string) ([]media.Artist, error)
	GetArtist(ctx context.Context, id string) (media.Artist, error)
	AddArtist(ctx context.Context, artist media.Artist) error
	UpdateArtist(ctx context.Context, artist media.Artist) error
	DeleteArtist(ctx context.Context, id string) error

	GetAllPlaylists(ctx context.Context) ([]media.Playlist, error)
	GetPlaylists(ctx context.Context, userID string) ([]media.Playlist, error)
	GetPlaylist(ctx context.Context, id string) (media.Playlist, error)
	AddPlaylist(ctx context.Context, playlist media.Playlist) error
	UpdatePlaylist(ctx context.Context, playlist media.Playlist) error
	DeletePlaylist(ctx context.Context, id string) error

	GetUsers(ctx context.Context) ([]media.DatabaseUser, error)
	GetUser(ctx context.Context, id string) (media.DatabaseUser, error)
	GetUserByUsername(ctx context.Context, username string) (media.DatabaseUser, error)
	CreateUser(ctx context.Context, user media.DatabaseUser) error
	UpdateUser(ctx context.Context, user media.DatabaseUser) error
	UsernameExists(ctx context.Context, username string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	DeleteUser(ctx context.Context, id string) error

	GetProviderUser(ctx context.Context, provider, providerUserID string) (media.DatabaseUser, error)
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
func GetAllPlayables(ctx context.Context) ([]media.Playable, error) {
	var playables []media.Playable

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
func GetPlayables(ctx context.Context, userID string) ([]media.Playable, error) {
	var playables []media.Playable

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
