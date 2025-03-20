package routes

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"

	"github.com/libramusic/libracore/db"
	"github.com/libramusic/libracore/storage"
)

// mimeType := mime.TypeByExtension(filepath.Ext(filePath))

//	@Summary	Get all playables
//	@ID			getAllPlayables
//	@Success	200	{array}	fakePlayable
//	@Success	200	"Returns a list of all playables"
//	@Failure	500	{object}	any
//	@Router		/playables [get]
func V1Playables(c echo.Context) error {
	ctx := c.Request().Context()

	playables, err := db.GetAllPlayables(ctx)
	if err != nil {
		log.Error("Error getting playables", "err", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "failed to retrieve playables"})
	}
	return c.JSON(http.StatusOK, echo.Map{"playables": playables})
}

//	@Summary	Get user's playables
//	@ID			getUserPlayables
//	@Param		id	path	string	true	"User ID"
//	@Success	200	{array}	fakePlayable
//	@Success	200	"Returns a list of user's playables"
//	@Failure	500	{object}	any
//	@Router		/playables/{id} [get]
func V1UserPlayables(c echo.Context) error {
	ctx := c.Request().Context()

	userID := c.Param("id")
	playables, err := db.GetPlayables(ctx, userID)
	if err != nil {
		log.Error("Error getting playables", "err", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "failed to retrieve playables"})
	}
	return c.JSON(http.StatusOK, echo.Map{"playables": playables})
}

//	@Summary	Search for playables by query
//	@ID			searchPlayables
//	@Param		q	query	string	true	"Search query"
//	@Success	200	{array}	fakePlayable
//	@Success	200	"Returns a list of playables matching the search query"
//	@Failure	500	{object}	any
//	@Router		/search [get]
func V1Search(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

// START TO REFACTOR

func V1Track(c echo.Context) error {
	ctx := c.Request().Context()

	trackID := c.Param("id")
	track, err := db.DB.GetTrack(ctx, trackID)
	if err != nil {
		log.Error("Error getting track", "err", err, "trackID", trackID)
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "failed to retrieve track"})
	}
	return c.JSON(http.StatusOK, track)
}

func V1TrackIsStored(c echo.Context) error {
	trackID := c.Param("id")
	return c.JSON(http.StatusOK, echo.Map{"stored": storage.IsContentStored("track", trackID)})
}

func V1TrackStream(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1TrackCover(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1TrackLyrics(c echo.Context) error {
	ctx := c.Request().Context()

	trackID := c.Param("id")
	track, err := db.DB.GetTrack(ctx, trackID)
	if err != nil {
		log.Error("Error getting track", "err", err, "trackID", trackID)
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "failed to retrieve track"})
	}
	return c.JSON(http.StatusOK, track.Lyrics)
}

func V1TrackLyricsLang(c echo.Context) error {
	ctx := c.Request().Context()

	trackID := c.Param("id")
	track, err := db.DB.GetTrack(ctx, trackID)
	if err != nil {
		log.Error("Error getting track", "err", err, "trackID", trackID)
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "failed to retrieve track"})
	}

	lang := c.Param("lang")
	if lyrics, ok := track.Lyrics[lang]; ok {
		return c.String(http.StatusOK, lyrics)
	}
	return c.NoContent(http.StatusNotFound)
}

func V1Album(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1AlbumCover(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1AlbumTracks(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1Video(c echo.Context) error {
	ctx := c.Request().Context()

	videoID := c.Param("id")
	video, err := db.DB.GetVideo(ctx, videoID)
	if err != nil {
		log.Error("Error getting video", "err", err, "videoID", videoID)
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "failed to retrieve video"})
	}
	return c.JSON(http.StatusOK, video)
}

func V1VideoIsStored(c echo.Context) error {
	videoID := c.Param("id")
	return c.JSON(http.StatusOK, echo.Map{"stored": storage.IsContentStored("video", videoID)})
}

func V1VideoStream(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1VideoCover(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1VideoSubtitles(c echo.Context) error {
	ctx := c.Request().Context()

	videoID := c.Param("id")
	video, err := db.DB.GetVideo(ctx, videoID)
	if err != nil {
		log.Error("Error getting video", "err", err, "videoID", videoID)
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "failed to retrieve video"})
	}
	return c.JSON(http.StatusOK, video.Subtitles)
}

func V1VideoSubtitlesLang(c echo.Context) error {
	ctx := c.Request().Context()

	videoID := c.Param("id")
	video, err := db.DB.GetVideo(ctx, videoID)
	if err != nil {
		log.Error("Error getting video", "err", err, "videoID", videoID)
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "failed to retrieve video"})
	}
	lang := c.Param("lang")
	if subtitles, ok := video.Subtitles[lang]; ok {
		return c.String(http.StatusOK, subtitles)
	}
	return c.NoContent(http.StatusNotFound)
}

func V1Playlist(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1PlaylistCover(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1PlaylistTracks(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1Artist(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1ArtistCover(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1ArtistAlbums(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

func V1ArtistTracks(c echo.Context) error {
	log.Error("unimplemented")
	return c.JSON(http.StatusOK, echo.Map{})
}

// END TO REFACTOR
