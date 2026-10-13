package config

import (
	_ "embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/charmbracelet/log"

	"github.com/libramusic/taurus/v2"

	"github.com/libramusic/libracore/media"
	"github.com/libramusic/libracore/server/routes/auth/providers"
	"github.com/libramusic/libracore/sources"
	"github.com/libramusic/libracore/storage"
)

//go:embed default_config.yaml
var defaultConfig string

const EnvPrefix = "LIBRA"

type Config struct {
	Application ApplicationConfig `yaml:"application"`
	Auth        AuthConfig        `yaml:"auth"`
	General     GeneralConfig     `yaml:"general"`
	Sources     SourcesConfig     `yaml:"sources"`
	Storage     storage.Options   `yaml:"storage"`
	Databases   []DatabaseConfig  `yaml:"databases"`
	Logs        LogsConfig        `yaml:"logs"`

	// DataDir is the persistent data directory for this configuration.
	// Since it's used to locate the config, it isn't serialized to/from YAML.
	DataDir string `yaml:"-"`
}

type ApplicationConfig struct {
	Port          int            `yaml:"port"`
	PublicURL     string         `yaml:"public_url"`
	SourceID      string         `yaml:"source_id"`
	SourceName    string         `yaml:"source_name"`
	SourceFormats []media.Format `yaml:"source_formats"`
}

type AuthConfig struct {
	JWT                        JWTAuthConfig            `yaml:"jwt"`
	Providers                  []providers.AuthProvider `yaml:"providers"`
	GlobalAPIRoutesRequireAuth bool                     `yaml:"global_api_routes_require_auth"`
	UserAPIRoutesRequireAuth   bool                     `yaml:"user_api_routes_require_auth"`
	UserAPIRequireSameUseUser  bool                     `yaml:"user_api_require_same_user"`
	DisableAccountCreation     bool                     `yaml:"disable_account_creation"`
}

type JWTAuthConfig struct {
	SigningMethod          string        `yaml:"signing_method"`
	SigningKey             string        `yaml:"signing_key"`
	RefreshTokenExpiration time.Duration `yaml:"refresh_token_expiration"`
	AccessTokenExpiration  time.Duration `yaml:"access_token_expiration"`
}

type GeneralConfig struct {
	IDLength                  int                               `yaml:"id_length"`
	InheritListenCounts       bool                              `yaml:"inherit_listen_counts"`
	ArtistListenCountsByTrack bool                              `yaml:"artist_listen_counts_by_track"`
	UserArtistLinking         bool                              `yaml:"user_artist_linking"`
	MaxSearchResults          int                               `yaml:"max_search_results"`
	MaxTrackDuration          time.Duration                     `yaml:"max_track_duration"`
	ReservedUsernames         []string                          `yaml:"reserved_usernames"`
	CustomDisplayNames        bool                              `yaml:"custom_display_names"`
	ReserveDisplayNames       bool                              `yaml:"reserve_display_names"`
	AdminPermissions          map[string]media.AdminPermissions `yaml:"admin_permissions"`
	DocumentFeedRoutes        bool                              `yaml:"document_feed_routes"`
}

type SourcesConfig struct {
	Definitions []SourceDefinition    `yaml:"definitions"`
	Priorities  []sources.BindingInfo `yaml:"priorities"`
}

type SourceDefinition struct {
	Type    string         `yaml:"type"`
	Alias   string         `yaml:"alias"`
	Options map[string]any `yaml:"options"`
}

type DatabaseConfig struct {
	Engine   string         `yaml:"engine"`
	Purposes []string       `yaml:"purposes"`
	Options  map[string]any `yaml:"options"`
}

type LogsConfig struct {
	Level  log.Level `yaml:"level"`
	Format string    `yaml:"format"`
}

func LoadConfig(dataDir string) (Config, error) {
	taurus.SetEnvPrefix(EnvPrefix)
	taurus.SetExpandEnv(true)
	taurus.BindEnvAlias("APPLICATION_PORT", "PORT")
	taurus.BindEnvAlias("LOGS_LEVEL", "LOG_LEVEL")
	taurus.BindEnvAlias("LOGS_FORMAT", "LOG_FORMAT")
	setupTypes()

	conf := LoadDefaultConfig()

	if dataDir == "" {
		dataDir = os.Getenv(EnvPrefix + "_DATA_DIR")
	}
	conf.DataDir = dataDir
	conf.Storage.DataDir = dataDir

	configFilePath, err := configFilePath(dataDir)
	if err != nil {
		return conf, err
	}

	if _, err := os.Stat(configFilePath); errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(configFilePath, []byte(defaultConfig), 0o644); err != nil {
			log.Warn("Failed to write default config", "err", err)
		}
	}

	if err := taurus.LoadFile(configFilePath, &conf); err != nil {
		log.Warn("Failed to read config", "err", err)
		log.Info("Using default config")
	}

	if err := taurus.LoadEnv(EnvPrefix, &conf); err != nil {
		log.Warn("Failed to read and merge environment variables", "err", err)
	}

	if err := taurus.LoadFlags(&conf); err != nil {
		log.Warn("Failed to read and merge flags", "err", err)
	}

	return conf, nil
}

func LoadDefaultConfig() Config {
	var conf Config
	if err := taurus.Load(defaultConfig, &conf); err != nil {
		log.Fatal("Failed to load default config", "err", err)
	}
	return conf
}

func setupTypes() {
	taurus.RegisterMarshaler(nil, func(level log.Level) ([]byte, error) {
		return []byte(level.String()), nil
	})

	taurus.RegisterUnmarshaler(nil, func(level *log.Level, data []byte) error {
		var err error
		*level, err = log.ParseLevel(string(data))
		return err
	})
	taurus.RegisterUnmarshaler(nil, func(duration *time.Duration, data []byte) error {
		var err error
		*duration, err = ParseHumanDuration(string(data))
		return err
	})
}

func configFilePath(dataDir string) (string, error) {
	if dataDir != "" {
		path, err := filepath.Abs(filepath.Join(dataDir, "config.yaml"))
		if err == nil {
			err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				log.Warn(
					"Failed to create directories for config.yaml in data directory",
					"dataDir",
					dataDir,
					"err",
					err,
				)
				return "", err
			}
			return path, nil
		}
		log.Warn("Failed to get absolute path for config.yaml in data directory", "dataDir", dataDir, "err", err)
	}

	path, err := filepath.Abs("config.yaml")
	if err != nil {
		log.Warn("Failed to get absolute path for config.yaml", "err", err)
	} else if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	path, err = os.UserConfigDir()
	if err != nil {
		log.Warn("Failed to get user config directory", "err", err)
	} else {
		path = filepath.Join(path, "libra", "config.yaml")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			if _, err := os.Stat("/etc/libra/config.yaml"); err == nil {
				path = "/etc/libra/config.yaml"
			}
		}
	}

	return path, nil
}
