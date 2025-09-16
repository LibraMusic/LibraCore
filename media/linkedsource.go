package media

import "strings"

func GetLinkedSourceID(linkedSource string) string {
	split := strings.Split(linkedSource, "::")
	if len(split) == 1 {
		return ""
	}
	return split[0]
}

func GetLinkedSourceURL(linkedSource string) string {
	split := strings.Split(linkedSource, "::")
	if len(split) == 1 {
		return split[0]
	}
	return split[1]
}
