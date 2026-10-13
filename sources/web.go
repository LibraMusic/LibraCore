//go:build web_source || !(no_web_source || no_sources)

package sources

import (
	"errors"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/log"

	"github.com/libramusic/libracore/media"
	"github.com/libramusic/libracore/versions"
)

type WebFactory struct{}

func (WebFactory) ID() string {
	return "web"
}

func (WebFactory) Satisfies(id string) bool {
	return slices.Contains([]string{
		"web",
		"http",
		"https",
	}, strings.ToLower(id))
}

func (WebFactory) SupportsMultiple() bool {
	return true
}

func (WebFactory) New(alias string, options map[string]any, info SharedInfo) (Source, error) {
	opts, err := parseWebOptions(options)
	if err != nil {
		return nil, err
	}
	source := &Web{alias: alias, options: opts, info: info}
	return source, nil
}

type WebOptions struct {
	URL string
}

func parseWebOptions(options map[string]any) (WebOptions, error) {
	opts := WebOptions{
		URL: "",
	}

	if url, ok := options["url"].(string); ok {
		opts.URL = url
	}
	if !isValidSourceURL(opts.URL) {
		return opts, errors.New("invalid value for option 'url'")
	}

	return opts, nil
}

func isValidSourceURL(urlStr string) bool {
	if strings.TrimSpace(urlStr) == "" {
		return false
	}
	return hasSupportedScheme(urlStr)
}

func hasSupportedScheme(urlStr string) bool {
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") ||
		!strings.Contains(urlStr, "://") {
		return true
	}
	return false
}

type Web struct {
	alias   string
	options WebOptions
	info    SharedInfo
}

func (*Web) ID() string {
	return "web"
}

func (*Web) Name() string {
	log.Error("unimplemented")
	return "Web"
}

func (*Web) Version() *semver.Version {
	log.Error("unimplemented")
	return versions.Libra
}

func (*Web) Categories() []Category {
	log.Error("unimplemented")
	return []Category{CategoryMedia, CategoryMetadata, CategoryLyrics}
}

func (*Web) Formats() []media.Format {
	log.Error("unimplemented")
	return []media.Format{media.FormatMusic, media.FormatVideo, media.FormatPlaylist}
}

func (*Web) Search(_ string, _, _ int, _ Filter, _ BindingInfo) ([]media.SourcePlayable, error) {
	var results []media.SourcePlayable

	log.Error("unimplemented")

	return results, nil
}

func (*Web) Media(_ media.SourcePlayable, _ BindingInfo) ([]byte, error) {
	log.Error("unimplemented")

	return nil, nil
}

func (*Web) Lyrics(_ media.LyricsPlayable, _ BindingInfo) (map[string]string, error) {
	result := map[string]string{}

	log.Error("unimplemented")

	return result, nil
}

func (*Web) CompleteMetadata(playable media.SourcePlayable, _ BindingInfo) (media.SourcePlayable, error) {
	log.Error("unimplemented")

	return playable, nil
}

func init() {
	Register(WebFactory{})
}
