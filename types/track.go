package types

type Track struct {
	ID             string            `json:"id"               example:"7nTwkcl51u4"`
	UserID         string            `json:"user_id"          example:"TPkrKcIZRRq"`
	ISRC           string            `json:"isrc"             example:"USSKG1912345"`
	Title          string            `json:"title"            example:"Lorem"`
	ArtistIDs      []string          `json:"artist_ids"       example:"h3r3VpPvSq8,R2QTLKbHamW"`
	AlbumIDs       []string          `json:"album_ids"        example:"BhRpYVlrMo8,poFEUbgBuwJ"`
	PrimaryAlbumID string            `json:"primary_album_id" example:"BhRpYVlrMo8"`
	TrackNumber    int               `json:"track_number"     example:"1"`
	Duration       int               `json:"duration"         example:"300"`
	Description    string            `json:"description"      example:"Lorem ipsum dolor sit amet."`
	ReleaseDate    string            `json:"release_date"     example:"2023-10-01"`
	Lyrics         map[string]string `json:"lyrics"`
	ListenCount    int               `json:"listen_count"     example:"150"`
	FavoriteCount  int               `json:"favorite_count"   example:"5"`
	AdditionDate   int64             `json:"addition_date"    example:"1634296980"`
	Tags           []string          `json:"tags"`
	AdditionalMeta map[string]any    `json:"additional_meta"`
	Permissions    map[string]string `json:"permissions"`
	LinkedItemIDs  []string          `json:"linked_item_ids"`
	ContentSource  string            `json:"content_source"`
	MetadataSource string            `json:"metadata_source"`
	LyricSources   map[string]string `json:"lyric_sources"`
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

func (t Track) GetAdditionalMeta() map[string]any {
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
