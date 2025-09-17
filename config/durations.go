package config

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ParseHumanDuration(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return ParseDuration("0s")
	}
	if strings.Contains(durationStr, ":") {
		split := strings.Split(durationStr, ":")
		if len(split) == 2 {
			return ParseDuration(split[0] + "m" + split[1] + "s")
		}
		if len(split) == 3 {
			return ParseDuration(split[0] + "h" + split[1] + "m" + split[2] + "s")
		}
	}
	if _, err := strconv.Atoi(durationStr); err != nil {
		durationStr = strings.ReplaceAll(durationStr, ",", "")
		durationStr = strings.ReplaceAll(durationStr, " ", "")
		durationStr = strings.ReplaceAll(durationStr, "weeks", "w")
		durationStr = strings.ReplaceAll(durationStr, "week", "w")
		durationStr = strings.ReplaceAll(durationStr, "days", "d")
		durationStr = strings.ReplaceAll(durationStr, "day", "d")
		durationStr = strings.ReplaceAll(durationStr, "hours", "h")
		durationStr = strings.ReplaceAll(durationStr, "hour", "h")
		durationStr = strings.ReplaceAll(durationStr, "minutes", "m")
		durationStr = strings.ReplaceAll(durationStr, "minute", "m")
		durationStr = strings.ReplaceAll(durationStr, "seconds", "s")
		durationStr = strings.ReplaceAll(durationStr, "second", "s")
		return ParseDuration(durationStr)
	}
	return ParseDuration(durationStr + "s")
}

// ParseDuration parses a duration string.
// Examples: "10d", "-1.5w" or "3Y4M5d".
// Add time units are "d"="D", "w"="W".
func ParseDuration(s string) (time.Duration, error) {
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}

	re := regexp.MustCompile(`(\d*\.\d+|\d+)[^\d]*`)
	unitMap := map[string]int{
		"d": 24,
		"D": 24,
		"w": 7 * 24,
		"W": 7 * 24,
	}

	strs := re.FindAllString(s, -1)
	var sumDur time.Duration
	for _, str := range strs {
		hours := 1
		for unit, h := range unitMap {
			if strings.Contains(str, unit) {
				str = strings.ReplaceAll(str, unit, "h")
				hours = h
				break
			}
		}

		dur, err := time.ParseDuration(str)
		if err != nil {
			return 0, err
		}

		sumDur += dur * time.Duration(hours)
	}

	if neg {
		sumDur = -sumDur
	}
	return sumDur, nil
}
