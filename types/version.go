package types

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major  int
	Minor  int
	Patch  int
	Suffix string
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Patch, v.Suffix)
}

func (v Version) IsAtLeast(other Version) bool {
	if v.Major > other.Major {
		return true
	}
	if v.Major < other.Major {
		return false
	}
	if v.Minor > other.Minor {
		return true
	}
	if v.Minor < other.Minor {
		return false
	}
	if v.Patch > other.Patch {
		return true
	}
	if v.Patch < other.Patch {
		return false
	}
	return strings.Compare(v.Suffix, other.Suffix) >= 0
}

func ParseVersion(version string) (Version, error) {
	var major, minor, patch int
	var suffix string
	var err error

	if strings.Contains(version, "-") {
		parts := strings.SplitN(version, "-", 2)
		suffix = "-" + parts[1]
		version = parts[0]
	}

	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		major, err = strconv.Atoi(parts[0])
		if err != nil {
			return Version{}, err
		}
	}
	if len(parts) > 1 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return Version{}, err
		}
	}
	if len(parts) > 2 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return Version{}, err
		}
	}

	return Version{
		Major:  major,
		Minor:  minor,
		Patch:  patch,
		Suffix: suffix,
	}, nil
}
