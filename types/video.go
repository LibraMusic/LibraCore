package types

type Video struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	Title          string                 `json:"title"`
	ArtistIDs      []string               `json:"artist_ids"`
	Duration       int                    `json:"duration"`
	Description    string                 `json:"description"`
	ReleaseDate    string                 `json:"release_date"`
	Subtitles      map[string]string      `json:"subtitles"`
	WatchCount     int                    `json:"watch_count"`
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

func (Video) GetType() string {
	return "video"
}

func (v Video) GetID() string {
	return v.ID
}

func (v Video) GetUserID() string {
	return v.UserID
}

func (v Video) GetTitle() string {
	return v.Title
}

func (v Video) GetDescription() string {
	return v.Description
}

func (v Video) GetReleaseDate() string {
	return v.ReleaseDate
}

func (v Video) GetAdditionDate() int64 {
	return v.AdditionDate
}

func (v Video) GetTags() []string {
	return v.Tags
}

func (v Video) GetAdditionalMeta() map[string]interface{} {
	return v.AdditionalMeta
}

func (v Video) GetPermissions() map[string]string {
	return v.Permissions
}

func (v Video) IsTemporary() bool {
	return v.ID == ""
}

func (v Video) GetLinkedItemIDs() []string {
	return v.LinkedItemIDs
}

func (v Video) GetViewCount() int {
	return v.WatchCount
}

func (v Video) GetMetadataSource() string {
	return v.MetadataSource
}

func (v Video) GetLyrics() map[string]string {
	return v.Subtitles
}

func (v Video) GetLyricSources() map[string]string {
	return v.LyricSources
}
