package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/charmbracelet/log"

	"github.com/libramusic/taurus"

	"github.com/libramusic/libracore/api"
	"github.com/libramusic/libracore/types"
	"github.com/libramusic/libracore/utils"
)

var Conf Config

//go:embed default_config.yaml
var defaultConfig string

const EnvPrefix = "LIBRA"

var DataDir string

type ApplicationConfig struct {
	Port       int      `yaml:"port"`
	PublicURL  string   `yaml:"public_url"`
	SourceID   string   `yaml:"source_id"`
	SourceName string   `yaml:"source_name"`
	MediaTypes []string `yaml:"media_types"`
}

type JWTAuthConfig struct {
	SigningMethod          string        `yaml:"signing_method"`
	SigningKey             string        `yaml:"signing_key"`
	RefreshTokenExpiration time.Duration `yaml:"refresh_token_expiration"`
	AccessTokenExpiration  time.Duration `yaml:"access_token_expiration"`
}

type OAuthAuthConfig struct {
	Providers []api.OAuthProvider `yaml:"providers"`
}

type AuthConfig struct {
	JWT                        JWTAuthConfig   `yaml:"jwt"`
	OAuth                      OAuthAuthConfig `yaml:"oauth"`
	GlobalAPIRoutesRequireAuth bool            `yaml:"global_api_routes_require_auth"`
	UserAPIRoutesRequireAuth   bool            `yaml:"user_api_routes_require_auth"`
	UserAPIRequireSameUseUser  bool            `yaml:"user_api_require_same_user"`
	DisableAccountCreation     bool            `yaml:"disable_account_creation"`
}

type GeneralConfig struct {
	IDLength                  int                               `yaml:"id_length"`
	IncludeVideoResults       bool                              `yaml:"include_video_results"`
	VideoAudioOnly            bool                              `yaml:"video_audio_only"`
	InheritListenCounts       bool                              `yaml:"inherit_listen_counts"`
	ArtistListenCountsByTrack bool                              `yaml:"artist_listen_counts_by_track"`
	UserArtistLinking         bool                              `yaml:"user_artist_linking"`
	MaxSearchResults          int                               `yaml:"max_search_results"`
	MaxTrackDuration          time.Duration                     `yaml:"max_track_duration"`
	ReservedUsernames         []string                          `yaml:"reserved_usernames"`
	CustomDisplayNames        bool                              `yaml:"custom_display_names"`
	ReserveDisplayNames       bool                              `yaml:"reserve_display_names"`
	AdminPermissions          map[string]types.AdminPermissions `yaml:"admin_permissions"`
	EnabledSources            []string                          `yaml:"enabled_sources"`
}

type SourceScriptsConfig struct {
	PythonCommand   string `yaml:"python_command"`
	YouTubeLocation string `yaml:"youtube_location"`
}

type LogsConfig struct {
	Level  log.Level `yaml:"level"`
	Format string    `yaml:"format"`
}

type StorageConfig struct {
	Location            string            `yaml:"location"`
	SizeLimit           datasize.ByteSize `yaml:"size_limit"`
	MinimumAgeThreshold time.Duration     `yaml:"minimum_age_threshold"`
}

type SQLiteDatabaseConfig struct {
	Path string `yaml:"path"`
}

type PostgreSQLDatabaseConfig struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	User   string `yaml:"user"`
	Pass   string `yaml:"pass"`
	DBName string `yaml:"db_name"`
	Params string `yaml:"params"`
}

type DatabaseConfig struct {
	Engine     string                   `yaml:"engine"`
	SQLite     SQLiteDatabaseConfig     `yaml:"sqlite"`
	PostgreSQL PostgreSQLDatabaseConfig `yaml:"postgresql"`
}

type Config struct {
	Application   ApplicationConfig   `yaml:"application"`
	Auth          AuthConfig          `yaml:"auth"`
	General       GeneralConfig       `yaml:"general"`
	SourceScripts SourceScriptsConfig `yaml:"source_scripts"`
	Logs          LogsConfig          `yaml:"logs"`
	Storage       StorageConfig       `yaml:"storage"`
	Database      DatabaseConfig      `yaml:"database"`
}

func LoadConfig() error {
	taurus.SetEnvPrefix(EnvPrefix)
	taurus.SetExpandEnv(true)
	taurus.BindEnvAlias("APPLICATION_PORT", "PORT")
	taurus.BindEnvAlias("LOGS_LEVEL", "LOG_LEVEL")
	taurus.BindEnvAlias("LOGS_FORMAT", "LOG_FORMAT")
	setupTypes()

	if DataDir == "" {
		DataDir = os.Getenv(EnvPrefix + "_DATA_DIR")
	}

	if err := taurus.Load(defaultConfig, &Conf); err != nil {
		return fmt.Errorf("failed to read default config: %w", err)
	}

	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configFilePath); errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(configFilePath, []byte(defaultConfig), 0o644); err != nil {
			log.Warn("Failed to write default config", "err", err)
		}
	}

	if err := taurus.LoadFile(configFilePath, &Conf); err != nil {
		log.Warn("Failed to read config", "err", err)
		log.Info("Using default config")
	}

	if err := taurus.LoadEnv(EnvPrefix, &Conf); err != nil {
		log.Warn("Failed to read and merge environment variables", "err", err)
	}

	if err := taurus.LoadFlags(&Conf); err != nil {
		log.Warn("Failed to read and merge flags", "err", err)
	}

	return nil
}

func setupTypes() {
	taurus.RegisterCustomMarshaler(func(level log.Level) ([]byte, error) {
		return []byte(level.String()), nil
	})

	taurus.RegisterCustomUnmarshaler(func(level *log.Level, data []byte) error {
		var err error
		*level, err = log.ParseLevel(string(data))
		return err
	})
	taurus.RegisterCustomUnmarshaler(func(duration *time.Duration, data []byte) error {
		var err error
		*duration, err = utils.ParseHumanDuration(string(data))
		return err
	})
}

func getConfigFilePath() (string, error) {
	if DataDir != "" {
		configFilePath, err := filepath.Abs(filepath.Join(DataDir, "config.yaml"))
		if err != nil {
			log.Warn("Failed to get absolute path for config.yaml in DataDir", "err", err)
		} else {
			err := os.MkdirAll(filepath.Dir(configFilePath), os.ModePerm)
			if err != nil {
				log.Warn("Failed to create directories for config.yaml in DataDir", "err", err)
				return "", err
			}
			return configFilePath, nil
		}
	}

	configFilePath, err := filepath.Abs("config.yaml")
	if err != nil {
		log.Warn("Failed to get absolute path for config.yaml", "err", err)
	} else if _, err := os.Stat(configFilePath); err == nil {
		return configFilePath, nil
	}

	configFilePath, err = os.UserConfigDir()
	if err != nil {
		log.Warn("Failed to get user config directory", "err", err)
	} else {
		configFilePath = filepath.Join(configFilePath, "libra", "config.yaml")
		if _, err := os.Stat(configFilePath); err == nil {
			return configFilePath, nil
		}
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if _, err := os.Stat(configFilePath); errors.Is(err, fs.ErrNotExist) {
			if _, err := os.Stat("/etc/libra/config.yaml"); err == nil {
				configFilePath = "/etc/libra/config.yaml"
			}
		}
	}

	return configFilePath, nil
}
