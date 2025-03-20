package types

type Video struct {
	ID             string            `json:"id"              example:"hCNchWdmbro"`
	UserID         string            `json:"user_id"         example:"TPkrKcIZRRq"`
	Title          string            `json:"title"           example:"Dolor Sit Amet"`
	ArtistIDs      []string          `json:"artist_ids"      example:"h3r3VpPvSq8,R2QTLKbHamW"`
	Duration       int               `json:"duration"        example:"300"`
	Description    string            `json:"description"     example:"Lorem ipsum dolor sit amet."`
	ReleaseDate    string            `json:"release_date"    example:"2023-10-01"`
	Subtitles      map[string]string `json:"subtitles"`
	WatchCount     int               `json:"watch_count"     example:"185"`
	FavoriteCount  int               `json:"favorite_count"  example:"10"`
	AdditionDate   int64             `json:"addition_date"`
	Tags           []string          `json:"tags"`
	AdditionalMeta map[string]any    `json:"additional_meta"`
	Permissions    map[string]string `json:"permissions"`
	LinkedItemIDs  []string          `json:"linked_item_ids"`
	ContentSource  string            `json:"content_source"`
	MetadataSource string            `json:"metadata_source"`
	LyricSources   map[string]string `json:"lyric_sources"`
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

func (v Video) GetAdditionalMeta() map[string]any {
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
