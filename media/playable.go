package media

import "math/rand/v2"

type Playable interface {
	GetType() string
	GetID() string
	GetUserID() string
	GetTitle() string
	GetDescription() string
	GetReleaseDate() string
	GetAdditionDate() int64
	GetTags() []string
	GetAdditionalMeta() map[string]any
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
	GetMetadataSource() string
}

type LyricsPlayable interface {
	SourcePlayable

	GetLyrics() map[string]string
	GetLyricSources() map[string]string
}

func GenerateID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLength := len(charset)

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.IntN(charsetLength)]
	}
	return string(b)
}
