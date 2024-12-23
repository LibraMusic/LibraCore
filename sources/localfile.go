package sources

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/log"
	"github.com/goccy/go-json"
	ffmpeg "github.com/u2takey/ffmpeg-go"

	"github.com/LibraMusic/LibraCore/types"
	"github.com/LibraMusic/LibraCore/utils"
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

func (*LocalFileSource) GetVersion() *semver.Version {
	return utils.LibraVersion
}

func (*LocalFileSource) GetSourceTypes() []string {
	return []string{"content", "metadata", "lyrics"}
}

func (*LocalFileSource) GetMediaTypes() []string {
	return []string{"music", "video"}
}

func (s *LocalFileSource) Search(_ string, _ int, _ int, filters map[string]interface{}) ([]types.SourcePlayable, error) {
	var results []types.SourcePlayable

	fileInfo, err := os.Stat(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return results, nil
		}
		return nil, err
	}

	allowVideos := false
	if filters["allow_videos"] != nil {
		if boolValue, ok := filters["allow_videos"].(bool); ok {
			allowVideos = boolValue
		}
	}

	searchedTypes := []string{"tracks"}
	if filters["types"] != nil {
		if arrayValue, ok := filters["types"].([]string); ok {
			searchedTypes = arrayValue
		}
	}

	if slices.Contains(searchedTypes, "tracks") && allowVideos && !slices.Contains(searchedTypes, "videos") {
		searchedTypes = append(searchedTypes, "videos")
	}

	if fileInfo.IsDir() {
		log.Error("unimplemented")
	} else if slices.Contains(searchedTypes, "tracks") || slices.Contains(searchedTypes, "videos") {
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

		log.Error("unimplemented")

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

	return results, nil
}

func (s *LocalFileSource) GetContent(playable types.SourcePlayable) ([]byte, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return nil, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	log.Error("unimplemented")

	return nil, nil
}

func (s *LocalFileSource) GetLyrics(playable types.LyricsPlayable) (map[string]string, error) {
	result := map[string]string{}

	if !SupportsMediaType(s, playable.GetType()) {
		return result, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	log.Error("unimplemented")

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

	log.Error("unimplemented")

	return playable, nil
}
