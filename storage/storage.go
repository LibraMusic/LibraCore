package storage

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/DevReaper0/libra/config"
	"github.com/DevReaper0/libra/db"
	"github.com/DevReaper0/libra/logging"
	"github.com/DevReaper0/libra/types"
)

const ContentPath = "content"
const CoversPath = "covers"

func CleanOverfilledStorage() {
	path, err := filepath.Abs(config.Conf.Storage.Location)
	if err != nil {
		logging.Error().Err(err).Msg("")
		return
	}
	path = filepath.Join(path, ContentPath)
	files, err := os.ReadDir(path)
	if err != nil {
		if !os.IsNotExist(err) {
			logging.Error().Err(err).Msg("")
		}
		return
	}

	dirs := getDirectories(files)
	contentFiles := getContentFiles(path, dirs)
	playables := getPlayables(contentFiles)

	sort.Slice(playables, func(i, j int) bool {
		return playables[i].GetViewCount() >= playables[j].GetViewCount()
	})

	var sum uint64
	storagePath, err := filepath.Abs(config.Conf.Storage.Location)
	if err != nil {
		logging.Error().Err(err).Msg("")
		return
	}
	err = filepath.Walk(storagePath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			logging.Error().Err(err).Msg("")
			return nil
		}
		if !info.IsDir() {
			sum += uint64(info.Size())
		}
		return nil
	})
	if err != nil {
		logging.Error().Err(err).Msg("")
	}

	for sum > config.Conf.Storage.SizeLimit.Bytes() {
		if len(playables) == 0 {
			logging.Warn().Msg("Storage is overfilled, but no playables are old enough to delete. Consider increasing the storage limit or decreasing the minimum age threshold.")
			break
		}

		playable := playables[0]
		playables = playables[1:]
		removePlayableFiles(storagePath, playable)
	}
}

func getDirectories(files []os.DirEntry) []os.FileInfo {
	var dirs []os.FileInfo
	for _, file := range files {
		if file.IsDir() {
			name := file.Name()
			if name != "tracks" && name != "videos" {
				continue
			}

			info, err := file.Info()
			if err != nil {
				logging.Error().Err(err).Msg("")
				continue
			}
			dirs = append(dirs, info)
		}
	}
	return dirs
}

func getContentFiles(path string, dirs []os.FileInfo) []string {
	var contentFiles []string
	for _, dir := range dirs {
		dirPath := filepath.Join(path, dir.Name())
		files, err := os.ReadDir(dirPath)
		if err != nil {
			logging.Error().Err(err).Msg("")
			continue
		}

		prefix := dir.Name()[:len(dir.Name())-1]

		for _, file := range files {
			if !file.IsDir() {
				contentFiles = append(contentFiles, prefix+"_"+file.Name())
			}
		}
	}
	return contentFiles
}

func getPlayables(contentFiles []string) []types.SourcePlayable {
	var playables []types.SourcePlayable

	tracks, err := db.DB.GetAllTracks()
	if err != nil {
		logging.Error().Err(err).Msg("Error getting all tracks from database")
	}
	for _, track := range tracks {
		if slices.Contains(contentFiles, "track_"+track.GetID()) {
			difference := time.Since(time.Unix(track.GetAdditionDate(), 0))
			if difference >= config.Conf.Storage.MinimumAgeThreshold {
				playables = append(playables, track)
			}
		}
	}

	videos, err := db.DB.GetAllVideos()
	if err != nil {
		logging.Error().Err(err).Msg("Error getting all videos from database")
	}
	for _, video := range videos {
		if slices.Contains(contentFiles, "video_"+video.GetID()) {
			difference := time.Since(time.Unix(video.GetAdditionDate(), 0))
			if difference >= config.Conf.Storage.MinimumAgeThreshold {
				playables = append(playables, video)
			}
		}
	}

	return playables
}

func removePlayableFiles(storagePath string, playable types.SourcePlayable) {
	playableType := playable.GetType()
	playableID := playable.GetID()

	baseContentPath := filepath.Join(storagePath, ContentPath, playableType+"s")
	files, err := os.ReadDir(baseContentPath)
	if err != nil {
		logging.Error().Err(err).Msg("")
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasPrefix(file.Name(), playableID+".") {
				err := os.Remove(filepath.Join(baseContentPath, file.Name()))
				if err != nil {
					logging.Error().Err(err).Msg("")
				}
				break
			}
		}
	}

	baseCoverPath := filepath.Join(storagePath, CoversPath, playableType+"s")
	files, err = os.ReadDir(baseCoverPath)
	if err != nil {
		logging.Error().Err(err).Msg("")
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasPrefix(file.Name(), playableID+".") {
				err := os.Remove(filepath.Join(baseCoverPath, file.Name()))
				if err != nil {
					logging.Error().Err(err).Msg("")
				}
				break
			}
		}
	}
}

func IsContentStored(contentType string, playableID string) bool {
	path, err := filepath.Abs(config.Conf.Storage.Location)
	if err != nil {
		logging.Error().Err(err).Msg("")
		return false
	}
	path = filepath.Join(path, ContentPath, contentType+"s")

	files, err := os.ReadDir(path)
	if err != nil {
		if !os.IsNotExist(err) {
			logging.Error().Err(err).Msg("")
		}
		return false
	}

	for _, file := range files {
		if !file.IsDir() {
			if strings.HasPrefix(file.Name(), playableID+".") {
				return true
			}
		}
	}

	return false
}

func StoreContent(contentType string, playableID string, data []byte, fileExtension string) {
	path, err := filepath.Abs(config.Conf.Storage.Location)
	if err != nil {
		logging.Error().Err(err).Msg("")
		return
	}
	path = filepath.Join(path, ContentPath, contentType+"s")
	os.MkdirAll(path, os.ModePerm)
	err = os.WriteFile(filepath.Join(path, playableID+fileExtension), data, 0644)
	if err != nil {
		logging.Error().Err(err).Msg("")
	}
}

func StoreCover(contentType string, playableID string, data []byte, fileExtension string) {
	path, err := filepath.Abs(config.Conf.Storage.Location)
	if err != nil {
		logging.Error().Err(err).Msg("")
		return
	}
	path = filepath.Join(path, CoversPath, contentType+"s")
	os.MkdirAll(path, os.ModePerm)
	err = os.WriteFile(filepath.Join(path, playableID+fileExtension), data, 0644)
	if err != nil {
		logging.Error().Err(err).Msg("")
	}
}
