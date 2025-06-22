package types

import "fmt"

type SourceInitializationError struct {
	SourceID string
	Err      error
}

func (e SourceInitializationError) Error() string {
	return fmt.Sprintf("initialization of source '%s' failed: %s", e.SourceID, e.Err.Error())
}

func (e SourceInitializationError) Unwrap() error {
	return e.Err
}
