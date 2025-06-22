//go:build localfile_source || !(no_localfile_source || no_sources)

package sources

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/log"
	"github.com/goccy/go-json"
	ffmpeg "github.com/u2takey/ffmpeg-go"

	"github.com/libramusic/libracore/types"
	"github.com/libramusic/libracore/utils"
)

type LocalFileSource struct {
	Path string
}

func InitLocalFileSource(path string) (*LocalFileSource, error) {
	return &LocalFileSource{
		Path: strings.TrimPrefix(path, "file:"),
	}, nil
}

func (*LocalFileSource) Satisfies(id string) bool {
	return strings.HasPrefix(id, "file:")
}

func (s *LocalFileSource) SupportsMultiple() bool {
	return s.Path == ""
}

func (s *LocalFileSource) DeriveNew(id string) (Source, error) {
	if s.SupportsMultiple() {
		return InitLocalFileSource(id)
	}
	return nil, fmt.Errorf("source '%s' does not support multiple instances", s.GetID())
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

func (s *LocalFileSource) Search(_ string, _, _ int, filters map[string]any) ([]types.SourcePlayable, error) {
	var results []types.SourcePlayable

	fileInfo, err := os.Stat(s.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
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

		var output map[string]any
		err = json.Unmarshal([]byte(out), &output)
		if err != nil {
			return nil, err
		}

		displayArtists := []string{
			output["format"].(map[string]any)["tags"].(map[string]any)["artist"].(string),
		}

		log.Error("unimplemented")

		result := types.Track{
			Title:       output["format"].(map[string]any)["tags"].(map[string]any)["title"].(string),
			Duration:    int(output["format"].(map[string]any)["duration"].(float64)),
			ReleaseDate: "",
			AdditionalMeta: map[string]any{
				"display_artists": displayArtists,
				"display_album":   output["format"].(map[string]any)["tags"].(map[string]any)["album"].(string),
			},
			MetadataSource: s.GetID() + "::" + filepath.Base(s.Path),
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

	var output []map[string]any
	err = json.Unmarshal([]byte(out), &output)
	if err != nil {
		return nil, err
	}

	log.Error("unimplemented")

	return playable, nil
}

func init() {
	source, err := InitLocalFileSource("")
	if err != nil {
		log.Warn("Source initialization failed", "source", source.GetID(), "error", err)
	} else {
		Registry[source.GetID()] = source
	}
}
