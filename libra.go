package libracore

import (
	"errors"
	"fmt"

	"github.com/libramusic/libracore/config"
	"github.com/libramusic/libracore/db"
)

// Libra represents a Libra instance that holds core components such as database connections.
// The database connections aren't necessarily separate, they could point to the same underlying database instance.
type Libra struct {
	Config  config.Config
	MediaDB db.Database
	AuthDB  db.Database
}

// New creates a new Libra instance with the provided configuration.
// If conf is nil, the default configuration will be used.
func New(conf *config.Config) (*Libra, error) {
	if conf == nil {
		c := config.LoadDefaultConfig()
		conf = &c
	}

	libra := &Libra{
		Config: *conf,
	}

	for i, dbConf := range conf.Databases {
		database, err := db.New(dbConf.Engine, dbConf.Options, conf.DataDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create database %d (%s): %w", i+1, dbConf.Engine, err)
		}

		for _, purpose := range dbConf.Purposes {
			switch purpose {
			case "media":
				if libra.MediaDB != nil {
					return nil, fmt.Errorf(
						"duplicate database purpose %q: already assigned to %s",
						purpose,
						libra.MediaDB.EngineName(),
					)
				}
				libra.MediaDB = database
			case "auth":
				if libra.AuthDB != nil {
					return nil, fmt.Errorf(
						"duplicate database purpose %q: already assigned to %s",
						purpose,
						libra.AuthDB.EngineName(),
					)
				}
				libra.AuthDB = database
			}
		}
	}

	for purpose, database := range map[string]db.Database{
		"media": libra.MediaDB,
		"auth":  libra.AuthDB,
	} {
		if database == nil {
			return nil, fmt.Errorf("database for purpose %q is not configured", purpose)
		}
	}

	return libra, nil
}

// Connect establishes connections to all configured databases.
func (l *Libra) Connect() error {
	connected := map[db.Database]bool{}
	for purpose, database := range map[string]db.Database{
		"media": l.MediaDB,
		"auth":  l.AuthDB,
	} {
		if database == nil {
			return fmt.Errorf("database for purpose %q is not configured", purpose)
		}

		if connected[database] {
			continue
		}

		if err := database.Connect(); err != nil {
			// Close any databases that were already connected
			for connectedDB := range connected {
				_ = connectedDB.Close()
			}
			return fmt.Errorf(
				"failed to connect to %s database for purpose %q: %w",
				database.EngineName(),
				purpose,
				err,
			)
		}
		connected[database] = true
	}

	return nil
}

// Close closes all database connections.
func (l *Libra) Close() error {
	var errs []error

	closed := map[db.Database]bool{}
	for _, database := range []db.Database{l.MediaDB, l.AuthDB} {
		if database == nil || closed[database] {
			continue
		}

		if err := database.Close(); err != nil {
			errs = append(errs, err)
		}
		closed[database] = true
	}

	return errors.Join(errs...)
}
