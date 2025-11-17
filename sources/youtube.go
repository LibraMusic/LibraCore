//go:build youtube_source || !(no_youtube_source || no_sources)

package sources

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/log"
	"github.com/goccy/go-json"

	"github.com/libramusic/libracore"
	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/media"
	"github.com/libramusic/libracore/storage"
)

//go:embed scripts/youtube.py
var youtubeScript string

type YouTubeSource struct{}

func InitYouTubeSource() (*YouTubeSource, error) {
	youtubeLocation := getYouTubeScriptPath()

	if _, err := os.Stat(youtubeLocation); errors.Is(err, fs.ErrNotExist) {
		err = os.MkdirAll(filepath.Dir(youtubeLocation), os.ModePerm)
		if err != nil {
			return &YouTubeSource{}, err
		}

		err = os.WriteFile(youtubeLocation, []byte(youtubeScript), 0o644)
		if err != nil {
			return &YouTubeSource{}, err
		}
	}

	return &YouTubeSource{}, nil
}

func getYouTubeScriptPath() string {
	path := config.Conf.SourceScripts.YouTubeLocation
	if !filepath.IsAbs(path) && config.DataDir != "" {
		absDataDir, err := filepath.Abs(config.DataDir)
		if err != nil {
			return filepath.Join(config.DataDir, path)
		}
		return filepath.Join(absDataDir, path)
	}
	return path
}

func (*YouTubeSource) Satisfies(id string) bool {
	return slices.Contains([]string{
		"youtube",
		"yt",
	}, strings.ToLower(id))
}

func (*YouTubeSource) SupportsMultiple() bool {
	return false
}

func (s *YouTubeSource) DeriveNew(_ string) (Source, error) {
	return nil, fmt.Errorf("source %q does not support multiple instances", s.GetID())
}

func (*YouTubeSource) GetID() string {
	return "youtube"
}

func (*YouTubeSource) GetName() string {
	return "YouTube"
}

func (*YouTubeSource) GetVersion() *semver.Version {
	return libracore.LibraVersion
}

func (*YouTubeSource) GetSourceTypes() []string {
	return []string{"content", "metadata", "lyrics"}
}

func (*YouTubeSource) GetMediaTypes() []string {
	return []string{"music", "video", "playlist"}
}

func (s *YouTubeSource) Search(query string, limit, _ int, filters map[string]any) ([]media.SourcePlayable, error) {
	var results []media.SourcePlayable

	// TODO: Implement pagination if possible.

	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return results, fmt.Errorf("error encoding filters: %w", err)
	}

	youtubeLocation := getYouTubeScriptPath()

	command := append(strings.Split(config.Conf.SourceScripts.PythonCommand, " "), []string{
		youtubeLocation,
		`search`,
		fmt.Sprintf(`query="%s"`, strings.ReplaceAll(query, `"`, `\"`)),
		fmt.Sprintf(`limit=%d`, limit),
		fmt.Sprintf(`filters="%s"`, strings.ReplaceAll(string(filtersJSON), `"`, `\"`)),
	}...)
	out, err := execCommand(command)
	if err != nil {
		return results, fmt.Errorf("error executing command: %w", err)
	}

	var output []map[string]any
	err = json.Unmarshal(out, &output)
	if err != nil {
		return results, fmt.Errorf("error parsing command output: %w", err)
	}

	for _, v := range output {
		result, err := s.parseSearchResult(v)
		if err != nil {
			return results, fmt.Errorf("error parsing search result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

func (s *YouTubeSource) parseSearchResult(v map[string]any) (media.SourcePlayable, error) {
	var result media.SourcePlayable
	var err error

	switch v["resultType"] {
	case "song":
		result, err = s.parseSongResult(v)
	case "album":
		result, err = s.parseAlbumResult(v)
	case "video":
		result, err = s.parseVideoResult(v)
	case "artist":
		result, err = s.parseArtistResult(v)
	case "playlist":
		result, err = s.parsePlaylistResult(v)
	default:
		err = media.UnsupportedMediaTypeError{MediaType: v["resultType"].(string)}
	}

	return result, err
}

func (s *YouTubeSource) parseSongResult(v map[string]any) (media.SourcePlayable, error) {
	var displayArtists []string
	for _, artist := range v["artists"].([]map[string]string) {
		displayArtists = append(displayArtists, artist["name"])
	}

	var year string
	if v["year"] == nil {
		year = ""
	} else {
		year = strconv.Itoa(v["year"].(int))
	}

	thumbnails := v["thumbnails"].([]map[string]any)
	thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
	thumbnail, err := storage.DownloadFile(thumbnailURL)
	if err != nil {
		return nil, err
	}

	return media.Track{
		Title:       v["title"].(string),
		Duration:    v["duration_seconds"].(int),
		ReleaseDate: year,
		AdditionalMeta: map[string]any{
			"display_artists":   displayArtists,
			"display_album":     v["album"].(map[string]string)["name"],
			"display_cover_art": thumbnail,
			"yt_id":             v["videoId"].(string),
			"yt_artists":        v["artists"].([]map[string]string),
			"yt_album":          v["album"].(map[string]string),
		},
		MetadataSource: s.GetID() + "::" + "https://music.youtube.com/watch?v=" + v["videoId"].(string),
	}, nil
}

func (s *YouTubeSource) parseAlbumResult(v map[string]any) (media.SourcePlayable, error) {
	var displayArtists []string
	for _, artist := range v["artists"].([]map[string]string) {
		displayArtists = append(displayArtists, artist["name"])
	}

	var year string
	if v["year"] == nil {
		year = ""
	} else {
		year = strconv.Itoa(v["year"].(int))
	}

	thumbnails := v["thumbnails"].([]map[string]any)
	thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
	thumbnail, err := storage.DownloadFile(thumbnailURL)
	if err != nil {
		return nil, err
	}

	return media.Album{
		Title:       v["title"].(string),
		ReleaseDate: year,
		AdditionalMeta: map[string]any{
			"display_artists":   displayArtists,
			"display_cover_art": thumbnail,
			"yt_id":             v["browseId"].(string),
			"yt_artists":        v["artists"].([]map[string]string),
		},
		MetadataSource: s.GetID() + "::" + "https://music.youtube.com/browse/" + v["browseId"].(string),
	}, nil
}

func (s *YouTubeSource) parseVideoResult(v map[string]any) (media.SourcePlayable, error) {
	if !config.Conf.General.IncludeVideoResults {
		return nil, media.UnsupportedMediaTypeError{MediaType: v["resultType"].(string)}
	}

	var displayArtists []string
	for _, artist := range v["artists"].([]map[string]string) {
		displayArtists = append(displayArtists, artist["name"])
	}

	var year string
	if v["year"] == nil {
		year = ""
	} else {
		year = strconv.Itoa(v["year"].(int))
	}

	if config.Conf.General.VideoAudioOnly {
		var album string
		if v["album"] == nil {
			album = ""
		} else {
			album = v["album"].(map[string]string)["name"]
		}

		thumbnails := v["thumbnails"].([]map[string]any)
		thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
		thumbnail, err := storage.DownloadFile(thumbnailURL)
		if err != nil {
			return nil, err
		}

		return media.Track{
			Title:       v["title"].(string),
			Duration:    v["duration_seconds"].(int),
			ReleaseDate: year,
			AdditionalMeta: map[string]any{
				"display_artists":   displayArtists,
				"display_album":     album,
				"display_cover_art": thumbnail,
				"is_video":          true,
				"yt_id":             v["videoId"].(string),
				"yt_artists":        v["artists"].([]map[string]string),
				"yt_album":          album,
			},
			MetadataSource: s.GetID() + "::" + "https://music.youtube.com/watch?v=" + v["videoId"].(string),
		}, nil
	}

	thumbnails := v["thumbnails"].([]map[string]any)
	thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
	thumbnail, err := storage.DownloadFile(thumbnailURL)
	if err != nil {
		return nil, err
	}

	return media.Video{
		Title:       v["title"].(string),
		Duration:    v["duration_seconds"].(int),
		ReleaseDate: year,
		AdditionalMeta: map[string]any{
			"display_artists":   displayArtists,
			"display_thumbnail": thumbnail,
			"yt_id":             v["videoId"].(string),
			"yt_artists":        v["artists"].([]map[string]string),
		},
		MetadataSource: s.GetID() + "::" + "https://www.youtube.com/watch?v=" + v["videoId"].(string),
	}, nil
}

func (s *YouTubeSource) parseArtistResult(v map[string]any) (media.SourcePlayable, error) {
	thumbnails := v["thumbnails"].([]map[string]any)
	thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
	thumbnail, err := storage.DownloadFile(thumbnailURL)
	if err != nil {
		return nil, err
	}

	return media.Artist{
		Name: v["artist"].(string),
		AdditionalMeta: map[string]any{
			"display_cover_art": thumbnail,
			"yt_id":             v["browseId"].(string),
		},
		MetadataSource: s.GetID() + "::" + "https://music.youtube.com/channel/" + v["browseId"].(string),
	}, nil
}

func (s *YouTubeSource) parsePlaylistResult(v map[string]any) (media.SourcePlayable, error) {
	var displayArtists []string
	for _, artist := range v["artists"].([]map[string]string) {
		displayArtists = append(displayArtists, artist["name"])
	}

	thumbnails := v["thumbnails"].([]map[string]any)
	thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
	thumbnail, err := storage.DownloadFile(thumbnailURL)
	if err != nil {
		return nil, err
	}

	return media.Playlist{
		Title: v["title"].(string),
		AdditionalMeta: map[string]any{
			"display_artists":   displayArtists,
			"display_cover_art": thumbnail,
			"yt_id":             v["browseId"].(string),
			"yt_artists":        v["artists"].([]map[string]string),
		},
		MetadataSource: s.GetID() + "::" + "https://music.youtube.com/playlist?list=" + v["browseId"].(string),
	}, nil
}

func (s *YouTubeSource) GetContent(playable media.SourcePlayable) ([]byte, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return nil, media.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	var command []string
	youtubeLocation := getYouTubeScriptPath()

	switch playable.GetType() {
	case "track":
		command = append(strings.Split(config.Conf.SourceScripts.PythonCommand, " "), []string{
			youtubeLocation,
			`content`,
			`type=audio`,
			`id=` + playable.GetAdditionalMeta()["yt_id"].(string),
		}...)
	case "video":
		command = append(strings.Split(config.Conf.SourceScripts.PythonCommand, " "), []string{
			youtubeLocation,
			`content`,
			`type=video`,
			`id=` + playable.GetAdditionalMeta()["yt_id"].(string),
		}...)
	default:
		return nil, media.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	out, err := execCommand(command)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *YouTubeSource) GetLyrics(playable media.LyricsPlayable) (map[string]string, error) {
	result := map[string]string{}

	if !SupportsMediaType(s, playable.GetType()) {
		return result, media.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	var command []string
	youtubeLocation := getYouTubeScriptPath()

	if playable.GetType() == "video" ||
		playable.GetAdditionalMeta()["is_video"] == true { //revive:disable-line:bool-literal-in-expr Value cannot be used as a boolean
		command = append(strings.Split(config.Conf.SourceScripts.PythonCommand, " "), []string{
			youtubeLocation,
			`subtitles`,
			`id=` + playable.GetAdditionalMeta()["yt_id"].(string),
		}...)
	} else {
		command = append(strings.Split(config.Conf.SourceScripts.PythonCommand, " "), []string{
			youtubeLocation,
			`lyrics`,
			`id=` + playable.GetAdditionalMeta()["yt_id"].(string),
		}...)
	}

	out, err := execCommand(command)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(out, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (s *YouTubeSource) CompleteMetadata(playable media.SourcePlayable) (media.SourcePlayable, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return playable, media.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	youtubeLocation := getYouTubeScriptPath()

	command := append(strings.Split(config.Conf.SourceScripts.PythonCommand, " "), []string{
		youtubeLocation,
		playable.GetType(),
		`id=` + playable.GetAdditionalMeta()["yt_id"].(string),
	}...)

	out, err := execCommand(command)
	if err != nil {
		return playable, err
	}

	var output map[string]any
	err = json.Unmarshal(out, &output)
	if err != nil {
		return playable, err
	}

	switch playable.GetType() {
	case "track":
		return s.completeTrackMetadata(playable, output)
	case "album":
		return s.completeAlbumMetadata(playable, output)
	case "video":
		return s.completeVideoMetadata(playable, output)
	case "artist":
		return s.completeArtistMetadata(playable, output)
	case "playlist":
		return s.completePlaylistMetadata(playable, output)
	}

	return playable, nil
}

func (*YouTubeSource) completeTrackMetadata(
	playable media.SourcePlayable,
	output map[string]any,
) (media.SourcePlayable, error) {
	result := playable.(media.Track)

	lyricsID := output["track"].(map[string]any)["lyricsId"]
	if lyricsID != nil {
		splitLyricsID := strings.Split(lyricsID.(string), "-")
		trackNumber, err := strconv.Atoi(splitLyricsID[len(splitLyricsID)-1])
		if err == nil {
			result.TrackNumber = trackNumber
		}
	}

	result.Description = output["video"].(map[string]any)["microformat"].(map[string]any)["microformatDataRenderer"].(map[string]any)["description"].(string)

	publishDateObj := output["video"].(map[string]any)["microformat"].(map[string]any)["microformatDataRenderer"].(map[string]any)["publishDate"]
	if publishDateObj != nil {
		publishDateStr := publishDateObj.(string)
		t, err := time.Parse(time.RFC3339, publishDateStr)
		if err != nil {
			return result, err
		}
		result.ReleaseDate = t.Format(time.DateTime)
	}

	if config.Conf.General.InheritListenCounts {
		viewCountObj := output["video"].(map[string]any)["microformat"].(map[string]any)["microformatDataRenderer"].(map[string]any)["viewCount"]
		if viewCountObj != nil {
			viewCountStr := viewCountObj.(string)
			viewCount, err := strconv.Atoi(viewCountStr)
			if err != nil {
				return result, err
			}
			result.ListenCount = viewCount
		}
	}

	thumbnails := output["track"].(map[string]any)["thumbnail"].([]map[string]any)
	thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
	thumbnail, err := storage.DownloadFile(thumbnailURL)
	if err != nil {
		return result, err
	}
	result.AdditionalMeta["display_cover_art"] = thumbnail

	return result, nil
}

func (*YouTubeSource) completeAlbumMetadata(
	playable media.SourcePlayable,
	output map[string]any,
) (media.SourcePlayable, error) {
	result := playable.(media.Album)

	result.Description = output["description"].(string)

	thumbnails := output["thumbnails"].([]map[string]any)
	thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
	thumbnail, err := storage.DownloadFile(thumbnailURL)
	if err != nil {
		return result, err
	}
	result.AdditionalMeta["display_cover_art"] = thumbnail

	result.AdditionalMeta["yt_tracks"] = output["tracks"].([]map[string]any)

	result.AdditionalMeta["display_track_count"] = output["trackCount"].(int)

	return result, nil
}

func (*YouTubeSource) completeVideoMetadata(
	playable media.SourcePlayable,
	output map[string]any,
) (media.SourcePlayable, error) {
	result := playable.(media.Video)

	result.Description = output["video"].(map[string]any)["microformat"].(map[string]any)["microformatDataRenderer"].(map[string]any)["description"].(string)

	publishDateObj := output["video"].(map[string]any)["microformat"].(map[string]any)["microformatDataRenderer"].(map[string]any)["publishDate"]
	if publishDateObj != nil {
		publishDateStr := publishDateObj.(string)
		t, err := time.Parse(time.RFC3339, publishDateStr)
		if err != nil {
			return result, err
		}
		result.ReleaseDate = t.Format(time.DateTime)
	}

	if config.Conf.General.InheritListenCounts {
		viewCountObj := output["video"].(map[string]any)["microformat"].(map[string]any)["microformatDataRenderer"].(map[string]any)["viewCount"]
		if viewCountObj != nil {
			viewCountStr := viewCountObj.(string)
			viewCount, err := strconv.Atoi(viewCountStr)
			if err != nil {
				return result, err
			}
			result.WatchCount = viewCount
		}
	}

	thumbnails := output["track"].(map[string]any)["thumbnail"].([]map[string]any)
	thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
	thumbnail, err := storage.DownloadFile(thumbnailURL)
	if err != nil {
		return result, err
	}
	result.AdditionalMeta["display_thumbnail"] = thumbnail

	return result, nil
}

func (*YouTubeSource) completeArtistMetadata(
	playable media.SourcePlayable,
	output map[string]any,
) (media.SourcePlayable, error) {
	result := playable.(media.Artist)

	result.Description = output["description"].(string)

	if config.Conf.General.InheritListenCounts && !config.Conf.General.ArtistListenCountsByTrack {
		viewCountObj := output["views"]
		if viewCountObj != nil {
			viewCountStr := viewCountObj.(string)
			viewCountStr = strings.ReplaceAll(viewCountStr, ",", "")
			viewCountStr = strings.ReplaceAll(viewCountStr, " views", "")
			viewCountStr = strings.ReplaceAll(viewCountStr, " view", "")
			viewCount, err := strconv.Atoi(viewCountStr)
			if err != nil {
				return result, err
			}
			result.ListenCount = viewCount
		}
	}

	thumbnails := output["thumbnails"].([]map[string]any)
	thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
	thumbnail, err := storage.DownloadFile(thumbnailURL)
	if err != nil {
		return result, err
	}
	result.AdditionalMeta["display_cover_art"] = thumbnail

	tracks := output["songs"].(map[string]any)["results"].([]map[string]any)
	albums := output["albums"].(map[string]any)["results"].([]map[string]any)
	singles := output["singles"].(map[string]any)["results"].([]map[string]any)
	videos := output["videos"].(map[string]any)["results"].([]map[string]any)
	result.AdditionalMeta["yt_tracks"] = tracks
	result.AdditionalMeta["yt_albums"] = albums
	result.AdditionalMeta["yt_singles"] = singles
	result.AdditionalMeta["yt_videos"] = videos
	result.AdditionalMeta["display_track_count"] = len(tracks)
	result.AdditionalMeta["display_album_count"] = len(albums)
	result.AdditionalMeta["display_single_count"] = len(singles)
	result.AdditionalMeta["display_video_count"] = len(videos)

	return result, nil
}

func (*YouTubeSource) completePlaylistMetadata(
	playable media.SourcePlayable,
	output map[string]any,
) (media.SourcePlayable, error) {
	result := playable.(media.Playlist)

	result.Description = output["description"].(string)

	yearObj := output["year"]
	if yearObj != nil {
		yearStr := yearObj.(string)
		result.CreationDate = yearStr
	}

	if config.Conf.General.InheritListenCounts {
		viewCountObj := output["views"]
		if viewCountObj != nil {
			viewCountStr := viewCountObj.(string)
			viewCount, err := strconv.Atoi(viewCountStr)
			if err != nil {
				return result, err
			}
			result.ListenCount = viewCount
		}
	}

	thumbnails := output["thumbnails"].([]map[string]any)
	thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
	thumbnail, err := storage.DownloadFile(thumbnailURL)
	if err != nil {
		return result, err
	}
	result.AdditionalMeta["display_cover_art"] = thumbnail

	result.AdditionalMeta["yt_tracks"] = output["tracks"].([]map[string]any)

	result.AdditionalMeta["display_track_count"] = output["trackCount"].(int)

	return result, nil
}

func init() {
	source, err := InitYouTubeSource()
	if err != nil {
		log.Warn("Source initialization failed", "source", source.GetID(), "error", err)
	} else {
		Registry[source.GetID()] = source
	}
}
