package types

type SourceError struct {
	SourceID string
	Err      error
}

func (e SourceError) Error() string {
	return e.Err.Error() + " (source: " + e.SourceID + ")"
}
