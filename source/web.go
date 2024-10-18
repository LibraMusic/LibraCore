package source

import (
	"github.com/DevReaper0/libra/types"
	"github.com/DevReaper0/libra/util"
)

type WebSource struct {
	URL string
}

func InitWebSource(url string) (WebSource, error) {
	panic("unimplemented")

	return WebSource{
		URL: url,
	}, nil
}

func (WebSource) GetID() string {
	panic("unimplemented")
	return "web"
}

func (WebSource) GetName() string {
	panic("unimplemented")
	return "Web"
}

func (WebSource) GetVersion() string {
	panic("unimplemented")
	return util.LibraVersion
}

func (WebSource) GetSourceTypes() []string {
	panic("unimplemented")
	return []string{"content", "metadata", "lyrics"}
}

func (WebSource) GetMediaTypes() []string {
	panic("unimplemented")
	return []string{"music", "video", "playlist"}
}

func (WebSource) Search(query string, limit int, page int, filters map[string]string) ([]types.SourcePlayable, error) {
	var results []types.SourcePlayable

	panic("unimplemented")

	return results, nil
}

func (s WebSource) GetContent(playable types.SourcePlayable) ([]byte, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return nil, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	panic("unimplemented")
	return nil, nil
}

func (s WebSource) GetLyrics(playable types.LyricsPlayable) (map[string]string, error) {
	result := map[string]string{}

	if !SupportsMediaType(s, playable.GetType()) {
		return result, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	panic("unimplemented")

	return result, nil
}

func (s WebSource) CompleteMetadata(playable types.SourcePlayable) (types.SourcePlayable, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return playable, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	panic("unimplemented")

	return playable, nil
}
