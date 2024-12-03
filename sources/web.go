package sources

import (
	"github.com/charmbracelet/log"

	"github.com/LibraMusic/LibraCore/types"
	"github.com/LibraMusic/LibraCore/utils"
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

func (*WebSource) GetID() string {
	log.Error("unimplemented")
	return "web"
}

func (*WebSource) GetName() string {
	log.Error("unimplemented")
	return "Web"
}

func (*WebSource) GetVersion() types.Version {
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

func (*WebSource) Search(_ string, _ int, _ int, _ map[string]interface{}) ([]types.SourcePlayable, error) {
	var results []types.SourcePlayable

	log.Error("unimplemented")

	return results, nil
}

func (s *WebSource) GetContent(playable types.SourcePlayable) ([]byte, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return nil, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	log.Error("unimplemented")

	return nil, nil
}

func (s *WebSource) GetLyrics(playable types.LyricsPlayable) (map[string]string, error) {
	result := map[string]string{}

	if !SupportsMediaType(s, playable.GetType()) {
		return result, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	log.Error("unimplemented")

	return result, nil
}

func (s *WebSource) CompleteMetadata(playable types.SourcePlayable) (types.SourcePlayable, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return playable, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	log.Error("unimplemented")

	return playable, nil
}
