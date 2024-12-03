package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/c2h5oh/datasize"
	"github.com/charmbracelet/log"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"

	"github.com/LibraMusic/LibraCore/types"
	"github.com/LibraMusic/LibraCore/utils"
)

var Conf Config

//go:embed default_config.yaml
var defaultConfig string

type ApplicationConfig struct {
	PublicURL  string   `mapstructure:"public_url"`
	Port       int      `mapstructure:"port"`
	SourceID   string   `mapstructure:"source_id"`
	SourceName string   `mapstructure:"source_name"`
	MediaTypes []string `mapstructure:"media_types"`
}

type AuthConfig struct {
	JWTSigningMethod           string        `mapstructure:"jwt_signing_method"`
	JWTSigningKey              string        `mapstructure:"jwt_signing_key"`
	JWTRefreshTokenExpiration  time.Duration `mapstructure:"jwt_refresh_token_expiration"`
	JWTAccessTokenExpiration   time.Duration `mapstructure:"jwt_access_token_expiration"`
	GlobalAPIRoutesRequireAuth bool          `mapstructure:"global_api_routes_require_auth"`
	UserAPIRoutesRequireAuth   bool          `mapstructure:"user_api_routes_require_auth"`
	UserAPIRequireSameUseUser  bool          `mapstructure:"user_api_require_same_user"`
}

type GeneralConfig struct {
	IDLength                  int                               `mapstructure:"id_length"`
	IncludeVideoResults       bool                              `mapstructure:"include_video_results"`
	VideoAudioOnly            bool                              `mapstructure:"video_audio_only"`
	InheritListenCounts       bool                              `mapstructure:"inherit_listen_counts"`
	ArtistListenCountsByTrack bool                              `mapstructure:"artist_listen_counts_by_track"`
	UserArtistLinking         bool                              `mapstructure:"user_artist_linking"`
	MaxSearchResults          int                               `mapstructure:"max_search_results"`
	MaxTrackDuration          time.Duration                     `mapstructure:"max_track_duration"`
	ReservedUsernames         []string                          `mapstructure:"reserved_usernames"`
	CustomDisplayNames        bool                              `mapstructure:"custom_display_names"`
	ReserveDisplayNames       bool                              `mapstructure:"reserve_display_names"`
	AdminPermissions          map[string]types.AdminPermissions `mapstructure:"admin_permissions"`
	EnabledSources            []string                          `mapstructure:"enabled_sources"`
}

type SourceScriptsConfig struct {
	PythonCommand   string `mapstructure:"python_command"`
	YouTubeLocation string `mapstructure:"youtube_location"`
}

type LogsConfig struct {
	LogLevel  log.Level `mapstructure:"log_level"`
	LogFormat string    `mapstructure:"log_format"`
}

type StorageConfig struct {
	Location            string            `mapstructure:"location"`
	SizeLimit           datasize.ByteSize `mapstructure:"size_limit"`
	MinimumAgeThreshold time.Duration     `mapstructure:"minimum_age_threshold"`
}

type SQLiteDatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type PostgreSQLDatabaseConfig struct {
	Host   string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	User   string `mapstructure:"user"`
	Pass   string `mapstructure:"pass"`
	DBName string `mapstructure:"db_name"`
	Params string `mapstructure:"params"`
}

type DatabaseConfig struct {
	Engine     string                   `mapstructure:"engine"`
	SQLite     SQLiteDatabaseConfig     `mapstructure:"sqlite"`
	PostgreSQL PostgreSQLDatabaseConfig `mapstructure:"postgresql"`
}

type Config struct {
	Application   ApplicationConfig   `mapstructure:"application"`
	Auth          AuthConfig          `mapstructure:"auth"`
	General       GeneralConfig       `mapstructure:"general"`
	SourceScripts SourceScriptsConfig `mapstructure:"source_scripts"`
	Logs          LogsConfig          `mapstructure:"logs"`
	Storage       StorageConfig       `mapstructure:"storage"`
	Database      DatabaseConfig      `mapstructure:"database"`
}

func LoadConfig() (Config, error) {
	var conf Config

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("LIBRA")
	viper.AutomaticEnv()

	if err := loadDefaultConfig(); err != nil {
		return conf, fmt.Errorf("failed to read default config: %w", err)
	}

	if err := mergeConfig(); err != nil {
		log.Warn("Failed to read config", "err", err)
		log.Warn("Using default config")
	}

	if err := unmarshalConfig(&conf); err != nil {
		return conf, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return conf, nil
}

func loadDefaultConfig() error {
	return viper.ReadConfig(strings.NewReader(defaultConfig))
}

func mergeConfig() error {
	err := viper.MergeInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		var configFilePath string
		configFilePath, err = getConfigFilePath()
		if err != nil {
			return err
		}

		if configFilePath != "" {
			os.MkdirAll(filepath.Dir(configFilePath), os.ModePerm)
			viper.SetConfigFile(configFilePath)
			err = viper.MergeInConfig()
			if _, ok := err.(viper.ConfigFileNotFoundError); ok || os.IsNotExist(err) {
				if err := os.WriteFile(configFilePath, []byte(defaultConfig), 0644); err != nil {
					log.Warn("Failed to write default config", "err", err)
				}
			}
			if err != nil {
				err = viper.MergeInConfig()
			}
		}
	}
	return err
}

func getConfigFilePath() (string, error) {
	if utils.DataDir != "" {
		configFilePath, err := filepath.Abs(filepath.Join(utils.DataDir, "config.yaml"))
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

	configFilePath, err = xdg.ConfigFile("libra/config.yaml")
	if err != nil {
		log.Warn("Failed to get config file path from XDG config directory", "err", err)
		return "", err
	}

	return configFilePath, nil
}

func unmarshalConfig(conf *Config) error {
	err := viper.Unmarshal(conf, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			func(from reflect.Kind, to reflect.Kind, data interface{}) (interface{}, error) {
				if from == reflect.String && to == reflect.TypeFor[time.Duration]().Kind() {
					return utils.ParseHumanDuration(data.(string))
				}
				if from == reflect.String && to == reflect.TypeFor[datasize.ByteSize]().Kind() {
					return datasize.ParseString(data.(string))
				}
				if from == reflect.String && to == reflect.TypeFor[log.Level]().Kind() {
					return log.ParseLevel(data.(string))
				}
				return data, nil
			},
		),
	))

	return err
}
