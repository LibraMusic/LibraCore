//go:build filesystem_source || !(no_filesystem_source || no_sources)

package sources

import (
	"errors"
	"io/fs"
	"os"
	"slices"
	"strings"

	"charm.land/log/v2"
	"github.com/Masterminds/semver/v3"
	"github.com/goccy/go-json"
	ffmpeg "github.com/u2takey/ffmpeg-go"

	"github.com/libramusic/libracore/media"
	"github.com/libramusic/libracore/versions"
)

type FileSystemFactory struct{}

func (FileSystemFactory) ID() string {
	return "filesystem"
}

func (FileSystemFactory) Satisfies(id string) bool {
	return slices.Contains([]string{
		"filesystem",
		"fs",
		"files",
		"file",
		"local",
		"localfile",
		"storage",
	}, strings.ToLower(id))
}

func (FileSystemFactory) SupportsMultiple() bool {
	return true
}

func (FileSystemFactory) New(alias string, options map[string]any, info SharedInfo) (Source, error) {
	opts, err := parseFileSystemOptions(options)
	if err != nil {
		return nil, err
	}
	source := &FileSystem{alias: alias, options: opts, info: info}
	return source, nil
}

type FileSystemOptions struct {
	Path string
}

func parseFileSystemOptions(options map[string]any) (FileSystemOptions, error) {
	opts := FileSystemOptions{
		Path: "",
	}

	if path, ok := options["path"].(string); ok {
		opts.Path = path
	}
	if strings.TrimSpace(opts.Path) == "" {
		return opts, errors.New("missing required option 'path'")
	}

	return opts, nil
}

type FileSystem struct {
	alias   string
	options FileSystemOptions
	info    SharedInfo
}

func (*FileSystem) ID() string {
	return "filesystem"
}

func (s *FileSystem) Name() string {
	return "File System (" + s.options.Path + ")"
}

func (*FileSystem) Version() *semver.Version {
	return versions.Libra
}

func (*FileSystem) Categories() []Category {
	return []Category{CategoryMedia, CategoryMetadata, CategoryLyrics}
}

func (*FileSystem) Formats() []media.Format {
	return []media.Format{media.FormatMusic, media.FormatVideo}
}

func (s *FileSystem) Search(_ string, _, _ int, _ Filter, _ BindingInfo) ([]media.SourcePlayable, error) {
	var results []media.SourcePlayable

	fileInfo, err := os.Stat(s.options.Path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return results, nil
		}
		return nil, err
	}

	if fileInfo.IsDir() {
		log.Error("unimplemented")
	} else {
		out, err := ffmpeg.Probe(s.options.Path)
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

		result := media.Track{
			Title:       output["format"].(map[string]any)["tags"].(map[string]any)["title"].(string),
			Duration:    int(output["format"].(map[string]any)["duration"].(float64)),
			ReleaseDate: "",
			AdditionalMeta: map[string]any{
				"display_artists": displayArtists,
				"display_album":   output["format"].(map[string]any)["tags"].(map[string]any)["album"].(string),
				".path":           s.options.Path,
			},
			MetadataSource: s.alias,
		}

		results = append(results, result)
	}

	return results, nil
}

func (*FileSystem) Media(_ media.SourcePlayable, _ BindingInfo) ([]byte, error) {
	log.Error("unimplemented")

	return nil, nil
}

func (*FileSystem) Lyrics(_ media.LyricsPlayable, _ BindingInfo) (map[string]string, error) {
	result := map[string]string{}

	log.Error("unimplemented")

	return result, nil
}

func (s *FileSystem) CompleteMetadata(playable media.SourcePlayable, _ BindingInfo) (media.SourcePlayable, error) {
	out, err := ffmpeg.Probe(s.options.Path)
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
	Register(FileSystemFactory{})
}
