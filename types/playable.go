package types

type Playable interface {
	GetType() string
	GetID() string
	GetUserID() string
	GetTitle() string
	GetDescription() string
	GetReleaseDate() string
	GetAdditionDate() int64
	GetTags() []string
	GetAdditionalMeta() map[string]interface{}
	GetPermissions() map[string]string
	IsTemporary() bool
}

type LinkablePlayable interface {
	Playable

	GetLinkedItemIDs() []string
}

type SourcePlayable interface {
	Playable

	GetViewCount() int
	GetMetadataSource() LinkedSource
}

type LyricsPlayable interface {
	SourcePlayable

	GetLyrics() map[string]string
	GetLyricSources() map[string]LinkedSource
}
