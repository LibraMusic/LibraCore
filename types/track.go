package types

type Track struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	ISRC           string                 `json:"isrc"`
	Title          string                 `json:"title"`
	ArtistIDs      []string               `json:"artist_ids"`
	AlbumIDs       []string               `json:"album_ids"`
	PrimaryAlbumID string                 `json:"primary_album_id"`
	TrackNumber    int                    `json:"track_number"`
	Duration       int                    `json:"duration"`
	Description    string                 `json:"description"`
	ReleaseDate    string                 `json:"release_date"`
	Lyrics         map[string]string      `json:"lyrics"`
	ListenCount    int                    `json:"listen_count"`
	FavoriteCount  int                    `json:"favorite_count"`
	AdditionDate   int64                  `json:"addition_date"`
	Tags           []string               `json:"tags"`
	AdditionalMeta map[string]interface{} `json:"additional_meta"`
	Permissions    map[string]string      `json:"permissions"`
	LinkedItemIDs  []string               `json:"linked_item_ids"`
	ContentSource  string                 `json:"content_source"`
	MetadataSource string                 `json:"metadata_source"`
	LyricSources   map[string]string      `json:"lyric_sources"`
}

func (Track) GetType() string {
	return "track"
}

func (t Track) GetID() string {
	return t.ID
}

func (t Track) GetUserID() string {
	return t.UserID
}

func (t Track) GetTitle() string {
	return t.Title
}

func (t Track) GetDescription() string {
	return t.Description
}

func (t Track) GetReleaseDate() string {
	return t.ReleaseDate
}

func (t Track) GetAdditionDate() int64 {
	return t.AdditionDate
}

func (t Track) GetTags() []string {
	return t.Tags
}

func (t Track) GetAdditionalMeta() map[string]interface{} {
	return t.AdditionalMeta
}

func (t Track) GetPermissions() map[string]string {
	return t.Permissions
}

func (t Track) IsTemporary() bool {
	return t.ID == ""
}

func (t Track) GetLinkedItemIDs() []string {
	return t.LinkedItemIDs
}

func (t Track) GetViewCount() int {
	return t.ListenCount
}

func (t Track) GetMetadataSource() string {
	return t.MetadataSource
}

func (t Track) GetLyrics() map[string]string {
	return t.Lyrics
}

func (t Track) GetLyricSources() map[string]string {
	return t.LyricSources
}
