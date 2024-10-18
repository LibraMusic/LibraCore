package types

type Artist struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	Name           string                 `json:"name"`
	AlbumIDs       []string               `json:"album_ids"`
	TrackIDs       []string               `json:"track_ids"`
	Description    string                 `json:"description"`
	CreationDate   string                 `json:"creation_date"`
	ListenCount    int                    `json:"listen_count"`
	FavoriteCount  int                    `json:"favorite_count"`
	AdditionDate   int64                  `json:"addition_date"`
	Tags           []string               `json:"tags"`
	AdditionalMeta map[string]interface{} `json:"additional_meta"`
	Permissions    map[string]string      `json:"permissions"`
	LinkedItemIDs  []string               `json:"linked_item_ids"`
	MetadataSource LinkedSource           `json:"metadata_source"`
}

func (a Artist) GetType() string {
	return "artist"
}

func (a Artist) GetID() string {
	return a.ID
}

func (a Artist) GetUserID() string {
	return a.UserID
}

func (a Artist) GetTitle() string {
	return a.Name
}

func (a Artist) GetDescription() string {
	return a.Description
}

func (a Artist) GetReleaseDate() string {
	return a.CreationDate
}

func (a Artist) GetAdditionDate() int64 {
	return a.AdditionDate
}

func (a Artist) GetTags() []string {
	return a.Tags
}

func (a Artist) GetAdditionalMeta() map[string]interface{} {
	return a.AdditionalMeta
}

func (a Artist) GetPermissions() map[string]string {
	return a.Permissions
}

func (a Artist) IsTemporary() bool {
	return a.ID == ""
}

func (a Artist) GetLinkedItemIDs() []string {
	return a.LinkedItemIDs
}

func (a Artist) GetViewCount() int {
	return a.ListenCount
}

func (a Artist) GetMetadataSource() LinkedSource {
	return a.MetadataSource
}
