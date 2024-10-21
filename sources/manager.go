package sources

import (
	"slices"
	"strings"

	"github.com/LibraMusic/LibraCore/config"
	"github.com/LibraMusic/LibraCore/logging"
	"github.com/LibraMusic/LibraCore/types"
	"github.com/LibraMusic/LibraCore/utils"
)

var SM Manager

type Manager struct {
	sources   map[string]Source
	sourceIDs []string
}

func InitManager() {
	if SM.sources != nil {
		logging.Warn().Msg("Source manager already initialized")
		return
	}
	SM = Manager{
		sources:   map[string]Source{},
		sourceIDs: []string{},
	}
}

func (*Manager) IsHigherPriority(first string, second string) bool {
	firstPriority := slices.Index(config.Conf.General.EnabledSources, first)
	secondPriority := slices.Index(config.Conf.General.EnabledSources, second)
	return firstPriority < secondPriority || (firstPriority == -1 && secondPriority != -1) || second == ""
}

func (sm *Manager) EnableSources() {
	for _, source := range config.Conf.General.EnabledSources {
		err := sm.EnableSource(source)
		if err != nil {
			logging.Warn().Err(err).Msgf("Error enabling source %s", source)
		}
	}
}

func (sm *Manager) EnableSource(sourceStr string) (err error) {
	var source Source

	switch strings.ToLower(sourceStr) {
	case "youtube", "yt":
		source, err = InitYouTubeSource()
	case "spotify", "sp":
		source, err = InitSpotifySource()
	default:
		if strings.HasPrefix(sourceStr, "file:") {
			source, err = InitLocalFileSource(strings.TrimPrefix(sourceStr, "file:"))
		} else if utils.IsValidSourceURL(sourceStr) {
			source, err = InitWebSource(sourceStr)
		} else {
			err = types.InvalidSourceError{SourceID: sourceStr}
			return
		}
	}
	if err != nil {
		if source != nil {
			return types.SourceError{SourceID: source.GetID(), Err: err}
		}
		return types.SourceError{SourceID: sourceStr, Err: err}
	}
	sm.sources[source.GetID()] = source
	sm.sourceIDs = append(sm.sourceIDs, source.GetID())
	return nil
}

// TODO: Implement Search

// TODO: Implement GetContent

// TODO: Implement GetLyrics

// TODO: Implement CompleteMetadata

/*func (sm *Manager) GetImage(searchResult types.SearchResult) ([]byte, error) {
	if _, ok := sm.sources[searchResult.ServiceID]; ok {
		return sm.sources[searchResult.ServiceID].GetImage(searchResult)
	}
	return []byte{}, nil
}

func (sm *Manager) GetContent(searchResult types.SearchResult) ([]byte, error) {
	if _, ok := sm.sources[searchResult.ServiceID]; ok {
		return sm.sources[searchResult.ServiceID].GetContent(searchResult)
	}
	return []byte{}, nil
}*/
