package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/c2h5oh/datasize"
	"github.com/go-viper/mapstructure/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/DevReaper0/libra/types"
	"github.com/DevReaper0/libra/util"
	// "github.com/DevReaper0/libra/util"
)

var Conf Config

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
	YouTubeURL      string `mapstructure:"youtube_url"`
}

type LogsConfig struct {
	Debug          bool `mapstructure:"debug"`
	ErrorWarnings  bool `mapstructure:"error_warnings"`
	AllErrorsFatal bool `mapstructure:"all_errors_fatal"`
}

type StorageConfig struct {
	Location            string            `mapstructure:"location"`
	SizeLimit           datasize.ByteSize `mapstructure:"size_limit"`
	MinimumAgeThreshold time.Duration     `mapstructure:"minimum_age_threshold"`
}

type PostgreSQLDatabaseConfig struct {
	Host   string `mapstructure:"host"`
	User   string `mapstructure:"user"`
	Pass   string `mapstructure:"pass"`
	DBName string `mapstructure:"db_name"`
	Params string `mapstructure:"params"`
}

type DatabaseConfig struct {
	Engine     string                   `mapstructure:"engine"`
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

var defaultConfig = `application:
    public_url: http://127.0.0.1
    port: 8080
    source_id: libra
    source_name: Libra
    media_types:
        - music
        - video
        - playlist
auth:
    jwt_signing_method: HS256 # Supported algorithms: HS256, HS384, HS512, RS256, RS384, RS512, PS256, PS384, PS512, ES256, ES384, ES512, EdDSA
    jwt_signing_key: secret # The secret key used to sign JWTs. In algorithms that use public/private keys, this should be the private key. Change this to a secure value. File paths are also supported with the format "file:<path>".
    jwt_refresh_token_expiration: 30d # Default is 30 days.
    jwt_access_token_expiration: 15m # Default is 15 minutes.
    global_api_routes_require_auth: true # Determines if "global" API routes (e.g., /api/v1/tracks) require authentication.
    user_api_routes_require_auth: true # Determines if "user" API routes (e.g., /api/v1/tracks/:id) require authentication.
    user_api_require_same_user: false # Determines if "user" API routes require that the authenticated user is the same as the user in the route. Disabled by default because enabling it disables song sharing and other interactions between users.
general:
    id_length: 7 # Adjust this value if there (somehow) aren't enough unique IDs. The calculation is 62^id_length, so a value of 7 gives 3.5 trillion unique IDs.
    include_video_results: true # Determines whether video results are included in search results from sources that support video.
    video_audio_only: true # If true, only the audio of videos is downloaded and they are treated as tracks.
    inherit_listen_counts: false # If true, the listen count for newly added content is initially set to the listen count from the source. If false, the listen count is initially set to 0.
    artist_listen_counts_by_track: true # If true, the listen count for an artist is the sum of the listen counts for all of their content. If false, the listen count for an artist is the number of times a user has played the artist as a whole (i.e., pressing "Play" on the artist page).
    user_artist_linking: true # If true, users can link their accounts to artists. A linked account will allow the user to view analytics for the artist.
    max_search_results: 20
    max_track_duration: 0s # A value of 0 means no limit.
    reserved_usernames: # These usernames cannot be used by users. In addition to these, the username "default" is always reserved.
        - owner
        - admin
    custom_display_names: true # If true, users can set a custom display name. If false, the display name is locked to only case changes.
    reserve_display_names: true # If true, display names follow the same rules as usernames.
    admin_permissions: {}
    enabled_sources:
        - spotify
        - youtube
source_scripts:
    python_command: python3
    youtube_location: source_scripts/youtube.py
    youtube_url: https://raw.githubusercontent.com/DevReaper0/libra/main/source_scripts/youtube.py
logs:
    debug: false
    error_warnings: false
    all_errors_fatal: false
storage:
    location: ./storage
    size_limit: 0B # A value of 0 means no limit.
    minimum_age_threshold: 1w # The minimum age of a file before it can be deleted when the storage size limit is reached. A value of 0 means no minimum age. Default is 1 week.
database:
    engine: postgresql
    postgresql:
        host: localhost
        user: libra
        pass: password
        db_name: libra
        params: sslmode=disable`

func LoadConfig() (conf Config) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := readDefaultConfig(); err != nil {
		log.Fatal().Msg("Failed to read default config")
		return
	}

	if err := mergeConfig(); err != nil {
		log.Warn().Err(err).Msg("Failed to read config")
		log.Warn().Msg("Using default config")
	}

	if err := unmarshalConfig(&conf); err != nil {
		log.Fatal().Err(err).Msg("Failed to unmarshal config")
	}

	if strings.TrimSpace(conf.Database.PostgreSQL.Params) != "" && !strings.HasPrefix(conf.Database.PostgreSQL.Params, "?") {
		conf.Database.PostgreSQL.Params = "?" + conf.Database.PostgreSQL.Params
	}
	return
}

func readDefaultConfig() error {
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
			viper.SetConfigFile(configFilePath)
			err = viper.MergeInConfig()
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				if err := os.WriteFile(configFilePath, []byte(defaultConfig), 0644); err != nil {
					log.Warn().Err(err).Msg("Failed to write default config")
				}
			}
			err = viper.MergeInConfig()
		}
	}
	return err
}

func getConfigFilePath() (string, error) {
	configFilePath, err := xdg.ConfigFile("libra/config.yaml")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get config file path")
		configFilePath, err = filepath.Abs("config.yaml")
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get config file path")
			return "", err
		}
	}
	return configFilePath, nil
}

func unmarshalConfig(conf *Config) error {
	return viper.Unmarshal(conf, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			func(from reflect.Kind, to reflect.Kind, data interface{}) (interface{}, error) {
				if from == reflect.String && to == reflect.TypeFor[time.Duration]().Kind() {
					return util.ParseHumanDuration(data.(string))
				}
				if from == reflect.String && to == reflect.TypeFor[datasize.ByteSize]().Kind() {
					return datasize.ParseString(data.(string))
				}
				return data, nil
			},
		),
	))
}
