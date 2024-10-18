package source

import (
	"encoding/json"
	"os"
	"path/filepath"

	ffmpeg "github.com/u2takey/ffmpeg-go"

	"github.com/DevReaper0/libra/types"
	"github.com/DevReaper0/libra/util"
)

type LocalFileSource struct {
	Path string
}

func InitLocalFileSource(filepath string) (*LocalFileSource, error) {
	return &LocalFileSource{
		Path: filepath,
	}, nil
}

func (s *LocalFileSource) GetID() string {
	return "file:" + s.Path
}

func (s *LocalFileSource) GetName() string {
	return "Local File (" + s.Path + ")"
}

func (s *LocalFileSource) GetVersion() string {
	return util.LibraVersion
}

func (s *LocalFileSource) GetSourceTypes() []string {
	return []string{"content", "metadata", "lyrics"}
}

func (s *LocalFileSource) GetMediaTypes() []string {
	return []string{"music", "video"}
}

func (s *LocalFileSource) Search(query string, limit int, page int, filters map[string]string) ([]types.SourcePlayable, error) {
	var results []types.SourcePlayable

	fileInfo, err := os.Stat(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return results, nil
		}
		return nil, err
	}

	if fileInfo.IsDir() {
		panic("unimplemented")
	} else {
		if filters["types"] == "tracks" {
			out, err := ffmpeg.Probe(s.Path)
			if err != nil {
				return nil, err
			}

			var output map[string]interface{}
			err = json.Unmarshal([]byte(out), &output)
			if err != nil {
				return nil, err
			}

			displayArtists := []string{
				output["format"].(map[string]interface{})["tags"].(map[string]interface{})["artist"].(string),
			}

			// TODO

			result := types.Track{
				Title:       output["format"].(map[string]interface{})["tags"].(map[string]interface{})["title"].(string),
				Duration:    int(output["format"].(map[string]interface{})["duration"].(float64)),
				ReleaseDate: "",
				AdditionalMeta: map[string]interface{}{
					"display_artists": displayArtists,
					"display_album":   output["format"].(map[string]interface{})["tags"].(map[string]interface{})["album"].(string),
				},
				MetadataSource: types.LinkedSource(s.GetID() + "::" + filepath.Base(s.Path)),
			}

			results = append(results, result)
		}
	}

	return results, nil
}

func (s *LocalFileSource) GetContent(playable types.SourcePlayable) ([]byte, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return nil, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	panic("unimplemented")

	return nil, nil
}

func (s *LocalFileSource) GetLyrics(playable types.LyricsPlayable) (map[string]string, error) {
	result := map[string]string{}

	if !SupportsMediaType(s, playable.GetType()) {
		return result, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	panic("unimplemented")

	return result, nil
}

func (s *LocalFileSource) CompleteMetadata(playable types.SourcePlayable) (types.SourcePlayable, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return playable, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	out, err := ffmpeg.Probe(s.Path)
	if err != nil {
		return nil, err
	}

	var output []map[string]interface{}
	err = json.Unmarshal([]byte(out), &output)
	if err != nil {
		return nil, err
	}

	panic("unimplemented")

	/*result := types.Track{
		Title: "Test",
	}*/

	return playable, nil
}
