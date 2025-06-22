//go:build spotify_source || !(no_spotify_source || no_sources)

package sources

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/log"

	"github.com/libramusic/libracore/types"
	"github.com/libramusic/libracore/utils"
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

func (s *SpotifySource) DeriveNew(_ string) (Source, error) {
	return nil, fmt.Errorf("source '%s' does not support multiple instances", s.GetID())
}

func (*SpotifySource) GetID() string {
	return "spotify"
}

func (*SpotifySource) GetName() string {
	return "Spotify"
}

func (*SpotifySource) GetVersion() *semver.Version {
	return utils.LibraVersion
}

func (*SpotifySource) GetSourceTypes() []string {
	return []string{"metadata", "lyrics"}
}

func (*SpotifySource) GetMediaTypes() []string {
	return []string{"music", "video", "playlist"}
}

func (*SpotifySource) Search(_ string, _, _ int, _ map[string]any) ([]types.SourcePlayable, error) {
	var results []types.SourcePlayable

	log.Error("unimplemented")

	return results, nil
}

func (*SpotifySource) GetContent(_ types.SourcePlayable) ([]byte, error) {
	return nil, types.UnsupportedSourceTypeError{SourceType: "content"}
}

func (s *SpotifySource) GetLyrics(playable types.LyricsPlayable) (map[string]string, error) {
	result := map[string]string{}

	if !SupportsMediaType(s, playable.GetType()) {
		return result, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	log.Error("unimplemented")

	return result, nil
}

func (s *SpotifySource) CompleteMetadata(playable types.SourcePlayable) (types.SourcePlayable, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return playable, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
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
