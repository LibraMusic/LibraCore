package types

type Playlist struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	Title          string                 `json:"title"`
	TrackIDs       []string               `json:"track_ids"`
	ListenCount    int                    `json:"listen_count"`
	FavoriteCount  int                    `json:"favorite_count"`
	Description    string                 `json:"description"`
	CreationDate   string                 `json:"creation_date"`
	AdditionDate   int64                  `json:"addition_date"`
	Tags           []string               `json:"tags"`
	AdditionalMeta map[string]interface{} `json:"additional_meta"`
	Permissions    map[string]string      `json:"permissions"`
	MetadataSource string                 `json:"metadata_source"`
}

func (Playlist) GetType() string {
	return "playlist"
}

func (p Playlist) GetID() string {
	return p.ID
}

func (p Playlist) GetUserID() string {
	return p.UserID
}

func (p Playlist) GetTitle() string {
	return p.Title
}

func (p Playlist) GetDescription() string {
	return p.Description
}

func (p Playlist) GetReleaseDate() string {
	return p.CreationDate
}

func (p Playlist) GetAdditionDate() int64 {
	return p.AdditionDate
}

func (p Playlist) GetTags() []string {
	return p.Tags
}

func (p Playlist) GetAdditionalMeta() map[string]interface{} {
	return p.AdditionalMeta
}

func (p Playlist) GetPermissions() map[string]string {
	return p.Permissions
}

func (p Playlist) IsTemporary() bool {
	return p.ID == ""
}

func (p Playlist) GetViewCount() int {
	return p.ListenCount
}

func (p Playlist) GetMetadataSource() string {
	return p.MetadataSource
}
