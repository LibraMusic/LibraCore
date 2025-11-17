//go:build spotify_source || !(no_spotify_source || no_sources)

package sources

import (
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/log"

	"github.com/libramusic/libracore"
	"github.com/libramusic/libracore/media"
)

type SpotifySource struct{}

func InitSpotifySource() (*SpotifySource, error) {
	return &SpotifySource{}, nil
}

func (*SpotifySource) Satisfies(id string) bool {
	return slices.Contains([]string{
		"spotify",
		"sp",
	}, strings.ToLower(id))
}

func (*SpotifySource) SupportsMultiple() bool {
	return false
}

func (*SpotifySource) Derive(_ string) (Source, error) {
	return nil, ErrMultipleInstancesNotSupported
}

func (*SpotifySource) GetID() string {
	return "spotify"
}

func (*SpotifySource) GetName() string {
	return "Spotify"
}

func (*SpotifySource) GetVersion() *semver.Version {
	return libracore.LibraVersion
}

func (*SpotifySource) GetSourceTypes() []string {
	return []string{"metadata", "lyrics"}
}

func (*SpotifySource) GetMediaTypes() []string {
	return []string{"music", "video", "playlist"}
}

func (*SpotifySource) Search(_ string, _, _ int, _ map[string]any) ([]media.SourcePlayable, error) {
	var results []media.SourcePlayable

	log.Error("unimplemented")

	return results, nil
}

func (*SpotifySource) GetContent(_ media.SourcePlayable) ([]byte, error) {
	return nil, ErrUnsupportedSourceType
}

func (s *SpotifySource) GetLyrics(playable media.LyricsPlayable) (map[string]string, error) {
	result := map[string]string{}

	if !SupportsMediaType(s, playable.GetType()) {
		return result, ErrUnsupportedMediaType
	}

	log.Error("unimplemented")

	return result, nil
}

func (s *SpotifySource) CompleteMetadata(playable media.SourcePlayable) (media.SourcePlayable, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return playable, ErrUnsupportedMediaType
	}

	log.Error("unimplemented")

	return playable, nil
}

func init() {
	source, err := InitSpotifySource()
	if err != nil {
		log.Warn("Source initialization failed", "source", source.GetID(), "error", err)
	} else {
		Registry[source.GetID()] = source
	}
}
