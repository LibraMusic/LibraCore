package sources

import (
	"github.com/DevReaper0/libra/logging"
	"github.com/DevReaper0/libra/types"
	"github.com/DevReaper0/libra/utils"
)

type SpotifySource struct {
}

func InitSpotifySource() (*SpotifySource, error) {
	return &SpotifySource{}, nil
}

func (*SpotifySource) GetID() string {
	return "spotify"
}

func (*SpotifySource) GetName() string {
	return "Spotify"
}

func (*SpotifySource) GetVersion() string {
	return utils.LibraVersion
}

func (*SpotifySource) GetSourceTypes() []string {
	return []string{"metadata", "lyrics"}
}

func (*SpotifySource) GetMediaTypes() []string {
	return []string{"music", "video", "playlist"}
}

func (*SpotifySource) Search(_ string, _ int, _ int, _ map[string]interface{}) ([]types.SourcePlayable, error) {
	var results []types.SourcePlayable

	logging.Error().Msg("unimplemented")

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

	logging.Error().Msg("unimplemented")

	return result, nil
}

func (s *SpotifySource) CompleteMetadata(playable types.SourcePlayable) (types.SourcePlayable, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return playable, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	logging.Error().Msg("unimplemented")

	return playable, nil
}