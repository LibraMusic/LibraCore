package types

type InvalidSourceError struct {
	SourceID string
}

func (e InvalidSourceError) Error() string {
	return "invalid source: " + e.SourceID
}
