package source

import (
	"github.com/DevReaper0/libra/types"
	"github.com/DevReaper0/libra/util"
)

type SpotifySource struct {
	manager Manager
}

func InitSpotifySource(manager Manager) (SpotifySource, error) {
	return SpotifySource{
		manager: manager,
	}, nil
}

func (s SpotifySource) GetID() string {
	return "spotify"
}

func (s SpotifySource) GetName() string {
	return "Spotify"
}

func (s SpotifySource) GetVersion() string {
	return util.LibraVersion
}

func (s SpotifySource) GetSourceTypes() []string {
	return []string{"metadata", "lyrics"}
}

func (s SpotifySource) GetMediaTypes() []string {
	return []string{"music", "video", "playlist"}
}

func (s SpotifySource) Search(query string, limit int, page int, filters map[string]string) ([]types.SourcePlayable, error) {
	var results []types.SourcePlayable

	panic("unimplemented")

	return results, nil
}

func (s SpotifySource) GetContent(playable types.SourcePlayable) ([]byte, error) {
	return nil, types.UnsupportedSourceTypeError{SourceType: "content"}
}

func (s SpotifySource) GetLyrics(playable types.LyricsPlayable) (map[string]string, error) {
	result := map[string]string{}

	if !SupportsMediaType(s, playable.GetType()) {
		return result, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	panic("unimplemented")

	return result, nil
}

func (s SpotifySource) CompleteMetadata(playable types.SourcePlayable) (types.SourcePlayable, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return playable, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	panic("unimplemented")

	return playable, nil
}
