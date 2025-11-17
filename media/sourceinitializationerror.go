package media

import "fmt"

type SourceInitializationError struct {
	SourceID string
	Err      error
}

func (e SourceInitializationError) Error() string {
	return fmt.Sprintf("initialization of source %q failed: %s", e.SourceID, e.Err.Error())
}

func (e SourceInitializationError) Unwrap() error {
	return e.Err
}
