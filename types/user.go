package types

import "strconv"

type User struct {
	ID              string            `json:"id"                example:"TPkrKcIZRRq"`
	Username        string            `json:"username"          example:"JohnDoe"`
	Email           string            `json:"email"             example:"john.doe@example.com"`
	DisplayName     string            `json:"display_name"      example:"John Doe"`
	Description     string            `json:"description"       example:"I am a person."`
	ListenedTo      map[string]int    `json:"listened_to"`
	Favorites       []string          `json:"favorites"`
	PublicViewCount int               `json:"public_view_count" example:"519"`
	CreationDate    int64             `json:"creation_date"`
	Permissions     map[string]string `json:"permissions"`
	LinkedArtistID  string            `json:"linked_artist_id"  example:"h3r3VpPvSq8"`
	LinkedSources   map[string]string `json:"linked_sources"`
}

type DatabaseUser struct {
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

func (du DatabaseUser) ToAPIUser() User {
	return User{
		ID:              du.ID,
		Username:        du.Username,
		Email:           du.Email,
		DisplayName:     du.DisplayName,
		Description:     du.Description,
		ListenedTo:      du.ListenedTo,
		Favorites:       du.Favorites,
		PublicViewCount: du.PublicViewCount,
		CreationDate:    du.CreationDate,
		Permissions:     du.Permissions,
		LinkedArtistID:  du.LinkedArtistID,
		LinkedSources:   du.LinkedSources,
	}
}

func (u User) ToDatabaseUser(passwordHash string) DatabaseUser {
	return DatabaseUser{
		ID:              u.ID,
		Username:        u.Username,
		Email:           u.Email,
		PasswordHash:    passwordHash,
		DisplayName:     u.DisplayName,
		Description:     u.Description,
		ListenedTo:      u.ListenedTo,
		Favorites:       u.Favorites,
		PublicViewCount: u.PublicViewCount,
		CreationDate:    u.CreationDate,
		Permissions:     u.Permissions,
		LinkedArtistID:  u.LinkedArtistID,
		LinkedSources:   u.LinkedSources,
	}
}

func (User) GetType() string {
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
	return strconv.FormatInt(u.CreationDate, 10)
}

func (u User) GetAdditionDate() int64 {
	return u.CreationDate
}

func (User) GetTags() []string {
	// Returns an empty array because users do not have tags.
	return []string{}
}

func (User) GetAdditionalMeta() map[string]interface{} {
	return map[string]interface{}{}
}

func (u User) GetPermissions() map[string]string {
	return u.Permissions
}

func (User) IsTemporary() bool {
	// Returns false because the only way a playable can be temporary is if it is a search result from a source, which a user cannot be.
	return false
}

type AdminPermissions struct{}
