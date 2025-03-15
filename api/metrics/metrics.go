package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/libramusic/libracore/db"
)

var (
	TracksTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "libra_tracks_total",
		Help: "Total number of tracks stored in the Libra database.",
	})
	AlbumsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "libra_albums_total",
		Help: "Total number of albums stored in the Libra database.",
	})
	VideosTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "libra_videos_total",
		Help: "Total number of videos stored in the Libra database.",
	})
	ArtistsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "libra_artists_total",
		Help: "Total number of artists stored in the Libra database.",
	})
	PlaylistsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "libra_playlists_total",
		Help: "Total number of playlists stored in the Libra database.",
	})
	UsersTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "libra_users_total",
		Help: "Total number of users registered in the Libra database.",
	})
)

func RegisterMetrics() error {
	if err := prometheus.Register(TracksTotal); err != nil {
		return err
	}
	tracks, err := db.DB.GetAllTracks(context.Background())
	if err != nil {
		return err
	}
	TracksTotal.Add(float64(len(tracks)))

	if err := prometheus.Register(AlbumsTotal); err != nil {
		return err
	}
	albums, err := db.DB.GetAllAlbums(context.Background())
	if err != nil {
		return err
	}
	AlbumsTotal.Add(float64(len(albums)))

	if err := prometheus.Register(VideosTotal); err != nil {
		return err
	}
	videos, err := db.DB.GetAllVideos(context.Background())
	if err != nil {
		return err
	}
	VideosTotal.Add(float64(len(videos)))

	if err := prometheus.Register(ArtistsTotal); err != nil {
		return err
	}
	artists, err := db.DB.GetAllArtists(context.Background())
	if err != nil {
		return err
	}
	ArtistsTotal.Add(float64(len(artists)))

	if err := prometheus.Register(PlaylistsTotal); err != nil {
		return err
	}
	playlists, err := db.DB.GetAllPlaylists(context.Background())
	if err != nil {
		return err
	}
	PlaylistsTotal.Add(float64(len(playlists)))

	if err := prometheus.Register(UsersTotal); err != nil {
		return err
	}
	users, err := db.DB.GetUsers(context.Background())
	if err != nil {
		return err
	}
	UsersTotal.Add(float64(len(users)))

	return nil
}
