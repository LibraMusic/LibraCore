package db

//go:generate go run ../tools/dbgen/main.go

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"slices"
	"strings"
	"time"

	"github.com/libramusic/libracore/media"
)

//go:embed migrations
var migrationsFS embed.FS

var (
	ErrNotFound      = errors.New("not found in database")
	ErrTooMany       = errors.New("too many found in database")
	ErrUnknownEngine = errors.New("unknown database engine")
)

// Factory provides methods for database driver registration and instantiation.
// Each driver registers a matcher with Register() during init().
type Factory interface {
	// EngineName returns the human-readable name of the database engine.
	EngineName() string
	// Satisfies checks if this driver can handle the given engine identifier.
	Satisfies(engine string) bool
	// New creates a new Database instance with the given options and dataDir.
	// The dataDir parameter is passed for drivers that need it (e.g., SQLite).
	New(options map[string]any, dataDir string) (Database, error)
}

// registry holds all registered database drivers.
var registry []Factory

// Register adds a database driver to the registry.
// This should be called from init() in the driver package.
func Register(driver Factory) {
	registry = append(registry, driver)
}

// New creates a new Database instance for the given engine and options.
// The dataDir parameter is passed for drivers that need it (e.g., SQLite).
func New(engine string, options map[string]any, dataDir string) (Database, error) {
	for _, driver := range registry {
		if driver.Satisfies(engine) {
			return driver.New(options, dataDir)
		}
	}
	return nil, ErrUnknownEngine
}

type Database interface {
	EngineName() string

	Connect() error
	// Closes the database connection.
	// Subsequent calls will always return nil.
	Close() error

	MigrateUp(steps int) error
	MigrateDown(steps int) error

	AllTracks(ctx context.Context) ([]media.Track, error)
	Tracks(ctx context.Context, userID string) ([]media.Track, error)
	Track(ctx context.Context, id string) (media.Track, error)
	AddTrack(ctx context.Context, track media.Track) error
	UpdateTrack(ctx context.Context, track media.Track) error
	DeleteTrack(ctx context.Context, id string) error

	AllAlbums(ctx context.Context) ([]media.Album, error)
	Albums(ctx context.Context, userID string) ([]media.Album, error)
	Album(ctx context.Context, id string) (media.Album, error)
	AddAlbum(ctx context.Context, album media.Album) error
	UpdateAlbum(ctx context.Context, album media.Album) error
	DeleteAlbum(ctx context.Context, id string) error

	AllVideos(ctx context.Context) ([]media.Video, error)
	Videos(ctx context.Context, userID string) ([]media.Video, error)
	Video(ctx context.Context, id string) (media.Video, error)
	AddVideo(ctx context.Context, video media.Video) error
	UpdateVideo(ctx context.Context, video media.Video) error
	DeleteVideo(ctx context.Context, id string) error

	AllArtists(ctx context.Context) ([]media.Artist, error)
	Artists(ctx context.Context, userID string) ([]media.Artist, error)
	Artist(ctx context.Context, id string) (media.Artist, error)
	AddArtist(ctx context.Context, artist media.Artist) error
	UpdateArtist(ctx context.Context, artist media.Artist) error
	DeleteArtist(ctx context.Context, id string) error

	AllPlaylists(ctx context.Context) ([]media.Playlist, error)
	Playlists(ctx context.Context, userID string) ([]media.Playlist, error)
	Playlist(ctx context.Context, id string) (media.Playlist, error)
	AddPlaylist(ctx context.Context, playlist media.Playlist) error
	UpdatePlaylist(ctx context.Context, playlist media.Playlist) error
	DeletePlaylist(ctx context.Context, id string) error

	Users(ctx context.Context) ([]media.DatabaseUser, error)
	User(ctx context.Context, id string) (media.DatabaseUser, error)
	UserByUsername(ctx context.Context, username string) (media.DatabaseUser, error)
	AddUser(ctx context.Context, user media.DatabaseUser) error
	UpdateUser(ctx context.Context, user media.DatabaseUser) error
	UsernameExists(ctx context.Context, username string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	DeleteUser(ctx context.Context, id string) error

	ProviderUser(ctx context.Context, provider, providerUserID string) (media.DatabaseUser, error)
	IsProviderLinked(ctx context.Context, provider, userID string) (bool, error)
	LinkProviderAccount(ctx context.Context, provider, userID, providerUserID string) error
	DisconnectProviderAccount(ctx context.Context, provider, userID string) error

	BlacklistToken(ctx context.Context, token string, expiration time.Time) error
	CleanExpiredTokens(ctx context.Context) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

// TODO: Add a way to filter the types of playables that are returned so we don't perform unnecessary database queries.
func AllPlayables(ctx context.Context, db Database) ([]media.Playable, error) {
	var playables []media.Playable

	tracks, err := db.AllTracks(ctx)
	if err != nil {
		return nil, err
	}
	for _, track := range tracks {
		playables = append(playables, track)
	}

	albums, err := db.AllAlbums(ctx)
	if err != nil {
		return nil, err
	}
	for _, album := range albums {
		playables = append(playables, album)
	}

	videos, err := db.AllVideos(ctx)
	if err != nil {
		return nil, err
	}
	for _, video := range videos {
		playables = append(playables, video)
	}

	artists, err := db.AllArtists(ctx)
	if err != nil {
		return nil, err
	}
	for _, artist := range artists {
		playables = append(playables, artist)
	}

	playlists, err := db.AllPlaylists(ctx)
	if err != nil {
		return nil, err
	}
	for _, playlist := range playlists {
		playables = append(playables, playlist)
	}

	return playables, nil
}

// TODO: Add a way to filter the types of playables that are returned so we don't perform unnecessary database queries.
func Playables(ctx context.Context, db Database, userID string) ([]media.Playable, error) {
	var playables []media.Playable

	tracks, err := db.Tracks(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, track := range tracks {
		playables = append(playables, track)
	}

	albums, err := db.Albums(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, album := range albums {
		playables = append(playables, album)
	}

	videos, err := db.Videos(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, video := range videos {
		playables = append(playables, video)
	}

	artists, err := db.Artists(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, artist := range artists {
		playables = append(playables, artist)
	}

	playlists, err := db.Playlists(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, playlist := range playlists {
		playables = append(playables, playlist)
	}

	return playables, nil
}

func OrderedMigrationFiles(entries []fs.DirEntry, up bool) []string {
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
