package routes

// fakePlayable is a stand-in for the Playable interface.
// It's schema will be replaced with one for the actual Playable interface during initialization.
type fakePlayable struct{}

var _ fakePlayable

// The fake route has all the implementations of Playable as possible responses so their schemas will be generated.

//	@Success	200	{object}	types.Track
//	@Success	200	{object}	types.Album
//	@Success	200	{object}	types.Video
//	@Success	200	{object}	types.Playlist
//	@Success	200	{object}	types.Artist
//	@Success	200	{object}	types.User
//	@Router		/fake [get]
func _() {}