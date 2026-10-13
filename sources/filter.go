package sources

import "github.com/libramusic/libracore/media"

type Filter struct {
	Categories []Category     `json:"categories"`
	Formats    []media.Format `json:"formats"`
}
