package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/c2h5oh/datasize"
	"github.com/charmbracelet/log"
	"github.com/goccy/go-yaml"
	"github.com/spf13/pflag"

	"github.com/LibraMusic/LibraCore/types"
	"github.com/LibraMusic/LibraCore/utils"
)

var Conf Config

//go:embed default_config.yaml
var defaultConfig string

const EnvPrefix = "LIBRA"

var envAliases = map[string][]string{
	"APPLICATION_PORT": {"PORT"},
}
var flags = make(map[string]*pflag.Flag)

type ApplicationConfig struct {
	PublicURL  string   `yaml:"public_url"`
	Port       int      `yaml:"port"`
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

type AuthConfig struct {
	JWT                        JWTAuthConfig `yaml:"jwt"`
	GlobalAPIRoutesRequireAuth bool          `yaml:"global_api_routes_require_auth"`
	UserAPIRoutesRequireAuth   bool          `yaml:"user_api_routes_require_auth"`
	UserAPIRequireSameUseUser  bool          `yaml:"user_api_require_same_user"`
	DisableAccountCreation     bool          `yaml:"disable_account_creation"`
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
	LogLevel  log.Level `yaml:"log_level"`
	LogFormat string    `yaml:"log_format"`
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
	setupYAML()

	if err := yaml.Unmarshal([]byte(defaultConfig), &Conf); err != nil {
		return fmt.Errorf("failed to read default config: %w", err)
	}

	if err := mergeFile(); err != nil {
		log.Warn("Failed to read config", "err", err)
		log.Info("Using default config")
	}

	if err := mergeEnv(EnvPrefix, &Conf); err != nil {
		log.Warn("Failed to read and merge environment variables", "err", err)
	}

	if err := mergeFlags("", &Conf); err != nil {
		log.Warn("Failed to read and merge flags", "err", err)
	}

	return nil
}

func BindFlag(fieldPath string, flag *pflag.Flag) {
	flags[fieldPath] = flag
}

func setupYAML() {
	yaml.RegisterCustomMarshaler[log.Level](func(level log.Level) ([]byte, error) {
		return []byte(level.String()), nil
	})

	yaml.RegisterCustomUnmarshaler[log.Level](func(level *log.Level, data []byte) error {
		var err error
		*level, err = log.ParseLevel(string(data))
		return err
	})
	yaml.RegisterCustomUnmarshaler[time.Duration](func(duration *time.Duration, data []byte) error {
		var err error
		*duration, err = utils.ParseHumanDuration(string(data))
		return err
	})
}

func mergeFile() error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if err := os.WriteFile(configFilePath, []byte(defaultConfig), 0o644); err != nil {
			log.Warn("Failed to write default config", "err", err)
		}
	}

	configFileContent, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	if err := yaml.Unmarshal(configFileContent, &Conf); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}
	return nil
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

func mergeEnv(prefix string, cfg interface{}) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("cfg must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Get the `yaml` tag or default to field name
		tag := field.Tag.Get("yaml")
		if tag == "" {
			tag = field.Name
		}
		envKey := prefix + "_" + strings.ToUpper(tag)

		if field.Type.Kind() == reflect.Struct {
			if err := mergeEnv(envKey, fieldValue.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		if !fieldValue.CanSet() {
			continue
		}

		envVal, exists := os.LookupEnv(envKey)
		if !exists {
			for _, key := range envAliases[strings.TrimPrefix(envKey, EnvPrefix+"_")] {
				envVal, exists = os.LookupEnv(EnvPrefix + "_" + key)
				if exists {
					envKey = EnvPrefix + "_" + key
					break
				}
			}
		}
		if !exists {
			continue
		}

		if err := setField(field, fieldValue, envVal); err != nil {
			return fmt.Errorf("error setting field for %s: %v", envKey, err)
		}
	}
	return nil
}

func mergeFlags(prefix string, cfg interface{}) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("cfg must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		fieldPath := field.Name
		if prefix != "" {
			fieldPath = prefix + "." + fieldPath
		}

		if field.Type.Kind() == reflect.Struct {
			if err := mergeFlags(fieldPath, fieldValue.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		if !fieldValue.CanSet() {
			continue
		}

		fl, exists := flags[fieldPath]
		if !exists || !fl.Changed {
			continue
		}

		if err := setField(field, fieldValue, fl.Value.String()); err != nil {
			return fmt.Errorf("error setting field for %s: %v", fieldPath, err)
		}
	}
	return nil
}

func setField(field reflect.StructField, fieldValue reflect.Value, value string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type == reflect.TypeFor[time.Duration]() {
			dur, err := utils.ParseHumanDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration: %v", err)
			}
			fieldValue.Set(reflect.ValueOf(dur))
		} else if field.Type == reflect.TypeFor[log.Level]() {
			level, err := log.ParseLevel(value)
			if err != nil {
				return fmt.Errorf("invalid log level: %v", err)
			}
			fieldValue.Set(reflect.ValueOf(level))
		} else {
			val, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int: %v", err)
			}
			fieldValue.SetInt(val)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Type == reflect.TypeFor[datasize.ByteSize]() {
			size, err := datasize.ParseString(value)
			if err != nil {
				return fmt.Errorf("invalid datasize: %v", err)
			}
			fieldValue.Set(reflect.ValueOf(size))
		} else {
			val, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid uint: %v", err)
			}
			fieldValue.SetUint(val)
		}
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %v", err)
		}
		fieldValue.SetFloat(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool: %v", err)
		}
		fieldValue.SetBool(val)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Type.Kind())
	}
	return nil
}
