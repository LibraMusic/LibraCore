package utils

import (
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/LibraMusic/LibraCore/types"
)

var (
	rawVersion      = "0.1.0-dev"
	LibraVersion, _ = types.ParseVersion(rawVersion)
)

func GetVersionInfo() string {
	application := "LibraCore"
	version := "v" + LibraVersion.String()

	info := getBuildInfo()
	if info == nil {
		return version
	}
	if info.Revision != "" {
		version += "+" + info.Revision
	}

	osArch := info.GoOS + "/" + info.GoArch

	date := info.RevisionTime
	if date == "" {
		date = "unknown"
	}

	versionInfo := fmt.Sprintf("%s %s %s BuildDate=%s", application, version, osArch, date)

	return versionInfo
}

type buildInfo struct {
	VersionControlSystem string
	Revision             string
	RevisionTime         string
	Modified             bool

	GoOS   string
	GoArch string

	*debug.BuildInfo
}

var (
	buildInfoInstance *buildInfo
	buildInfoOnce     sync.Once
)

func getBuildInfo() *buildInfo {
	buildInfoOnce.Do(func() {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return
		}

		buildInfoInstance = &buildInfo{
			BuildInfo: info,
		}

		for _, s := range info.Settings {
			switch s.Key {
			case "vcs":
				buildInfoInstance.VersionControlSystem = s.Value
			case "vcs.revision":
				buildInfoInstance.Revision = s.Value
			case "vcs.time":
				buildInfoInstance.RevisionTime = s.Value
			case "vcs.modified":
				buildInfoInstance.Modified = s.Value == "true"
			case "GOOS":
				buildInfoInstance.GoOS = s.Value
			case "GOARCH":
				buildInfoInstance.GoArch = s.Value
			}
		}
	})

	return buildInfoInstance
}
