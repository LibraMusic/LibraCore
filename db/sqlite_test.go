package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/LibraMusic/LibraCore/types"
)

func setupSQLiteTestDB(t *testing.T) *SQLiteDatabase {
	db, err := ConnectSQLite()
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	return db
}

func teardownSQLiteTestDB(db *SQLiteDatabase) {
	db.Close()
}

func TestSQLiteDatabase_AddTrack(t *testing.T) {
	db := setupSQLiteTestDB(t)
	defer teardownSQLiteTestDB(db)

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
		ContentSource:  "source",
		MetadataSource: "metadata",
		LyricSources:   map[string]types.LinkedSource{},
	}

	err := db.AddTrack(track)
	assert.NoError(t, err)

	retrievedTrack, err := db.GetTrack("1")
	assert.NoError(t, err)
	assert.Equal(t, track, retrievedTrack)
}

func TestSQLiteDatabase_GetTracks(t *testing.T) {
	db := setupSQLiteTestDB(t)
	defer teardownSQLiteTestDB(db)

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
		ContentSource:  "source",
		MetadataSource: "metadata",
		LyricSources:   map[string]types.LinkedSource{},
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
		ContentSource:  "source2",
		MetadataSource: "metadata2",
		LyricSources:   map[string]types.LinkedSource{},
	}

	err := db.AddTrack(track1)
	assert.NoError(t, err)
	err = db.AddTrack(track2)
	assert.NoError(t, err)

	tracks, err := db.GetTracks("user1")
	assert.NoError(t, err)
	assert.Len(t, tracks, 2)
	assert.Contains(t, tracks, track1)
	assert.Contains(t, tracks, track2)
}

func TestSQLiteDatabase_UpdateTrack(t *testing.T) {
	db := setupSQLiteTestDB(t)
	defer teardownSQLiteTestDB(db)

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
		ContentSource:  "source",
		MetadataSource: "metadata",
		LyricSources:   map[string]types.LinkedSource{},
	}

	err := db.AddTrack(track)
	assert.NoError(t, err)

	track.Title = "Updated Test Track"
	err = db.UpdateTrack(track)
	assert.NoError(t, err)

	retrievedTrack, err := db.GetTrack("1")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Test Track", retrievedTrack.Title)
}

func TestSQLiteDatabase_DeleteTrack(t *testing.T) {
	db := setupSQLiteTestDB(t)
	defer teardownSQLiteTestDB(db)

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
		ContentSource:  "source",
		MetadataSource: "metadata",
		LyricSources:   map[string]types.LinkedSource{},
	}

	err := db.AddTrack(track)
	assert.NoError(t, err)

	err = db.DeleteTrack("1")
	assert.NoError(t, err)

	_, err = db.GetTrack("1")
	assert.Error(t, err)
}
