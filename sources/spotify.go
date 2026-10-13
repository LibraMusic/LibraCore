//go:build spotify_source || !(no_spotify_source || no_sources)

package sources

import (
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/log"

	"github.com/libramusic/libracore/media"
	"github.com/libramusic/libracore/versions"
)

type SpotifyFactory struct{}

func (SpotifyFactory) ID() string {
	return "spotify"
}

func (SpotifyFactory) Satisfies(id string) bool {
	return slices.Contains([]string{
		"spotify",
		"sp",
	}, strings.ToLower(id))
}

func (SpotifyFactory) SupportsMultiple() bool {
	return false
}

func (SpotifyFactory) New(alias string, _ map[string]any, info SharedInfo) (Source, error) {
	return &Spotify{alias: alias, info: info}, nil
}

type Spotify struct {
	alias string
	info  SharedInfo
}

func (*Spotify) ID() string {
	return "spotify"
}

func (*Spotify) Name() string {
	return "Spotify"
}

func (*Spotify) Version() *semver.Version {
	return versions.Libra
}

func (*Spotify) Categories() []Category {
	return []Category{CategoryMetadata, CategoryLyrics}
}

func (*Spotify) Formats() []media.Format {
	return []media.Format{media.FormatMusic, media.FormatVideo, media.FormatPlaylist}
}

func (*Spotify) Search(_ string, _, _ int, _ Filter, _ BindingInfo) ([]media.SourcePlayable, error) {
	var results []media.SourcePlayable

	log.Error("unimplemented")

	return results, nil
}

func (*Spotify) Media(_ media.SourcePlayable, _ BindingInfo) ([]byte, error) {
	return nil, ErrUnsupportedCategory
}

func (*Spotify) Lyrics(_ media.LyricsPlayable, _ BindingInfo) (map[string]string, error) {
	result := map[string]string{}

	log.Error("unimplemented")

	return result, nil
}

func (*Spotify) CompleteMetadata(playable media.SourcePlayable, _ BindingInfo) (media.SourcePlayable, error) {
	log.Error("unimplemented")

	return playable, nil
}

func init() {
	Register(SpotifyFactory{})
}
