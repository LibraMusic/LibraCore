package media

import "strings"

func LinkedSourceID(linkedSource string) string {
	split := strings.Split(linkedSource, "::")
	if len(split) == 1 {
		return ""
	}
	return split[0]
}

func LinkedSourceURL(linkedSource string) string {
	split := strings.Split(linkedSource, "::")
	if len(split) == 1 {
		return split[0]
	}
	return split[1]
}
