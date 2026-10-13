package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/libramusic/libracore"
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

func RegisterMetrics(l *libracore.Libra) error {
	if err := registerTotalCounter(TracksTotal, l.MediaDB.AllTracks); err != nil {
		return err
	}
	if err := registerTotalCounter(AlbumsTotal, l.MediaDB.AllAlbums); err != nil {
		return err
	}
	if err := registerTotalCounter(VideosTotal, l.MediaDB.AllVideos); err != nil {
		return err
	}
	if err := registerTotalCounter(ArtistsTotal, l.MediaDB.AllArtists); err != nil {
		return err
	}
	if err := registerTotalCounter(PlaylistsTotal, l.MediaDB.AllPlaylists); err != nil {
		return err
	}
	return registerTotalCounter(UsersTotal, l.AuthDB.Users)
}

func registerTotalCounter[T any](metric prometheus.Counter, fetchFunc func(context.Context) ([]T, error)) error {
	if err := prometheus.Register(metric); err != nil {
		return err
	}
	items, err := fetchFunc(context.Background())
	if err != nil {
		return err
	}
	metric.Add(float64(len(items)))
	return nil
}
