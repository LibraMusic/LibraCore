package db

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/LibraMusic/LibraCore/config"
	"github.com/LibraMusic/LibraCore/types"
	"github.com/LibraMusic/LibraCore/utils"
)

var testCases = []struct {
	name         string
	setupFunc    func(t *testing.T) Database
	shutdownFunc func(t *testing.T, db Database) error
}{
	{"SQLite", setupSQLiteTestDB, shutdownSQLiteTestDB},
	{"PostgreSQL", setupPostgreSQLTestDB, shutdownPostgreSQLTestDB},
}

func setupSQLiteTestDB(t *testing.T) Database {
	if testing.Short() {
		t.Skip("Skipping SQLite tests in short mode")
	}

	if os.Getenv("SKIP_SQLITE_TESTS") != "" {
		t.Skip("Skipping SQLite tests")
	}

	config.Conf.Database.SQLite.Path = ":memory:"

	db, err := ConnectSQLite()
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	return db
}

func shutdownSQLiteTestDB(t *testing.T, db Database) error {
	err := db.Close()
	if err != nil {
		t.Fatalf("Failed to close SQLite connection: %v", err)
	}
	return err
}

func setupPostgreSQLTestDB(t *testing.T) Database {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL tests in short mode")
	}

	if os.Getenv("SKIP_POSTGRESQL_TESTS") != "" {
		t.Skip("Skipping PostgreSQL tests")
	}

	config.Conf.Database.PostgreSQL.Host = "localhost"
	config.Conf.Database.PostgreSQL.User = "postgres"
	config.Conf.Database.PostgreSQL.Pass = "password"
	config.Conf.Database.PostgreSQL.DBName = "libra_test_" + strings.ToLower(utils.GenerateID(6))
	config.Conf.Database.PostgreSQL.Params = "?sslmode=disable"

	db, err := createPostgreSQLDatabase()
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	return db
}

func shutdownPostgreSQLTestDB(t *testing.T, db Database) error {
	err := db.Close()
	if err != nil {
		t.Fatalf("Failed to close PostgreSQL connection: %v", err)
	}

	err = dropPostgreSQLDatabase()
	if err != nil {
		t.Fatalf("Failed to drop PostgreSQL database: %v", err)
	}

	return err
}

func TestDatabaseAddTrack(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			track := types.Track{
				ID:             "1",
				UserID:         "user1",
				ISRC:           "ISRC123",
				Title:          "Test Track",
				ArtistIDs:      []string{"artist1"},
				AlbumIDs:       []string{"album1"},
				PrimaryAlbumID: "album1",
				TrackNumber:    1,
				Duration:       180,
				Description:    "Test Description",
				ReleaseDate:    "2023-01-01",
				Lyrics:         map[string]string{"en": "Test Lyrics"},
				ListenCount:    100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				ContentSource:  "mock::1",
				MetadataSource: "mock::1",
				LyricSources:   map[string]types.LinkedSource{"en": "mock::1"},
			}

			err := db.AddTrack(track)
			assert.NoError(t, err)

			retrievedTrack, err := db.GetTrack("1")
			assert.NoError(t, err)
			assert.Equal(t, track, retrievedTrack)
		})
	}
}

func TestDatabaseGetTracks(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			track1 := types.Track{
				ID:             "1",
				UserID:         "user1",
				ISRC:           "ISRC123",
				Title:          "Test Track 1",
				ArtistIDs:      []string{"artist1"},
				AlbumIDs:       []string{"album1"},
				PrimaryAlbumID: "album1",
				TrackNumber:    1,
				Duration:       180,
				Description:    "Test Description 1",
				ReleaseDate:    "2023-01-01",
				Lyrics:         map[string]string{"en": "Test Lyrics 1"},
				ListenCount:    100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				ContentSource:  "mock::1",
				MetadataSource: "mock::1",
				LyricSources:   map[string]types.LinkedSource{"en": "mock::1"},
			}

			track2 := types.Track{
				ID:             "2",
				UserID:         "user1",
				ISRC:           "ISRC124",
				Title:          "Test Track 2",
				ArtistIDs:      []string{"artist2"},
				AlbumIDs:       []string{"album2"},
				PrimaryAlbumID: "album2",
				TrackNumber:    2,
				Duration:       200,
				Description:    "Test Description 2",
				ReleaseDate:    "2023-01-02",
				Lyrics:         map[string]string{"en": "Test Lyrics 2"},
				ListenCount:    200,
				FavoriteCount:  100,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag3", "tag4"},
				AdditionalMeta: map[string]interface{}{"key": "value2"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item2"},
				ContentSource:  "mock::2",
				MetadataSource: "mock::2",
				LyricSources:   map[string]types.LinkedSource{"en": "mock::2"},
			}

			track3 := types.Track{
				ID:             "3",
				UserID:         "user2",
				ISRC:           "ISRC125",
				Title:          "Test Track 3",
				ArtistIDs:      []string{"artist3"},
				AlbumIDs:       []string{"album3"},
				PrimaryAlbumID: "album3",
				TrackNumber:    3,
				Duration:       300,
				Description:    "Test Description 3",
				ReleaseDate:    "2023-01-03",
				Lyrics:         map[string]string{"en": "Test Lyrics 3"},
				ListenCount:    300,
				FavoriteCount:  150,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag5", "tag6"},
				AdditionalMeta: map[string]interface{}{"key": "value3"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item3"},
				ContentSource:  "mock::3",
				MetadataSource: "mock::3",
				LyricSources:   map[string]types.LinkedSource{"en": "mock::3"},
			}

			err := db.AddTrack(track1)
			assert.NoError(t, err)
			err = db.AddTrack(track2)
			assert.NoError(t, err)
			err = db.AddTrack(track3)
			assert.NoError(t, err)

			tracks, err := db.GetAllTracks()
			assert.NoError(t, err)
			assert.Len(t, tracks, 3)
			assert.Contains(t, tracks, track1)
			assert.Contains(t, tracks, track2)
			assert.Contains(t, tracks, track3)

			tracks, err = db.GetTracks("user1")
			assert.NoError(t, err)
			assert.Len(t, tracks, 2)
			assert.Contains(t, tracks, track1)
			assert.Contains(t, tracks, track2)

			tracks, err = db.GetTracks("user2")
			assert.NoError(t, err)
			assert.Len(t, tracks, 1)
			assert.Contains(t, tracks, track3)
		})
	}
}

func TestDatabaseUpdateTrack(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			track := types.Track{
				ID:             "1",
				UserID:         "user1",
				ISRC:           "ISRC123",
				Title:          "Test Track",
				ArtistIDs:      []string{"artist1"},
				AlbumIDs:       []string{"album1"},
				PrimaryAlbumID: "album1",
				TrackNumber:    1,
				Duration:       180,
				Description:    "Test Description",
				ReleaseDate:    "2023-01-01",
				Lyrics:         map[string]string{"en": "Test Lyrics"},
				ListenCount:    100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				ContentSource:  "mock::1",
				MetadataSource: "mock::1",
				LyricSources:   map[string]types.LinkedSource{"en": "mock::1"},
			}

			err := db.AddTrack(track)
			assert.NoError(t, err)

			track.Title = "Updated Test Track"
			err = db.UpdateTrack(track)
			assert.NoError(t, err)

			retrievedTrack, err := db.GetTrack("1")
			assert.NoError(t, err)
			assert.Equal(t, "Updated Test Track", retrievedTrack.Title)
		})
	}
}

func TestDatabaseDeleteTrack(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			track := types.Track{
				ID:             "1",
				UserID:         "user1",
				ISRC:           "ISRC123",
				Title:          "Test Track",
				ArtistIDs:      []string{"artist1"},
				AlbumIDs:       []string{"album1"},
				PrimaryAlbumID: "album1",
				TrackNumber:    1,
				Duration:       180,
				Description:    "Test Description",
				ReleaseDate:    "2023-01-01",
				Lyrics:         map[string]string{"en": "Test Lyrics"},
				ListenCount:    100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				ContentSource:  "mock::1",
				MetadataSource: "mock::1",
				LyricSources:   map[string]types.LinkedSource{"en": "mock::1"},
			}

			err := db.AddTrack(track)
			assert.NoError(t, err)

			err = db.DeleteTrack("1")
			assert.NoError(t, err)

			_, err = db.GetTrack("1")
			assert.Error(t, err)
		})
	}
}

func TestDatabaseAddAlbum(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			album := types.Album{
				ID:             "1",
				UserID:         "user1",
				UPC:            "UPC123",
				Title:          "Test Album",
				ArtistIDs:      []string{"artist1"},
				TrackIDs:       []string{"track1"},
				Description:    "Test Description",
				ReleaseDate:    "2023-01-01",
				ListenCount:    100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				MetadataSource: "mock::1",
			}

			err := db.AddAlbum(album)
			assert.NoError(t, err)

			retrievedAlbum, err := db.GetAlbum("1")
			assert.NoError(t, err)
			assert.Equal(t, album, retrievedAlbum)
		})
	}
}

func TestDatabaseGetAlbums(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			album1 := types.Album{
				ID:             "1",
				UserID:         "user1",
				UPC:            "UPC123",
				Title:          "Test Album 1",
				ArtistIDs:      []string{"artist1"},
				TrackIDs:       []string{"track1"},
				Description:    "Test Description 1",
				ReleaseDate:    "2023-01-01",
				ListenCount:    100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				MetadataSource: "mock::1",
			}

			album2 := types.Album{
				ID:             "2",
				UserID:         "user1",
				UPC:            "UPC124",
				Title:          "Test Album 2",
				ArtistIDs:      []string{"artist2"},
				TrackIDs:       []string{"track2"},
				Description:    "Test Description 2",
				ReleaseDate:    "2023-01-02",
				ListenCount:    200,
				FavoriteCount:  100,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag3", "tag4"},
				AdditionalMeta: map[string]interface{}{"key": "value2"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item2"},
				MetadataSource: "mock::2",
			}

			album3 := types.Album{
				ID:             "3",
				UserID:         "user2",
				UPC:            "UPC125",
				Title:          "Test Album 3",
				ArtistIDs:      []string{"artist3"},
				TrackIDs:       []string{"track3"},
				Description:    "Test Description 3",
				ReleaseDate:    "2023-01-03",
				ListenCount:    300,
				FavoriteCount:  150,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag5", "tag6"},
				AdditionalMeta: map[string]interface{}{"key": "value3"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item3"},
				MetadataSource: "mock::3",
			}

			err := db.AddAlbum(album1)
			assert.NoError(t, err)
			err = db.AddAlbum(album2)
			assert.NoError(t, err)
			err = db.AddAlbum(album3)
			assert.NoError(t, err)

			albums, err := db.GetAllAlbums()
			assert.NoError(t, err)
			assert.Len(t, albums, 3)
			assert.Contains(t, albums, album1)
			assert.Contains(t, albums, album2)
			assert.Contains(t, albums, album3)

			albums, err = db.GetAlbums("user1")
			assert.NoError(t, err)
			assert.Len(t, albums, 2)
			assert.Contains(t, albums, album1)
			assert.Contains(t, albums, album2)

			albums, err = db.GetAlbums("user2")
			assert.NoError(t, err)
			assert.Len(t, albums, 1)
			assert.Contains(t, albums, album3)
		})
	}
}

func TestDatabaseUpdateAlbum(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			album := types.Album{
				ID:             "1",
				UserID:         "user1",
				UPC:            "UPC123",
				Title:          "Test Album",
				ArtistIDs:      []string{"artist1"},
				TrackIDs:       []string{"track1"},
				Description:    "Test Description",
				ReleaseDate:    "2023-01-01",
				ListenCount:    100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				MetadataSource: "mock::1",
			}

			err := db.AddAlbum(album)
			assert.NoError(t, err)

			album.Title = "Updated Test Album"
			err = db.UpdateAlbum(album)
			assert.NoError(t, err)

			retrievedAlbum, err := db.GetAlbum("1")
			assert.NoError(t, err)
			assert.Equal(t, "Updated Test Album", retrievedAlbum.Title)
		})
	}
}

func TestDatabaseDeleteAlbum(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			album := types.Album{
				ID:             "1",
				UserID:         "user1",
				UPC:            "UPC123",
				Title:          "Test Album",
				ArtistIDs:      []string{"artist1"},
				TrackIDs:       []string{"track1"},
				Description:    "Test Description",
				ReleaseDate:    "2023-01-01",
				ListenCount:    100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				MetadataSource: "mock::1",
			}

			err := db.AddAlbum(album)
			assert.NoError(t, err)

			err = db.DeleteAlbum("1")
			assert.NoError(t, err)

			_, err = db.GetAlbum("1")
			assert.Error(t, err)
		})
	}
}

func TestDatabaseAddVideo(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			video := types.Video{
				ID:             "1",
				UserID:         "user1",
				Title:          "Test Video",
				ArtistIDs:      []string{"artist1"},
				Duration:       300,
				Description:    "Test Description",
				ReleaseDate:    "2023-01-01",
				Subtitles:      map[string]string{"en": "Test Subtitles"},
				WatchCount:     100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				ContentSource:  "mock::1",
				MetadataSource: "mock::1",
				LyricSources:   map[string]types.LinkedSource{"en": "mock::1"},
			}

			err := db.AddVideo(video)
			assert.NoError(t, err)

			retrievedVideo, err := db.GetVideo("1")
			assert.NoError(t, err)
			assert.Equal(t, video, retrievedVideo)
		})
	}
}

func TestDatabaseGetVideos(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			video1 := types.Video{
				ID:             "1",
				UserID:         "user1",
				Title:          "Test Video 1",
				ArtistIDs:      []string{"artist1"},
				Duration:       300,
				Description:    "Test Description 1",
				ReleaseDate:    "2023-01-01",
				Subtitles:      map[string]string{"en": "Test Subtitles 1"},
				WatchCount:     100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				ContentSource:  "mock::1",
				MetadataSource: "mock::1",
				LyricSources:   map[string]types.LinkedSource{"en": "mock::1"},
			}

			video2 := types.Video{
				ID:             "2",
				UserID:         "user1",
				Title:          "Test Video 2",
				ArtistIDs:      []string{"artist2"},
				Duration:       400,
				Description:    "Test Description 2",
				ReleaseDate:    "2023-01-02",
				Subtitles:      map[string]string{"en": "Test Subtitles 2"},
				WatchCount:     200,
				FavoriteCount:  100,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag3", "tag4"},
				AdditionalMeta: map[string]interface{}{"key": "value2"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item2"},
				ContentSource:  "mock::2",
				MetadataSource: "mock::2",
				LyricSources:   map[string]types.LinkedSource{"en": "mock::2"},
			}

			video3 := types.Video{
				ID:             "3",
				UserID:         "user2",
				Title:          "Test Video 3",
				ArtistIDs:      []string{"artist3"},
				Duration:       500,
				Description:    "Test Description 3",
				ReleaseDate:    "2023-01-03",
				Subtitles:      map[string]string{"en": "Test Subtitles 3"},
				WatchCount:     300,
				FavoriteCount:  150,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag5", "tag6"},
				AdditionalMeta: map[string]interface{}{"key": "value3"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item3"},
				ContentSource:  "mock::3",
				MetadataSource: "mock::3",
				LyricSources:   map[string]types.LinkedSource{"en": "mock::3"},
			}

			err := db.AddVideo(video1)
			assert.NoError(t, err)
			err = db.AddVideo(video2)
			assert.NoError(t, err)
			err = db.AddVideo(video3)
			assert.NoError(t, err)

			videos, err := db.GetAllVideos()
			assert.NoError(t, err)
			assert.Len(t, videos, 3)
			assert.Contains(t, videos, video1)
			assert.Contains(t, videos, video2)
			assert.Contains(t, videos, video3)

			videos, err = db.GetVideos("user1")
			assert.NoError(t, err)
			assert.Len(t, videos, 2)
			assert.Contains(t, videos, video1)
			assert.Contains(t, videos, video2)

			videos, err = db.GetVideos("user2")
			assert.NoError(t, err)
			assert.Len(t, videos, 1)
			assert.Contains(t, videos, video3)
		})
	}
}

func TestDatabaseUpdateVideo(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			video := types.Video{
				ID:             "1",
				UserID:         "user1",
				Title:          "Test Video",
				ArtistIDs:      []string{"artist1"},
				Duration:       300,
				Description:    "Test Description",
				ReleaseDate:    "2023-01-01",
				Subtitles:      map[string]string{"en": "Test Subtitles"},
				WatchCount:     100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				ContentSource:  "mock::1",
				MetadataSource: "mock::1",
				LyricSources:   map[string]types.LinkedSource{"en": "mock::1"},
			}

			err := db.AddVideo(video)
			assert.NoError(t, err)

			video.Title = "Updated Test Video"
			err = db.UpdateVideo(video)
			assert.NoError(t, err)

			retrievedVideo, err := db.GetVideo("1")
			assert.NoError(t, err)
			assert.Equal(t, "Updated Test Video", retrievedVideo.Title)
		})
	}
}

func TestDatabaseDeleteVideo(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			video := types.Video{
				ID:             "1",
				UserID:         "user1",
				Title:          "Test Video",
				ArtistIDs:      []string{"artist1"},
				Duration:       300,
				Description:    "Test Description",
				ReleaseDate:    "2023-01-01",
				Subtitles:      map[string]string{"en": "Test Subtitles"},
				WatchCount:     100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				ContentSource:  "mock::1",
				MetadataSource: "mock::1",
				LyricSources:   map[string]types.LinkedSource{"en": "mock::1"},
			}

			err := db.AddVideo(video)
			assert.NoError(t, err)

			err = db.DeleteVideo("1")
			assert.NoError(t, err)

			_, err = db.GetVideo("1")
			assert.Error(t, err)
		})
	}
}

func TestDatabaseAddArtist(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			artist := types.Artist{
				ID:             "1",
				UserID:         "user1",
				Name:           "Test Artist",
				AlbumIDs:       []string{"album1"},
				TrackIDs:       []string{"track1"},
				Description:    "Test Description",
				CreationDate:   "2023-01-01",
				ListenCount:    100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				MetadataSource: "mock::1",
			}

			err := db.AddArtist(artist)
			assert.NoError(t, err)

			retrievedArtist, err := db.GetArtist("1")
			assert.NoError(t, err)
			assert.Equal(t, artist, retrievedArtist)
		})
	}
}

func TestDatabaseGetArtists(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			artist1 := types.Artist{
				ID:             "1",
				UserID:         "user1",
				Name:           "Test Artist 1",
				AlbumIDs:       []string{"album1"},
				TrackIDs:       []string{"track1"},
				Description:    "Test Description 1",
				CreationDate:   "2023-01-01",
				ListenCount:    100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				MetadataSource: "mock::1",
			}

			artist2 := types.Artist{
				ID:             "2",
				UserID:         "user1",
				Name:           "Test Artist 2",
				AlbumIDs:       []string{"album2"},
				TrackIDs:       []string{"track2"},
				Description:    "Test Description 2",
				CreationDate:   "2023-01-02",
				ListenCount:    200,
				FavoriteCount:  100,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag3", "tag4"},
				AdditionalMeta: map[string]interface{}{"key": "value2"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item2"},
				MetadataSource: "mock::2",
			}

			artist3 := types.Artist{
				ID:             "3",
				UserID:         "user2",
				Name:           "Test Artist 3",
				AlbumIDs:       []string{"album3"},
				TrackIDs:       []string{"track3"},
				Description:    "Test Description 3",
				CreationDate:   "2023-01-03",
				ListenCount:    300,
				FavoriteCount:  150,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag5", "tag6"},
				AdditionalMeta: map[string]interface{}{"key": "value3"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item3"},
				MetadataSource: "mock::3",
			}

			err := db.AddArtist(artist1)
			assert.NoError(t, err)
			err = db.AddArtist(artist2)
			assert.NoError(t, err)
			err = db.AddArtist(artist3)
			assert.NoError(t, err)

			artists, err := db.GetAllArtists()
			assert.NoError(t, err)
			assert.Len(t, artists, 3)
			assert.Contains(t, artists, artist1)
			assert.Contains(t, artists, artist2)
			assert.Contains(t, artists, artist3)

			artists, err = db.GetArtists("user1")
			assert.NoError(t, err)
			assert.Len(t, artists, 2)
			assert.Contains(t, artists, artist1)
			assert.Contains(t, artists, artist2)

			artists, err = db.GetArtists("user2")
			assert.NoError(t, err)
			assert.Len(t, artists, 1)
			assert.Contains(t, artists, artist3)
		})
	}
}

func TestDatabaseUpdateArtist(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			artist := types.Artist{
				ID:             "1",
				UserID:         "user1",
				Name:           "Test Artist",
				AlbumIDs:       []string{"album1"},
				TrackIDs:       []string{"track1"},
				Description:    "Test Description",
				CreationDate:   "2023-01-01",
				ListenCount:    100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				MetadataSource: "mock::1",
			}

			err := db.AddArtist(artist)
			assert.NoError(t, err)

			artist.Name = "Updated Test Artist"
			err = db.UpdateArtist(artist)
			assert.NoError(t, err)

			retrievedArtist, err := db.GetArtist("1")
			assert.NoError(t, err)
			assert.Equal(t, "Updated Test Artist", retrievedArtist.Name)
		})
	}
}

func TestDatabaseDeleteArtist(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			artist := types.Artist{
				ID:             "1",
				UserID:         "user1",
				Name:           "Test Artist",
				AlbumIDs:       []string{"album1"},
				TrackIDs:       []string{"track1"},
				Description:    "Test Description",
				CreationDate:   "2023-01-01",
				ListenCount:    100,
				FavoriteCount:  50,
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				LinkedItemIDs:  []string{"item1"},
				MetadataSource: "mock::1",
			}

			err := db.AddArtist(artist)
			assert.NoError(t, err)

			err = db.DeleteArtist("1")
			assert.NoError(t, err)

			_, err = db.GetArtist("1")
			assert.Error(t, err)
		})
	}
}

func TestDatabaseAddPlaylist(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			playlist := types.Playlist{
				ID:             "1",
				UserID:         "user1",
				Title:          "Test Playlist",
				TrackIDs:       []string{"track1", "track2"},
				ListenCount:    100,
				FavoriteCount:  50,
				Description:    "Test Description",
				CreationDate:   "2023-01-01",
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				MetadataSource: "mock::1",
			}

			err := db.AddPlaylist(playlist)
			assert.NoError(t, err)

			retrievedPlaylist, err := db.GetPlaylist("1")
			assert.NoError(t, err)
			assert.Equal(t, playlist, retrievedPlaylist)
		})
	}
}

func TestDatabaseGetPlaylists(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			playlist1 := types.Playlist{
				ID:             "1",
				UserID:         "user1",
				Title:          "Test Playlist 1",
				TrackIDs:       []string{"track1", "track2"},
				ListenCount:    100,
				FavoriteCount:  50,
				Description:    "Test Description 1",
				CreationDate:   "2023-01-01",
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				MetadataSource: "mock::1",
			}

			playlist2 := types.Playlist{
				ID:             "2",
				UserID:         "user1",
				Title:          "Test Playlist 2",
				TrackIDs:       []string{"track3", "track4"},
				ListenCount:    200,
				FavoriteCount:  100,
				Description:    "Test Description 2",
				CreationDate:   "2023-01-02",
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag3", "tag4"},
				AdditionalMeta: map[string]interface{}{"key": "value2"},
				Permissions:    map[string]string{"read": "all"},
				MetadataSource: "mock::2",
			}

			playlist3 := types.Playlist{
				ID:             "3",
				UserID:         "user2",
				Title:          "Test Playlist 3",
				TrackIDs:       []string{"track5", "track6"},
				ListenCount:    300,
				FavoriteCount:  150,
				Description:    "Test Description 3",
				CreationDate:   "2023-01-03",
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag5", "tag6"},
				AdditionalMeta: map[string]interface{}{"key": "value3"},
				Permissions:    map[string]string{"read": "all"},
				MetadataSource: "mock::3",
			}

			err := db.AddPlaylist(playlist1)
			assert.NoError(t, err)
			err = db.AddPlaylist(playlist2)
			assert.NoError(t, err)
			err = db.AddPlaylist(playlist3)
			assert.NoError(t, err)

			playlists, err := db.GetAllPlaylists()
			assert.NoError(t, err)
			assert.Len(t, playlists, 3)
			assert.Contains(t, playlists, playlist1)
			assert.Contains(t, playlists, playlist2)
			assert.Contains(t, playlists, playlist3)

			playlists, err = db.GetPlaylists("user1")
			assert.NoError(t, err)
			assert.Len(t, playlists, 2)
			assert.Contains(t, playlists, playlist1)
			assert.Contains(t, playlists, playlist2)

			playlists, err = db.GetPlaylists("user2")
			assert.NoError(t, err)
			assert.Len(t, playlists, 1)
			assert.Contains(t, playlists, playlist3)
		})
	}
}

func TestDatabaseUpdatePlaylist(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			playlist := types.Playlist{
				ID:             "1",
				UserID:         "user1",
				Title:          "Test Playlist",
				TrackIDs:       []string{"track1", "track2"},
				ListenCount:    100,
				FavoriteCount:  50,
				Description:    "Test Description",
				CreationDate:   "2023-01-01",
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				MetadataSource: "mock::1",
			}

			err := db.AddPlaylist(playlist)
			assert.NoError(t, err)

			playlist.Title = "Updated Test Playlist"
			err = db.UpdatePlaylist(playlist)
			assert.NoError(t, err)

			retrievedPlaylist, err := db.GetPlaylist("1")
			assert.NoError(t, err)
			assert.Equal(t, "Updated Test Playlist", retrievedPlaylist.Title)
		})
	}
}

func TestDatabaseDeletePlaylist(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			playlist := types.Playlist{
				ID:             "1",
				UserID:         "user1",
				Title:          "Test Playlist",
				TrackIDs:       []string{"track1", "track2"},
				ListenCount:    100,
				FavoriteCount:  50,
				Description:    "Test Description",
				CreationDate:   "2023-01-01",
				AdditionDate:   time.Now().Unix(),
				Tags:           []string{"tag1", "tag2"},
				AdditionalMeta: map[string]interface{}{"key": "value"},
				Permissions:    map[string]string{"read": "all"},
				MetadataSource: "mock::1",
			}

			err := db.AddPlaylist(playlist)
			assert.NoError(t, err)

			err = db.DeletePlaylist("1")
			assert.NoError(t, err)

			_, err = db.GetPlaylist("1")
			assert.Error(t, err)
		})
	}
}

func TestDatabaseAddUser(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			user := types.User{
				ID:              "1",
				Username:        "testuser",
				Email:           "testuser@example.com",
				PasswordHash:    "hashedpassword",
				DisplayName:     "Test User",
				Description:     "Test Description",
				ListenedTo:      map[string]int{"track1": 1},
				Favorites:       []string{"track1", "track2"},
				PublicViewCount: 100,
				CreationDate:    time.Now().Unix(),
				Permissions:     map[string]string{"admin": "true"},
				LinkedArtistID:  "artist1",
				LinkedSources:   map[string]string{"source1": "link1"},
			}

			err := db.CreateUser(user)
			assert.NoError(t, err)

			retrievedUser, err := db.GetUser("1")
			assert.NoError(t, err)
			assert.Equal(t, user, retrievedUser)
		})
	}
}

func TestDatabaseGetUsers(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			user1 := types.User{
				ID:              "1",
				Username:        "testuser1",
				Email:           "testuser1@example.com",
				PasswordHash:    "hashedpassword1",
				DisplayName:     "Test User 1",
				Description:     "Test Description 1",
				ListenedTo:      map[string]int{"track1": 1},
				Favorites:       []string{"track1", "track2"},
				PublicViewCount: 100,
				CreationDate:    time.Now().Unix(),
				Permissions:     map[string]string{"admin": "true"},
				LinkedArtistID:  "artist1",
				LinkedSources:   map[string]string{"source1": "link1"},
			}

			user2 := types.User{
				ID:              "2",
				Username:        "testuser2",
				Email:           "testuser2@example.com",
				PasswordHash:    "hashedpassword2",
				DisplayName:     "Test User 2",
				Description:     "Test Description 2",
				ListenedTo:      map[string]int{"track2": 2},
				Favorites:       []string{"track3", "track4"},
				PublicViewCount: 200,
				CreationDate:    time.Now().Unix(),
				Permissions:     map[string]string{"admin": "false"},
				LinkedArtistID:  "artist2",
				LinkedSources:   map[string]string{"source2": "link2"},
			}

			err := db.CreateUser(user1)
			assert.NoError(t, err)
			err = db.CreateUser(user2)
			assert.NoError(t, err)

			users, err := db.GetUsers()
			assert.NoError(t, err)
			assert.Len(t, users, 2)
			assert.Contains(t, users, user1)
			assert.Contains(t, users, user2)
		})
	}
}

func TestDatabaseUpdateUser(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			user := types.User{
				ID:              "1",
				Username:        "testuser",
				Email:           "testuser@example.com",
				PasswordHash:    "hashedpassword",
				DisplayName:     "Test User",
				Description:     "Test Description",
				ListenedTo:      map[string]int{"track1": 1},
				Favorites:       []string{"track1", "track2"},
				PublicViewCount: 100,
				CreationDate:    time.Now().Unix(),
				Permissions:     map[string]string{"admin": "true"},
				LinkedArtistID:  "artist1",
				LinkedSources:   map[string]string{"source1": "link1"},
			}

			err := db.CreateUser(user)
			assert.NoError(t, err)

			user.DisplayName = "Updated Test User"
			err = db.UpdateUser(user)
			assert.NoError(t, err)

			retrievedUser, err := db.GetUser("1")
			assert.NoError(t, err)
			assert.Equal(t, "Updated Test User", retrievedUser.DisplayName)
		})
	}
}

func TestDatabaseDeleteUser(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			user := types.User{
				ID:              "1",
				Username:        "testuser",
				Email:           "testuser@example.com",
				PasswordHash:    "hashedpassword",
				DisplayName:     "Test User",
				Description:     "Test Description",
				ListenedTo:      map[string]int{"track1": 1},
				Favorites:       []string{"track1", "track2"},
				PublicViewCount: 100,
				CreationDate:    time.Now().Unix(),
				Permissions:     map[string]string{"admin": "true"},
				LinkedArtistID:  "artist1",
				LinkedSources:   map[string]string{"source1": "link1"},
			}

			err := db.CreateUser(user)
			assert.NoError(t, err)

			err = db.DeleteUser("1")
			assert.NoError(t, err)

			_, err = db.GetUser("1")
			assert.Error(t, err)
		})
	}
}

func TestDatabaseUsernameExists(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			user := types.User{
				ID:              "1",
				Username:        "testuser",
				Email:           "testuser@example.com",
				PasswordHash:    "hashedpassword",
				DisplayName:     "Test User",
				Description:     "Test Description",
				ListenedTo:      map[string]int{"track1": 1},
				Favorites:       []string{"track1", "track2"},
				PublicViewCount: 100,
				CreationDate:    time.Now().Unix(),
				Permissions:     map[string]string{"admin": "true"},
				LinkedArtistID:  "artist1",
				LinkedSources:   map[string]string{"source1": "link1"},
			}

			err := db.CreateUser(user)
			assert.NoError(t, err)

			exists, err := db.UsernameExists("testuser")
			assert.NoError(t, err)
			assert.True(t, exists)

			exists, err = db.UsernameExists("nonexistentuser")
			assert.NoError(t, err)
			assert.False(t, exists)
		})
	}
}

func TestDatabaseEmailExists(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			user := types.User{
				ID:              "1",
				Username:        "testuser",
				Email:           "testuser@example.com",
				PasswordHash:    "hashedpassword",
				DisplayName:     "Test User",
				Description:     "Test Description",
				ListenedTo:      map[string]int{"track1": 1},
				Favorites:       []string{"track1", "track2"},
				PublicViewCount: 100,
				CreationDate:    time.Now().Unix(),
				Permissions:     map[string]string{"admin": "true"},
				LinkedArtistID:  "artist1",
				LinkedSources:   map[string]string{"source1": "link1"},
			}

			err := db.CreateUser(user)
			assert.NoError(t, err)

			exists, err := db.EmailExists("testuser@example.com")
			assert.NoError(t, err)
			assert.True(t, exists)

			exists, err = db.EmailExists("nonexistent@example.com")
			assert.NoError(t, err)
			assert.False(t, exists)
		})
	}
}

func TestDatabaseBlacklistToken(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			token := "testtoken"
			expiration := time.Now().Add(1 * time.Hour)

			err := db.BlacklistToken(token, expiration)
			assert.NoError(t, err)

			isBlacklisted, err := db.IsTokenBlacklisted(token)
			assert.NoError(t, err)
			assert.True(t, isBlacklisted)
		})
	}
}

func TestDatabaseCleanExpiredTokens(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			token := "expiredtoken"
			expiration := time.Now().Add(-1 * time.Hour)

			err := db.BlacklistToken(token, expiration)
			assert.NoError(t, err)

			err = db.CleanExpiredTokens()
			assert.NoError(t, err)

			isBlacklisted, err := db.IsTokenBlacklisted(token)
			assert.NoError(t, err)
			assert.False(t, isBlacklisted)
		})
	}
}

func TestDatabaseIsTokenBlacklisted(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setupFunc(t)
			assert.NotNil(t, db)
			defer tc.shutdownFunc(t, db)

			token := "testtoken"
			expiration := time.Now().Add(1 * time.Hour)

			err := db.BlacklistToken(token, expiration)
			assert.NoError(t, err)

			isBlacklisted, err := db.IsTokenBlacklisted(token)
			assert.NoError(t, err)
			assert.True(t, isBlacklisted)

			nonExistentToken := "nonexistenttoken"
			isBlacklisted, err = db.IsTokenBlacklisted(nonExistentToken)
			assert.NoError(t, err)
			assert.False(t, isBlacklisted)
		})
	}
}
