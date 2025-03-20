package types

type Album struct {
	ID             string            `json:"id"              example:"BhRpYVlrMo8"`
	UserID         string            `json:"user_id"         example:"TPkrKcIZRRq"`
	UPC            string            `json:"upc"             example:"012345678905"`
	EAN            string            `json:"ean"             example:"0012345678905"`
	Title          string            `json:"title"           example:"Lorem Ipsum"`
	ArtistIDs      []string          `json:"artist_ids"      example:"h3r3VpPvSq8,R2QTLKbHamW"`
	TrackIDs       []string          `json:"track_ids"       example:"7nTwkcl51u4,OBTwkAXODLd"`
	Description    string            `json:"description"     example:"Lorem ipsum dolor sit amet."`
	ReleaseDate    string            `json:"release_date"    example:"2023-10-01"`
	ListenCount    int               `json:"listen_count"    example:"150"`
	FavoriteCount  int               `json:"favorite_count"  example:"5"`
	AdditionDate   int64             `json:"addition_date"   example:"1634296980"`
	Tags           []string          `json:"tags"`
	AdditionalMeta map[string]any    `json:"additional_meta"`
	Permissions    map[string]string `json:"permissions"`
	LinkedItemIDs  []string          `json:"linked_item_ids"`
	MetadataSource string            `json:"metadata_source"`
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

func (a Album) GetAdditionalMeta() map[string]any {
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

func (a Album) GetMetadataSource() string {
	return a.MetadataSource
}
