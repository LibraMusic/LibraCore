package sources

import (
	"slices"

	"github.com/LibraMusic/LibraCore/types"
)

type Source interface {
	GetID() string
	GetName() string
	GetVersion() string
	GetSourceTypes() []string
	GetMediaTypes() []string
	Search(query string, limit int, page int, filters map[string]interface{}) ([]types.SourcePlayable, error)
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
