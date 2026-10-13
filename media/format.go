package media

type Format string

const (
	FormatMusic    Format = "music" // Music is equivalent to Track + Album + Artist.
	FormatTrack    Format = "track"
	FormatAlbum    Format = "album"
	FormatArtist   Format = "artist"
	FormatVideo    Format = "video"
	FormatPlaylist Format = "playlist"
	FormatUser     Format = "user"
)

func (f Format) Matches(other Format) bool {
	if f == other {
		return true
	}
	if (f == FormatTrack || f == FormatAlbum || f == FormatArtist) && other == FormatMusic {
		return true
	}
	return false
}
