package storage

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/log"

	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/db"
	"github.com/libramusic/libracore/types"
)

const (
	ContentPath = "content"
	CoversPath  = "covers"
)

func getStoragePath() (string, error) {
	path := config.Conf.Storage.Location
	if !filepath.IsAbs(path) && config.DataDir != "" {
		path = filepath.Join(config.DataDir, path)
	}
	return filepath.Abs(path)
}

func CleanOverfilledStorage(ctx context.Context) {
	path, err := getStoragePath()
	if err != nil {
		log.Error("Error getting storage path", "err", err)
		return
	}
	path = filepath.Join(path, ContentPath)
	files, err := os.ReadDir(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			log.Error("Error reading storage directory", "err", err)
		}
		return
	}

	dirs := getDirectories(files)
	contentFiles := getContentFiles(path, dirs)
	playables := getPlayables(ctx, contentFiles)

	sort.Slice(playables, func(i, j int) bool {
		return playables[i].GetViewCount() >= playables[j].GetViewCount()
	})

	var sum uint64
	storagePath, err := getStoragePath()
	if err != nil {
		log.Error("Error getting storage path", "err", err)
		return
	}
	err = filepath.Walk(storagePath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error("Error walking storage path", "err", err)
			return nil
		}
		if !info.IsDir() {
			sum += uint64(info.Size())
		}
		return nil
	})
	if err != nil {
		log.Error("Error walking storage path", "err", err)
	}

	for sum > config.Conf.Storage.SizeLimit.Bytes() {
		if len(playables) == 0 {
			log.Warn(
				"Storage is overfilled, but no playables are old enough to delete. Consider increasing the storage limit or decreasing the minimum age threshold",
			)
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
				log.Error("Error getting file info", "err", err)
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
			log.Error("Error reading directory", "err", err)
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

func getPlayables(ctx context.Context, contentFiles []string) []types.SourcePlayable {
	var playables []types.SourcePlayable

	tracks, err := db.DB.GetAllTracks(ctx)
	if err != nil {
		log.Error("Error getting all tracks from database", "err", err)
	}
	for _, track := range tracks {
		if slices.Contains(contentFiles, "track_"+track.GetID()) {
			difference := time.Since(time.Unix(track.GetAdditionDate(), 0))
			if difference >= config.Conf.Storage.MinimumAgeThreshold {
				playables = append(playables, track)
			}
		}
	}

	videos, err := db.DB.GetAllVideos(ctx)
	if err != nil {
		log.Error("Error getting all videos from database", "err", err)
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
		log.Error("Error reading directory", "err", err)
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasPrefix(file.Name(), playableID+".") {
				err := os.Remove(filepath.Join(baseContentPath, file.Name()))
				if err != nil {
					log.Error("Error removing file", "err", err)
				}
				break
			}
		}
	}

	baseCoverPath := filepath.Join(storagePath, CoversPath, playableType+"s")
	files, err = os.ReadDir(baseCoverPath)
	if err != nil {
		log.Error("Error reading directory", "err", err)
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasPrefix(file.Name(), playableID+".") {
				err := os.Remove(filepath.Join(baseCoverPath, file.Name()))
				if err != nil {
					log.Error("Error removing file", "err", err)
				}
				break
			}
		}
	}
}

func IsContentStored(contentType, playableID string) bool {
	path, err := getStoragePath()
	if err != nil {
		log.Error("Error getting storage path", "err", err)
		return false
	}
	path = filepath.Join(path, ContentPath, contentType+"s")

	files, err := os.ReadDir(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			log.Error("Error reading directory", "err", err)
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

func StoreContent(contentType, playableID string, data []byte, fileExtension string) {
	path, err := getStoragePath()
	if err != nil {
		log.Error("Error getting storage path", "err", err)
		return
	}
	path = filepath.Join(path, ContentPath, contentType+"s")
	_ = os.MkdirAll(path, os.ModePerm)
	err = os.WriteFile(filepath.Join(path, playableID+fileExtension), data, 0o644)
	if err != nil {
		log.Error("Error writing file", "err", err)
	}
}

func StoreCover(contentType, playableID string, data []byte, fileExtension string) {
	path, err := getStoragePath()
	if err != nil {
		log.Error("Error getting storage path", "err", err)
		return
	}
	path = filepath.Join(path, CoversPath, contentType+"s")
	_ = os.MkdirAll(path, os.ModePerm)
	err = os.WriteFile(filepath.Join(path, playableID+fileExtension), data, 0o644)
	if err != nil {
		log.Error("Error writing file", "err", err)
	}
}
