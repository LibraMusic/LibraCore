package types

import "strings"

type LinkedSource string

func (l LinkedSource) GetID() string {
	split := strings.Split(string(l), "::")
	if len(split) == 1 {
		return ""
	}
	return split[0]
}

func (l LinkedSource) GetURL() string {
	split := strings.Split(string(l), "::")
	if len(split) == 1 {
		return split[0]
	}
	return split[1]
}
