package sources

import (
	"slices"

	"github.com/Masterminds/semver/v3"

	"github.com/libramusic/libracore/types"
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

	Search(query string, limit, page int, filters map[string]any) ([]types.SourcePlayable, error)
	GetContent(playable types.SourcePlayable) ([]byte, error)
	GetLyrics(playable types.LyricsPlayable) (map[string]string, error)
	CompleteMetadata(playable types.SourcePlayable) (types.SourcePlayable, error)
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
