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
	"github.com/goccy/go-json"

	"github.com/libramusic/libracore/media"
	"github.com/libramusic/libracore/storage"
	"github.com/libramusic/libracore/versions"
)

//go:embed scripts/youtube.py
var youtubeScript string

type YouTubeFactory struct{}

func (YouTubeFactory) ID() string {
	return "youtube"
}

func (YouTubeFactory) Satisfies(id string) bool {
	return slices.Contains([]string{
		"youtube",
		"yt",
	}, strings.ToLower(id))
}

func (YouTubeFactory) SupportsMultiple() bool {
	return false
}

func (YouTubeFactory) New(alias string, options map[string]any, info SharedInfo) (Source, error) {
	opts := parseYouTubeOptions(options)
	source := &YouTube{alias: alias, options: opts, info: info}

	// Ensure script exists at script path.
	youtubeLocation := source.scriptPath()
	if _, err := os.Stat(youtubeLocation); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(youtubeLocation), os.ModePerm); err != nil {
			return source, err
		}
		if err := os.WriteFile(youtubeLocation, []byte(youtubeScript), 0o644); err != nil {
			return source, err
		}
	}

	return source, nil
}

type YouTubeOptions struct {
	PythonCommand  string
	ScriptPath     string
	VideoAudioOnly bool
}

func parseYouTubeOptions(options map[string]any) YouTubeOptions {
	opts := YouTubeOptions{
		PythonCommand:  "python3",
		ScriptPath:     "source_scripts/youtube.py",
		VideoAudioOnly: true,
	}

	if pythonCommand, ok := options["python_command"].(string); ok {
		opts.PythonCommand = pythonCommand
	}
	if scriptPath, ok := options["script_path"].(string); ok {
		opts.ScriptPath = scriptPath
	}
	if videoAudioOnly, ok := options["video_audio_only"].(bool); ok {
		opts.VideoAudioOnly = videoAudioOnly
	}

	return opts
}

type YouTube struct {
	alias   string
	options YouTubeOptions
	info    SharedInfo
}

func (s *YouTube) scriptPath() string {
	path := s.options.ScriptPath
	if path == "" {
		path = "source_scripts/youtube.py"
	}
	if s.info.DataDir != "" && !filepath.IsAbs(path) {
		absDataDir, err := filepath.Abs(s.info.DataDir)
		if err != nil {
			return filepath.Join(s.info.DataDir, path)
		}
		return filepath.Join(absDataDir, path)
	}
	return path
}

func (*YouTube) ID() string {
	return "youtube"
}

func (*YouTube) Name() string {
	return "YouTube"
}

func (*YouTube) Version() *semver.Version {
	return versions.Libra
}

func (*YouTube) Categories() []Category {
	return []Category{CategoryMedia, CategoryMetadata, CategoryLyrics}
}

func (*YouTube) Formats() []media.Format {
	return []media.Format{media.FormatMusic, media.FormatVideo, media.FormatPlaylist}
}

func (s *YouTube) Search(query string, limit, _ int, filter Filter, _ BindingInfo) ([]media.SourcePlayable, error) {
	var results []media.SourcePlayable

	// TODO: Implement pagination if possible.

	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return results, fmt.Errorf("error encoding filters: %w", err)
	}

	youtubeLocation := s.scriptPath()

	command := append(strings.Split(s.options.PythonCommand, " "), []string{
		youtubeLocation,
		"search",
		fmt.Sprintf("query=%s", query),
		fmt.Sprintf("limit=%d", limit),
		fmt.Sprintf("filters=%s", filterJSON),
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

func (s *YouTube) parseSearchResult(v map[string]any) (media.SourcePlayable, error) {
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
		err = ErrUnsupportedFormat
	}

	return result, err
}

func (s *YouTube) parseSongResult(v map[string]any) (media.SourcePlayable, error) {
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
			"url":               "https://music.youtube.com/watch?v=" + v["videoId"].(string),
		},
		MetadataSource: s.alias, // TODO?
	}, nil
}

func (s *YouTube) parseAlbumResult(v map[string]any) (media.SourcePlayable, error) {
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
			"url":               "https://music.youtube.com/browse/" + v["browseId"].(string),
		},
		MetadataSource: s.alias, // TODO?
	}, nil
}

func (s *YouTube) parseVideoResult(v map[string]any) (media.SourcePlayable, error) {
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

	if s.options.VideoAudioOnly {
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
				"url":               "https://music.youtube.com/watch?v=" + v["videoId"].(string),
			},
			MetadataSource: s.alias, // TODO?
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
			"url":               "https://www.youtube.com/watch?v=" + v["videoId"].(string),
		},
		MetadataSource: s.alias, // TODO?
	}, nil
}

func (s *YouTube) parseArtistResult(v map[string]any) (media.SourcePlayable, error) {
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
			"url":               "https://music.youtube.com/channel/" + v["browseId"].(string),
		},
		MetadataSource: s.alias, // TODO?
	}, nil
}

func (s *YouTube) parsePlaylistResult(v map[string]any) (media.SourcePlayable, error) {
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
			"url":               "https://music.youtube.com/playlist?list=" + v["browseId"].(string),
		},
		MetadataSource: s.alias, // TODO?
	}, nil
}

func (s *YouTube) Media(playable media.SourcePlayable, _ BindingInfo) ([]byte, error) {
	var command []string
	youtubeLocation := s.scriptPath()

	switch playable.GetFormat() {
	case media.FormatTrack:
		command = append(strings.Split(s.options.PythonCommand, " "), []string{
			youtubeLocation,
			"media",
			"format=audio",
			fmt.Sprintf("id=%s", playable.GetAdditionalMeta()["yt_id"]),
		}...)
	case media.FormatVideo:
		command = append(strings.Split(s.options.PythonCommand, " "), []string{
			youtubeLocation,
			"media",
			"format=video",
			fmt.Sprintf("id=%s", playable.GetAdditionalMeta()["yt_id"]),
		}...)
	default:
		return nil, ErrUnsupportedFormat
	}

	out, err := execCommand(command)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *YouTube) Lyrics(playable media.LyricsPlayable, _ BindingInfo) (map[string]string, error) {
	result := map[string]string{}

	var command []string
	youtubeLocation := s.scriptPath()

	if playable.GetFormat() == media.FormatVideo ||
		playable.GetAdditionalMeta()["is_video"] == true { //revive:disable-line:bool-literal-in-expr Value cannot be used as a boolean
		command = append(strings.Split(s.options.PythonCommand, " "), []string{
			youtubeLocation,
			"subtitles",
			fmt.Sprintf("id=%s", playable.GetAdditionalMeta()["yt_id"]),
		}...)
	} else {
		command = append(strings.Split(s.options.PythonCommand, " "), []string{
			youtubeLocation,
			"lyrics",
			fmt.Sprintf("id=%s", playable.GetAdditionalMeta()["yt_id"]),
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

func (s *YouTube) CompleteMetadata(playable media.SourcePlayable, _ BindingInfo) (media.SourcePlayable, error) {
	youtubeLocation := s.scriptPath()

	command := append(strings.Split(s.options.PythonCommand, " "), []string{
		youtubeLocation,
		string(playable.GetFormat()),
		fmt.Sprintf("id=%s", playable.GetAdditionalMeta()["yt_id"]),
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

	switch playable.GetFormat() {
	case media.FormatTrack:
		return s.completeTrackMetadata(playable, output)
	case media.FormatAlbum:
		return s.completeAlbumMetadata(playable, output)
	case media.FormatVideo:
		return s.completeVideoMetadata(playable, output)
	case media.FormatArtist:
		return s.completeArtistMetadata(playable, output)
	case media.FormatPlaylist:
		return s.completePlaylistMetadata(playable, output)
	default:
		return playable, nil
	}
}

func (s *YouTube) completeTrackMetadata(
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

	if s.info.InheritListenCounts {
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

func (*YouTube) completeAlbumMetadata(
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

func (s *YouTube) completeVideoMetadata(
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

	if s.info.InheritListenCounts {
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

func (s *YouTube) completeArtistMetadata(
	playable media.SourcePlayable,
	output map[string]any,
) (media.SourcePlayable, error) {
	result := playable.(media.Artist)

	result.Description = output["description"].(string)

	if s.info.InheritListenCounts && !s.info.ArtistListenCountsByTrack {
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

func (s *YouTube) completePlaylistMetadata(
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

	if s.info.InheritListenCounts {
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
	Register(YouTubeFactory{})
}
