package types

type Artist struct {
	ID             string            `json:"id"              example:"h3r3VpPvSq8"`
	UserID         string            `json:"user_id"         example:"TPkrKcIZRRq"`
	Name           string            `json:"name"            example:"John Doe"`
	AlbumIDs       []string          `json:"album_ids"       example:"BhRpYVlrMo8,poFEUbgBuwJ"`
	TrackIDs       []string          `json:"track_ids"       example:"7nTwkcl51u4,OBTwkAXODLd"`
	Description    string            `json:"description"     example:"Artist description here."`
	CreationDate   string            `json:"creation_date"   example:"2023-10-01"`
	ListenCount    int               `json:"listen_count"    example:"150"`
	FavoriteCount  int               `json:"favorite_count"  example:"5"`
	AdditionDate   int64             `json:"addition_date"   example:"1634296980"`
	Tags           []string          `json:"tags"`
	AdditionalMeta map[string]any    `json:"additional_meta"`
	Permissions    map[string]string `json:"permissions"`
	LinkedItemIDs  []string          `json:"linked_item_ids"`
	MetadataSource string            `json:"metadata_source"`
}

func (Artist) GetType() string {
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

func (a Artist) GetAdditionalMeta() map[string]any {
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

func (a Artist) GetMetadataSource() string {
	return a.MetadataSource
}
