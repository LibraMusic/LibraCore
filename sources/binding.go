package sources

import (
	"slices"

	"github.com/libramusic/libracore/media"
)

type BindingInfo struct {
	Alias      string         `yaml:"alias"`
	Categories []Category     `yaml:"categories"`
	Formats    []media.Format `yaml:"formats"`
}

type Binding struct {
	Source Source
	Info   BindingInfo
}

func (b Binding) SupportsCategory(category Category) bool {
	return slices.Contains(b.Source.Categories(), category) &&
		(len(b.Info.Categories) == 0 || slices.Contains(b.Info.Categories, category))
}

func (b Binding) SupportsFormat(format media.Format) bool {
	return slices.ContainsFunc(b.Source.Formats(), format.Matches) &&
		(len(b.Info.Formats) == 0 || slices.ContainsFunc(b.Info.Formats, format.Matches))
}

func (b Binding) Search(query string, limit, page int, filter Filter) ([]media.SourcePlayable, error) {
	return b.Source.Search(query, limit, page, filter, b.Info)
}

func (b Binding) Media(playable media.SourcePlayable) ([]byte, error) {
	return b.Source.Media(playable, b.Info)
}

func (b Binding) Lyrics(playable media.LyricsPlayable) (map[string]string, error) {
	return b.Source.Lyrics(playable, b.Info)
}

func (b Binding) CompleteMetadata(playable media.SourcePlayable) (media.SourcePlayable, error) {
	return b.Source.CompleteMetadata(playable, b.Info)
}
