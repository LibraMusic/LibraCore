package util

import "strings"

func IsValidSourceURL(urlStr string) bool {
	return HasSupportedScheme(urlStr)
}

func HasSupportedScheme(urlStr string) bool {
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") || !strings.Contains(urlStr, "://") {
		return true
	}
	return false
}
