package sources

import (
	"errors"
	"os/exec"

	"github.com/Masterminds/semver/v3"

	"github.com/libramusic/libracore/media"
)

var (
	ErrUnknownSource       = errors.New("unknown source")
	ErrUnsupportedCategory = errors.New("unsupported source category")
	ErrUnsupportedFormat   = errors.New("unsupported content format")
)

type SharedInfo struct {
	DataDir                   string
	InheritListenCounts       bool
	ArtistListenCountsByTrack bool
}

type Source interface {
	ID() string
	Name() string
	Version() *semver.Version
	Categories() []Category
	Formats() []media.Format

	Search(query string, limit, page int, filter Filter, info BindingInfo) ([]media.SourcePlayable, error)
	Media(playable media.SourcePlayable, info BindingInfo) ([]byte, error)
	Lyrics(playable media.LyricsPlayable, info BindingInfo) (map[string]string, error)
	CompleteMetadata(playable media.SourcePlayable, info BindingInfo) (media.SourcePlayable, error)
}

func execCommand(command []string) ([]byte, error) {
	if len(command) == 0 {
		return nil, errors.New("no command provided")
	} else if len(command) == 1 {
		return exec.Command(command[0]).Output()
	}
	return exec.Command(command[0], command[1:]...).Output()
}
