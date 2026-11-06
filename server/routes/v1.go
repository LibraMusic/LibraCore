package routes

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v5"

	"github.com/libramusic/libracore"
	"github.com/libramusic/libracore/db"
	"github.com/libramusic/libracore/storage"
)

// @Summary	Get all playables
// @ID			getAllPlayables
// @Success	200	{array}	fakePlayable
// @Success	200	"Returns a list of all playables"
// @Failure	500	{object}	any
// @Router		/playables [get]
func V1Playables(c *echo.Context) error {
	ctx := c.Request().Context()
	l := c.Get("libra").(*libracore.Libra)

	playables, err := db.AllPlayables(ctx, l.MediaDB)
	if err != nil {
		log.Error("Error getting playables", "err", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": "failed to retrieve playables"})
	}
	return c.JSON(http.StatusOK, map[string]any{"playables": playables})
}

// @Summary	Get user's playables
// @ID			getUserPlayables
// @Param		id	path	string	true	"User ID"
// @Success	200	{array}	fakePlayable
// @Success	200	"Returns a list of user's playables"
// @Failure	500	{object}	any
// @Router		/playables/{id} [get]
func V1UserPlayables(c *echo.Context) error {
	ctx := c.Request().Context()
	l := c.Get("libra").(*libracore.Libra)

	userID := c.Param("id")
	playables, err := db.Playables(ctx, l.MediaDB, userID)
	if err != nil {
		log.Error("Error getting playables", "err", err)
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": "failed to retrieve playables"})
	}
	return c.JSON(http.StatusOK, map[string]any{"playables": playables})
}

// @Summary	Search for playables by query
// @ID			searchPlayables
// @Param		q	query	string	true	"Search query"
// @Success	200	{array}	fakePlayable
// @Success	200	"Returns a list of playables matching the search query"
// @Failure	500	{object}	any
// @Router		/search [get]
func V1Search(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

// START TO REFACTOR

func V1Track(c *echo.Context) error {
	ctx := c.Request().Context()
	l := c.Get("libra").(*libracore.Libra)

	trackID := c.Param("id")
	track, err := l.MediaDB.Track(ctx, trackID)
	if err != nil {
		log.Error("Error getting track", "err", err, "trackID", trackID)
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": "failed to retrieve track"})
	}
	return c.JSON(http.StatusOK, track)
}

func V1TrackIsStored(c *echo.Context) error {
	l := c.Get("libra").(*libracore.Libra)

	trackID := c.Param("id")
	return c.JSON(
		http.StatusOK,
		map[string]any{"stored": storage.IsContentStored(l.Config.Storage, "track", trackID)},
	)
}

func V1TrackStream(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1TrackCover(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1TrackLyrics(c *echo.Context) error {
	ctx := c.Request().Context()
	l := c.Get("libra").(*libracore.Libra)

	trackID := c.Param("id")
	track, err := l.MediaDB.Track(ctx, trackID)
	if err != nil {
		log.Error("Error getting track", "err", err, "trackID", trackID)
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": "failed to retrieve track"})
	}
	return c.JSON(http.StatusOK, track.Lyrics)
}

func V1TrackLyricsLang(c *echo.Context) error {
	ctx := c.Request().Context()
	l := c.Get("libra").(*libracore.Libra)

	trackID := c.Param("id")
	track, err := l.MediaDB.Track(ctx, trackID)
	if err != nil {
		log.Error("Error getting track", "err", err, "trackID", trackID)
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": "failed to retrieve track"})
	}

	lang := c.Param("lang")
	if lyrics, ok := track.Lyrics[lang]; ok {
		return c.String(http.StatusOK, lyrics)
	}
	return c.NoContent(http.StatusNotFound)
}

func V1Album(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1AlbumCover(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1AlbumTracks(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1Video(c *echo.Context) error {
	ctx := c.Request().Context()
	l := c.Get("libra").(*libracore.Libra)

	videoID := c.Param("id")
	video, err := l.MediaDB.Video(ctx, videoID)
	if err != nil {
		log.Error("Error getting video", "err", err, "videoID", videoID)
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": "failed to retrieve video"})
	}
	return c.JSON(http.StatusOK, video)
}

func V1VideoIsStored(c *echo.Context) error {
	l := c.Get("libra").(*libracore.Libra)

	videoID := c.Param("id")
	return c.JSON(
		http.StatusOK,
		map[string]any{"stored": storage.IsContentStored(l.Config.Storage, "video", videoID)},
	)
}

func V1VideoStream(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1VideoCover(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1VideoSubtitles(c *echo.Context) error {
	ctx := c.Request().Context()
	l := c.Get("libra").(*libracore.Libra)

	videoID := c.Param("id")
	video, err := l.MediaDB.Video(ctx, videoID)
	if err != nil {
		log.Error("Error getting video", "err", err, "videoID", videoID)
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": "failed to retrieve video"})
	}
	return c.JSON(http.StatusOK, video.Subtitles)
}

func V1VideoSubtitlesLang(c *echo.Context) error {
	ctx := c.Request().Context()
	l := c.Get("libra").(*libracore.Libra)

	videoID := c.Param("id")
	video, err := l.MediaDB.Video(ctx, videoID)
	if err != nil {
		log.Error("Error getting video", "err", err, "videoID", videoID)
		return c.JSON(http.StatusInternalServerError, map[string]any{"message": "failed to retrieve video"})
	}
	lang := c.Param("lang")
	if subtitles, ok := video.Subtitles[lang]; ok {
		return c.String(http.StatusOK, subtitles)
	}
	return c.NoContent(http.StatusNotFound)
}

func V1Playlist(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1PlaylistCover(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1PlaylistTracks(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1Artist(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1ArtistCover(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1ArtistAlbums(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

func V1ArtistTracks(c *echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, map[string]any{})
}

// END TO REFACTOR
