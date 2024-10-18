package types

type Album struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	UPC            string                 `json:"upc"`
	Title          string                 `json:"title"`
	ArtistIDs      []string               `json:"artist_ids"`
	TrackIDs       []string               `json:"track_ids"`
	Description    string                 `json:"description"`
	ReleaseDate    string                 `json:"release_date"`
	ListenCount    int                    `json:"listen_count"`
	FavoriteCount  int                    `json:"favorite_count"`
	AdditionDate   int64                  `json:"addition_date"`
	Tags           []string               `json:"tags"`
	AdditionalMeta map[string]interface{} `json:"additional_meta"`
	Permissions    map[string]string      `json:"permissions"`
	LinkedItemIDs  []string               `json:"linked_item_ids"`
	MetadataSource LinkedSource           `json:"metadata_source"`
}

func (Album) GetType() string {
	return "album"
}

func (a Album) GetID() string {
	return a.ID
}

func (a Album) GetUserID() string {
	return a.UserID
}

func (a Album) GetTitle() string {
	return a.Title
}

func (a Album) GetDescription() string {
	return a.Description
}

func (a Album) GetReleaseDate() string {
	return a.ReleaseDate
}

func (a Album) GetAdditionDate() int64 {
	return a.AdditionDate
}

func (a Album) GetTags() []string {
	return a.Tags
}

func (a Album) GetAdditionalMeta() map[string]interface{} {
	return a.AdditionalMeta
}

func (a Album) GetPermissions() map[string]string {
	return a.Permissions
}

func (a Album) IsTemporary() bool {
	return a.ID == ""
}

func (a Album) GetLinkedItemIDs() []string {
	return a.LinkedItemIDs
}

func (a Album) GetViewCount() int {
	return a.ListenCount
}

func (a Album) GetMetadataSource() LinkedSource {
	return a.MetadataSource
}
