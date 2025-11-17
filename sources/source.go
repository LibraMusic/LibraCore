package sources

import (
	"errors"
	"os/exec"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/libramusic/libracore/media"
)

var (
	ErrInvalidSource                 = errors.New("invalid source")
	ErrUnsupportedSourceType         = errors.New("unsupported source type")
	ErrUnsupportedMediaType          = errors.New("unsupported media type")
	ErrMultipleInstancesNotSupported = errors.New("source does not support multiple instances")
)

type Source interface {
	Satisfies(id string) bool
	SupportsMultiple() bool
	Derive(id string) (Source, error)

	ID() string
	Name() string
	Version() *semver.Version
	SourceTypes() []string
	MediaTypes() []string

	Search(query string, limit, page int, filters map[string]any) ([]media.SourcePlayable, error)
	Content(playable media.SourcePlayable) ([]byte, error)
	Lyrics(playable media.LyricsPlayable) (map[string]string, error)
	CompleteMetadata(playable media.SourcePlayable) (media.SourcePlayable, error)
}

func SupportsMediaType(s Source, mediaType string) bool {
	switch mediaType {
	case "music", "track", "album", "artist":
		return slices.Contains(s.MediaTypes(), "music")
	case "video":
		return slices.Contains(s.MediaTypes(), "video")
	case "playlist":
		return slices.Contains(s.MediaTypes(), "playlist")
	}
	return false
}

func IsValidSourceURL(urlStr string) bool {
	return HasSupportedScheme(urlStr)
}

func HasSupportedScheme(urlStr string) bool {
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") ||
		!strings.Contains(urlStr, "://") {
		return true
	}
	return false
}

func execCommand(command []string) ([]byte, error) {
	if len(command) == 0 {
		return nil, errors.New("no command provided")
	} else if len(command) == 1 {
		return exec.Command(command[0]).Output()
	}
	return exec.Command(command[0], command[1:]...).Output()
}
