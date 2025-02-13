package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/libramusic/libracore/db"
)

var (
	TrackCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "libra_track_count",
		Help: "Total number of tracks stored in the Libra database.",
	})
	AlbumCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "libra_album_count",
		Help: "Total number of albums stored in the Libra database.",
	})
	VideoCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "libra_video_count",
		Help: "Total number of videos stored in the Libra database.",
	})
	ArtistCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "libra_artist_count",
		Help: "Total number of artists stored in the Libra database.",
	})
	PlaylistCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "libra_playlist_count",
		Help: "Total number of playlists stored in the Libra database.",
	})
	UserCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "libra_user_count",
		Help: "Total number of users registered in the Libra database.",
	})
)

func RegisterMetrics() error {
	if err := prometheus.Register(TrackCount); err != nil {
		return err
	}
	tracks, err := db.DB.GetAllTracks()
	if err != nil {
		return err
	}
	TrackCount.Set(float64(len(tracks)))

	if err := prometheus.Register(AlbumCount); err != nil {
		return err
	}
	albums, err := db.DB.GetAllAlbums()
	if err != nil {
		return err
	}
	AlbumCount.Set(float64(len(albums)))

	if err := prometheus.Register(VideoCount); err != nil {
		return err
	}
	videos, err := db.DB.GetAllVideos()
	if err != nil {
		return err
	}
	VideoCount.Set(float64(len(videos)))

	if err := prometheus.Register(ArtistCount); err != nil {
		return err
	}
	artists, err := db.DB.GetAllArtists()
	if err != nil {
		return err
	}
	ArtistCount.Set(float64(len(artists)))

	if err := prometheus.Register(PlaylistCount); err != nil {
		return err
	}
	playlists, err := db.DB.GetAllPlaylists()
	if err != nil {
		return err
	}
	PlaylistCount.Set(float64(len(playlists)))

	if err := prometheus.Register(UserCount); err != nil {
		return err
	}
	users, err := db.DB.GetUsers()
	if err != nil {
		return err
	}
	UserCount.Set(float64(len(users)))

	return nil
}
