package source

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DevReaper0/libra/config"
	"github.com/DevReaper0/libra/types"
	"github.com/DevReaper0/libra/util"
)

type YouTubeSource struct {
	manager Manager
}

func InitYouTubeSource(manager Manager) (YouTubeSource, error) {
	if _, err := os.Stat(config.Conf.SourceScripts.YouTubeLocation); os.IsNotExist(err) {
		err = util.DownloadFileTo(config.Conf.SourceScripts.YouTubeURL, config.Conf.SourceScripts.YouTubeLocation)
		if err != nil {
			return YouTubeSource{}, err
		}
	}

	return YouTubeSource{
		manager: manager,
	}, nil
}

func (s YouTubeSource) GetID() string {
	return "youtube"
}

func (s YouTubeSource) GetName() string {
	return "YouTube"
}

func (s YouTubeSource) GetVersion() string {
	return util.LibraVersion
}

func (s YouTubeSource) GetSourceTypes() []string {
	return []string{"content", "metadata", "lyrics"}
}

func (s YouTubeSource) GetMediaTypes() []string {
	return []string{"music", "video", "playlist"}
}

func (s YouTubeSource) Search(query string, limit int, page int, filters map[string]string) ([]types.SourcePlayable, error) {
	var results []types.SourcePlayable

	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return results, err
	}

	command := append(strings.Split(config.Conf.SourceScripts.PythonCommand, " "), []string{
		config.Conf.SourceScripts.YouTubeLocation,
		`search`,
		fmt.Sprintf(`query="%s"`, strings.ReplaceAll(query, `"`, `\"`)),
		fmt.Sprintf(`limit=%d`, limit),
		fmt.Sprintf(`filters="%s"`, strings.ReplaceAll(string(filtersJSON), `"`, `\"`)),
	}...)
	out, err := util.ExecCommand(command)
	if err != nil {
		return results, err
	}

	var output []map[string]interface{}
	err = json.Unmarshal(out, &output)
	if err != nil {
		return results, err
	}

	for _, v := range output {
		var result types.SourcePlayable
		switch v["resultType"] {
		case "song":
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

			thumbnails := v["thumbnails"].([]map[string]interface{})
			thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
			thumbnail, err := util.DownloadFile(thumbnailURL)
			if err != nil {
				return results, err
			}

			result = types.Track{
				Title:       v["title"].(string),
				Duration:    v["duration_seconds"].(int),
				ReleaseDate: year,
				AdditionalMeta: map[string]interface{}{
					"display_artists":   displayArtists,
					"display_album":     v["album"].(map[string]string)["name"],
					"display_cover_art": thumbnail,
					"yt_id":             v["videoId"].(string),
					"yt_artists":        v["artists"].([]map[string]string),
					"yt_album":          v["album"].(map[string]string),
				},
				MetadataSource: types.LinkedSource(s.GetID() + "::" + "https://music.youtube.com/watch?v=" + v["videoId"].(string)),
			}
		case "album":
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

			thumbnails := v["thumbnails"].([]map[string]interface{})
			thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
			thumbnail, err := util.DownloadFile(thumbnailURL)
			if err != nil {
				return results, err
			}

			result = types.Album{
				Title:       v["title"].(string),
				ReleaseDate: year,
				AdditionalMeta: map[string]interface{}{
					"display_artists":   displayArtists,
					"display_cover_art": thumbnail,
					"yt_id":             v["browseId"].(string),
					"yt_artists":        v["artists"].([]map[string]string),
				},
				MetadataSource: types.LinkedSource(s.GetID() + "::" + "https://music.youtube.com/browse/" + v["browseId"].(string)),
			}
		case "video":
			if !config.Conf.General.IncludeVideoResults {
				fmt.Println(types.SourceError{SourceID: s.GetID(), Err: types.UnsupportedMediaTypeError{MediaType: v["resultType"].(string)}})
				continue
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

				thumbnails := v["thumbnails"].([]map[string]interface{})
				thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
				thumbnail, err := util.DownloadFile(thumbnailURL)
				if err != nil {
					return results, err
				}

				result = types.Track{
					Title:       v["title"].(string),
					Duration:    v["duration_seconds"].(int),
					ReleaseDate: year,
					AdditionalMeta: map[string]interface{}{
						"display_artists":   displayArtists,
						"display_album":     album,
						"display_cover_art": thumbnail,
						"is_video":          true,
						"yt_id":             v["videoId"].(string),
						"yt_artists":        v["artists"].([]map[string]string),
						"yt_album":          album,
					},
					MetadataSource: types.LinkedSource(s.GetID() + "::" + "https://music.youtube.com/watch?v=" + v["videoId"].(string)),
				}
			} else {
				thumbnails := v["thumbnails"].([]map[string]interface{})
				thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
				thumbnail, err := util.DownloadFile(thumbnailURL)
				if err != nil {
					return results, err
				}

				result = types.Video{
					Title:       v["title"].(string),
					Duration:    v["duration_seconds"].(int),
					ReleaseDate: year,
					AdditionalMeta: map[string]interface{}{
						"display_artists":   displayArtists,
						"display_thumbnail": thumbnail,
						"yt_id":             v["videoId"].(string),
						"yt_artists":        v["artists"].([]map[string]string),
					},
					MetadataSource: types.LinkedSource(s.GetID() + "::" + "https://www.youtube.com/watch?v=" + v["videoId"].(string)),
				}
			}
		case "artist":
			thumbnails := v["thumbnails"].([]map[string]interface{})
			thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
			thumbnail, err := util.DownloadFile(thumbnailURL)
			if err != nil {
				return results, err
			}

			result = types.Artist{
				Name: v["artist"].(string),
				AdditionalMeta: map[string]interface{}{
					"display_cover_art": thumbnail,
					"yt_id":             v["browseId"].(string),
				},
				MetadataSource: types.LinkedSource(s.GetID() + "::" + "https://music.youtube.com/channel/" + v["browseId"].(string)),
			}
		case "playlist":
			var displayArtists []string
			for _, artist := range v["artists"].([]map[string]string) {
				displayArtists = append(displayArtists, artist["name"])
			}

			thumbnails := v["thumbnails"].([]map[string]interface{})
			thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
			thumbnail, err := util.DownloadFile(thumbnailURL)
			if err != nil {
				return results, err
			}

			result = types.Playlist{
				Title: v["title"].(string),
				AdditionalMeta: map[string]interface{}{
					"display_artists":   displayArtists,
					"display_cover_art": thumbnail,
					"yt_id":             v["browseId"].(string),
					"yt_artists":        v["artists"].([]map[string]string),
				},
				MetadataSource: types.LinkedSource(s.GetID() + "::" + "https://music.youtube.com/playlist?list=" + v["browseId"].(string)),
			}
		default:
			fmt.Println(types.SourceError{SourceID: s.GetID(), Err: types.UnsupportedMediaTypeError{MediaType: v["resultType"].(string)}})
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

func (s YouTubeSource) GetContent(playable types.SourcePlayable) ([]byte, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return nil, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	var command []string
	switch playable.GetType() {
	case "track":
		command = append(strings.Split(config.Conf.SourceScripts.PythonCommand, " "), []string{
			config.Conf.SourceScripts.YouTubeLocation,
			`content`,
			`type=audio`,
			`id=` + playable.GetAdditionalMeta()["yt_id"].(string),
		}...)
	case "video":
		command = append(strings.Split(config.Conf.SourceScripts.PythonCommand, " "), []string{
			config.Conf.SourceScripts.YouTubeLocation,
			`content`,
			`type=video`,
			`id=` + playable.GetAdditionalMeta()["yt_id"].(string),
		}...)
	default:
		return nil, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	out, err := util.ExecCommand(command)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s YouTubeSource) GetLyrics(playable types.LyricsPlayable) (map[string]string, error) {
	result := map[string]string{}

	if !SupportsMediaType(s, playable.GetType()) {
		return result, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	var command []string
	if playable.GetType() == "video" || playable.GetAdditionalMeta()["is_video"] == true {
		command = append(strings.Split(config.Conf.SourceScripts.PythonCommand, " "), []string{
			config.Conf.SourceScripts.YouTubeLocation,
			`subtitles`,
			`id=` + playable.GetAdditionalMeta()["yt_id"].(string),
		}...)
	} else {
		command = append(strings.Split(config.Conf.SourceScripts.PythonCommand, " "), []string{
			config.Conf.SourceScripts.YouTubeLocation,
			`lyrics`,
			`id=` + playable.GetAdditionalMeta()["yt_id"].(string),
		}...)
	}

	out, err := util.ExecCommand(command)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(out, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (s YouTubeSource) CompleteMetadata(playable types.SourcePlayable) (types.SourcePlayable, error) {
	if !SupportsMediaType(s, playable.GetType()) {
		return playable, types.UnsupportedMediaTypeError{MediaType: playable.GetType()}
	}

	command := append(strings.Split(config.Conf.SourceScripts.PythonCommand, " "), []string{
		config.Conf.SourceScripts.YouTubeLocation,
		playable.GetType(),
		`id=` + playable.GetAdditionalMeta()["yt_id"].(string),
	}...)

	out, err := util.ExecCommand(command)
	if err != nil {
		return playable, err
	}

	var output map[string]interface{}
	err = json.Unmarshal(out, &output)
	if err != nil {
		return playable, err
	}

	switch playable.GetType() {
	case "track":
		result := playable.(types.Track)

		lyricsID := output["track"].(map[string]interface{})["lyricsId"]
		if lyricsID != nil {
			splitLyricsID := strings.Split(lyricsID.(string), "-")
			trackNumber, err := strconv.Atoi(splitLyricsID[len(splitLyricsID)-1])
			if err == nil {
				result.TrackNumber = trackNumber
			}
		}

		result.Description = output["video"].(map[string]interface{})["microformat"].(map[string]interface{})["microformatDataRenderer"].(map[string]interface{})["description"].(string)

		publishDateObj := output["video"].(map[string]interface{})["microformat"].(map[string]interface{})["microformatDataRenderer"].(map[string]interface{})["publishDate"]
		if publishDateObj != nil {
			publishDateStr := publishDateObj.(string)
			t, err := time.Parse(time.RFC3339, publishDateStr)
			if err != nil {
				return result, err
			}
			result.ReleaseDate = t.Format(time.DateTime)
		}

		if config.Conf.General.InheritListenCounts {
			viewCountObj := output["video"].(map[string]interface{})["microformat"].(map[string]interface{})["microformatDataRenderer"].(map[string]interface{})["viewCount"]
			if viewCountObj != nil {
				viewCountStr := viewCountObj.(string)
				viewCount, err := strconv.Atoi(viewCountStr)
				if err != nil {
					return result, err
				}
				result.ListenCount = viewCount
			}
		}

		thumbnails := output["track"].(map[string]interface{})["thumbnail"].([]map[string]interface{})
		thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
		thumbnail, err := util.DownloadFile(thumbnailURL)
		if err != nil {
			return result, err
		}
		result.AdditionalMeta["display_cover_art"] = thumbnail
	case "album":
		result := playable.(types.Album)

		result.Description = output["description"].(string)

		thumbnails := output["thumbnails"].([]map[string]interface{})
		thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
		thumbnail, err := util.DownloadFile(thumbnailURL)
		if err != nil {
			return result, err
		}
		result.AdditionalMeta["display_cover_art"] = thumbnail

		result.AdditionalMeta["yt_tracks"] = output["tracks"].([]map[string]interface{})

		result.AdditionalMeta["display_track_count"] = output["trackCount"].(int)
	case "video":
		result := playable.(types.Video)

		result.Description = output["video"].(map[string]interface{})["microformat"].(map[string]interface{})["microformatDataRenderer"].(map[string]interface{})["description"].(string)

		publishDateObj := output["video"].(map[string]interface{})["microformat"].(map[string]interface{})["microformatDataRenderer"].(map[string]interface{})["publishDate"]
		if publishDateObj != nil {
			publishDateStr := publishDateObj.(string)
			t, err := time.Parse(time.RFC3339, publishDateStr)
			if err != nil {
				return result, err
			}
			result.ReleaseDate = t.Format(time.DateTime)
		}

		if config.Conf.General.InheritListenCounts {
			viewCountObj := output["video"].(map[string]interface{})["microformat"].(map[string]interface{})["microformatDataRenderer"].(map[string]interface{})["viewCount"]
			if viewCountObj != nil {
				viewCountStr := viewCountObj.(string)
				viewCount, err := strconv.Atoi(viewCountStr)
				if err != nil {
					return result, err
				}
				result.WatchCount = viewCount
			}
		}

		thumbnails := output["track"].(map[string]interface{})["thumbnail"].([]map[string]interface{})
		thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
		thumbnail, err := util.DownloadFile(thumbnailURL)
		if err != nil {
			return result, err
		}
		result.AdditionalMeta["display_thumbnail"] = thumbnail
	case "artist":
		result := playable.(types.Artist)

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

		thumbnails := output["thumbnails"].([]map[string]interface{})
		thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
		thumbnail, err := util.DownloadFile(thumbnailURL)
		if err != nil {
			return result, err
		}
		result.AdditionalMeta["display_cover_art"] = thumbnail

		tracks := output["songs"].(map[string]interface{})["results"].([]map[string]interface{})
		albums := output["albums"].(map[string]interface{})["results"].([]map[string]interface{})
		singles := output["singles"].(map[string]interface{})["results"].([]map[string]interface{})
		videos := output["videos"].(map[string]interface{})["results"].([]map[string]interface{})
		result.AdditionalMeta["yt_tracks"] = tracks
		result.AdditionalMeta["yt_albums"] = albums
		result.AdditionalMeta["yt_singles"] = singles
		result.AdditionalMeta["yt_videos"] = videos
		result.AdditionalMeta["display_track_count"] = len(tracks)
		result.AdditionalMeta["display_album_count"] = len(albums)
		result.AdditionalMeta["display_single_count"] = len(singles)
		result.AdditionalMeta["display_video_count"] = len(videos)
	case "playlist":
		result := playable.(types.Playlist)

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

		thumbnails := output["thumbnails"].([]map[string]interface{})
		thumbnailURL := thumbnails[len(thumbnails)-1]["url"].(string)
		thumbnail, err := util.DownloadFile(thumbnailURL)
		if err != nil {
			return result, err
		}
		result.AdditionalMeta["display_cover_art"] = thumbnail

		result.AdditionalMeta["yt_tracks"] = output["tracks"].([]map[string]interface{})

		result.AdditionalMeta["display_track_count"] = output["trackCount"].(int)
	}

	return playable, nil
}
