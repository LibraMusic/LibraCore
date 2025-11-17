package sources

import (
	"slices"

	"github.com/charmbracelet/log"

	"github.com/libramusic/libracore/config"
)

var (
	Registry       = map[string]Source{}
	enabledSources = []string{}
)

func IsHigherPriority(first, second string) bool {
	firstPriority := slices.Index(config.Conf.General.EnabledSources, first)
	secondPriority := slices.Index(config.Conf.General.EnabledSources, second)
	return firstPriority < secondPriority || (firstPriority == -1 && secondPriority != -1) || second == ""
}

func EnableAll() {
	for _, source := range config.Conf.General.EnabledSources {
		err := Enable(source)
		if err != nil {
			log.Warn("Error enabling source", "source", source, "err", err)
		}
	}
}

func Enable(sourceStr string) error {
	for _, source := range Registry {
		if source.Satisfies(sourceStr) {
			if source.SupportsMultiple() {
				newSource, err := source.Derive(sourceStr)
				if err != nil {
					return err
				}
				Registry[newSource.GetID()] = newSource
				enabledSources = append(enabledSources, newSource.GetID())
				return nil
			}
			enabledSources = append(enabledSources, source.GetID())
			return nil
		}
	}
	return ErrInvalidSource
}

// TODO: Implement Search

// TODO: Implement GetContent

// TODO: Implement GetLyrics

// TODO: Implement CompleteMetadata

/* func GetImage(searchResult media.SearchResult) ([]byte, error) {
	if _, ok := Registry[searchResult.ServiceID]; ok {
		return Registry[searchResult.ServiceID].GetImage(searchResult)
	}
	return []byte{}, nil
}

func GetContent(searchResult media.SearchResult) ([]byte, error) {
	if _, ok := Registry[searchResult.ServiceID]; ok {
		return Registry[searchResult.ServiceID].GetContent(searchResult)
	}
	return []byte{}, nil
} */
