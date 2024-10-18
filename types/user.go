package types

type User struct {
	ID              string            `json:"id"`
	Username        string            `json:"username"`
	Email           string            `json:"email"`
	PasswordHash    string            `json:"password_hash"`
	DisplayName     string            `json:"display_name"`
	Description     string            `json:"description"`
	ListenedTo      map[string]int    `json:"listened_to"`
	Favorites       []string          `json:"favorites"`
	PublicViewCount int               `json:"public_view_count"`
	CreationDate    int64             `json:"creation_date"`
	Permissions     map[string]string `json:"permissions"`
	LinkedArtistID  string            `json:"linked_artist_id"`
	LinkedSources   map[string]string `json:"linked_sources"`
}

func (u User) GetType() string {
	return "user"
}

func (u User) GetID() string {
	return u.ID
}

func (u User) GetUserID() string {
	return u.ID
}

func (u User) GetTitle() string {
	return u.DisplayName
}

func (u User) GetDescription() string {
	return u.Description
}

func (u User) GetReleaseDate() string {
	return string(u.CreationDate)
}

func (u User) GetAdditionDate() int64 {
	return u.CreationDate
}

func (u User) GetTags() []string {
	// Returns an empty array because users do not have tags
	return []string{}
}

func (u User) GetAdditionalMeta() map[string]interface{} {
	return map[string]interface{}{}
}

func (u User) GetPermissions() map[string]string {
	return u.Permissions
}

func (u User) IsTemporary() bool {
	// Returns false because the only way a playable can be temporary is if it is a search result from a source, which a user cannot be
	return false
}

type AdminPermissions struct {
}
