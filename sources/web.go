//go:build web_source || !(no_web_source || no_sources)

package sources

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/log"

	"github.com/libramusic/libracore/media"
	"github.com/libramusic/libracore/utils"
)

type WebSource struct {
	URL string
}

func InitWebSource(url string) (*WebSource, error) {
	log.Error("unimplemented")

	return &WebSource{
		URL: url,
	}, nil
}

func (*WebSource) Satisfies(id string) bool {
	return !strings.HasPrefix(id, "file:") && utils.IsValidSourceURL(id)
}

func (s *WebSource) SupportsMultiple() bool {
	return s.URL == ""
}

func (s *WebSource) DeriveNew(id string) (Source, error) {
	if s.SupportsMultiple() {
		return InitWebSource(id)
	}
	return nil, fmt.Errorf("source '%s' does not support multiple instances", s.GetID())
}

func (*WebSource) GetID() string {
	log.Error("unimplemented")
	return "web"
}

func (*WebSource) GetName() string {
	log.Error("unimplemented")
	return "Web"
}

func (*WebSource) GetVersion() *semver.Version {
	log.Error("unimplemented")
	return utils.LibraVersion
}

func (*WebSource) GetSourceTypes() []string {
	log.Error("unimplemented")
	return []string{"content", "metadata", "lyrics"}
}

func (*WebSource) GetMediaTypes() []string {
	log.Error("unimplemented")
	return []string{"music", "video", "playlist"}
}

func (*WebSource) Search(_ string, _, _ int, _ map[string]any) ([]media.SourcePlayable, error) {
	var results []media.SourcePlayable

	log.Error("unimplemented")

	return results, nil
}

func (s *WebSource) GetContent(playable media.SourcePlayable) ([]byte, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return nil, media.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	log.Error("unimplemented")

	return nil, nil
}

func (s *WebSource) GetLyrics(playable media.LyricsPlayable) (map[string]string, error) {
	result := map[string]string{}

	if !SupportsMediaType(s, playable.GetType()) {
		return result, media.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	log.Error("unimplemented")

	return result, nil
}

func (s *WebSource) CompleteMetadata(playable media.SourcePlayable) (media.SourcePlayable, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return playable, media.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	log.Error("unimplemented")

	return playable, nil
}

func init() {
	source, err := InitWebSource("")
	if err != nil {
		log.Warn("Source initialization failed", "source", source.GetID(), "error", err)
	} else {
		Registry[source.GetID()] = source
	}
}
