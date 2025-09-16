package media

type UnsupportedSourceTypeError struct {
	SourceType string
}

func (e UnsupportedSourceTypeError) Error() string {
	return "unsupported source type: " + e.SourceType
}
