package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/LibraMusic/LibraCore/db"
	"github.com/LibraMusic/LibraCore/logging"
	"github.com/LibraMusic/LibraCore/storage"
)

// mimeType := mime.TypeByExtension(filepath.Ext(filePath))

func V1Playables(c *fiber.Ctx) error {
	playables, err := db.GetAllPlayables()
	if err != nil {
		logging.Error().Err(err).Msg("Error getting playables")
	}
	return c.JSON(fiber.Map{"playables": playables})
}

func V1PlayablesUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	playables, err := db.GetPlayables(userID)
	if err != nil {
		logging.Error().Err(err).Msg("Error getting playables")
	}
	return c.JSON(fiber.Map{"playables": playables})
}

func V1Search(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

// START TO REFACTOR

func V1Track(c *fiber.Ctx) error {
	trackID := c.Params("id")
	track, err := db.DB.GetTrack(trackID)
	if err != nil {
		logging.Error().Err(err).Msgf("Error getting track %s", trackID)
	}
	return c.JSON(track)
}

func V1TrackIsStored(c *fiber.Ctx) error {
	trackID := c.Params("id")
	return c.JSON(fiber.Map{"stored": storage.IsContentStored("track", trackID)})
}

func V1TrackStream(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	// https://pcpratheesh.medium.com/streaming-video-with-golang-fiber-a-practical-tutorial-a2170584ae9f
	// https://pkg.go.dev/bytes#NewReader
	// https://pkg.go.dev/io#Copy
	// https://pkg.go.dev/io#CopyN
	// c.Response().BodyWriter()
	return c.JSON(fiber.Map{})
}

func V1TrackCover(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

func V1TrackLyrics(c *fiber.Ctx) error {
	trackID := c.Params("id")
	track, err := db.DB.GetTrack(trackID)
	if err != nil {
		logging.Error().Err(err).Msgf("Error getting track %s", trackID)
	}
	return c.JSON(track.Lyrics)
}

func V1TrackLyricsLang(c *fiber.Ctx) error {
	trackID := c.Params("id")
	track, err := db.DB.GetTrack(trackID)
	if err != nil {
		logging.Error().Err(err).Msgf("Error getting track %s", trackID)
	}

	lang := c.Params("lang")
	if lyrics, ok := track.Lyrics[lang]; ok {
		return c.SendString(lyrics)
	}
	return c.SendStatus(fiber.StatusNotFound)
}

func V1Album(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

func V1AlbumCover(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

func V1AlbumTracks(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

func V1Video(c *fiber.Ctx) error {
	videoID := c.Params("id")
	video, err := db.DB.GetVideo(videoID)
	if err != nil {
		logging.Error().Err(err).Msgf("Error getting video %s", videoID)
	}
	return c.JSON(video)
}

func V1VideoIsStored(c *fiber.Ctx) error {
	videoID := c.Params("id")
	return c.JSON(fiber.Map{"stored": storage.IsContentStored("video", videoID)})
}

func V1VideoStream(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

func V1VideoCover(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

func V1VideoSubtitles(c *fiber.Ctx) error {
	videoID := c.Params("id")
	video, err := db.DB.GetVideo(videoID)
	if err != nil {
		logging.Error().Err(err).Msgf("Error getting video %s", videoID)
	}
	return c.JSON(video.Subtitles)
}

func V1VideoSubtitlesLang(c *fiber.Ctx) error {
	videoID := c.Params("id")
	video, err := db.DB.GetVideo(videoID)
	if err != nil {
		logging.Error().Err(err).Msgf("Error getting video %s", videoID)
	}
	lang := c.Params("lang")
	if subtitles, ok := video.Subtitles[lang]; ok {
		return c.SendString(subtitles)
	}
	return c.SendStatus(fiber.StatusNotFound)
}

func V1Playlist(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

func V1PlaylistCover(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

func V1PlaylistTracks(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

func V1Artist(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

func V1ArtistCover(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

func V1ArtistAlbums(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

func V1ArtistTracks(c *fiber.Ctx) error {
	logging.Error().Msg("unimplemented")
	return c.JSON(fiber.Map{})
}

// END TO REFACTOR
