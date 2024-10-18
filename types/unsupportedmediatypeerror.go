package types

type UnsupportedMediaTypeError struct {
	MediaType string
}

func (e UnsupportedMediaTypeError) Error() string {
	return "unsupported media type: " + e.MediaType
}
