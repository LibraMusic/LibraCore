package sources

import (
	"slices"

	"github.com/Masterminds/semver/v3"

	"github.com/libramusic/libracore/media"
)

type Source interface {
	Satisfies(id string) bool
	SupportsMultiple() bool
	DeriveNew(id string) (Source, error)

	GetID() string
	GetName() string
	GetVersion() *semver.Version
	GetSourceTypes() []string
	GetMediaTypes() []string

	Search(query string, limit, page int, filters map[string]any) ([]media.SourcePlayable, error)
	GetContent(playable media.SourcePlayable) ([]byte, error)
	GetLyrics(playable media.LyricsPlayable) (map[string]string, error)
	CompleteMetadata(playable media.SourcePlayable) (media.SourcePlayable, error)
}

func SupportsMediaType(s Source, mediaType string) bool {
	switch mediaType {
	case "music", "track", "album", "artist":
		return slices.Contains(s.GetMediaTypes(), "music")
	case "video":
		return slices.Contains(s.GetMediaTypes(), "video")
	case "playlist":
		return slices.Contains(s.GetMediaTypes(), "playlist")
	}
	return false
}
