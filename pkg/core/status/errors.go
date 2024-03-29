package status

import (
	"github.com/oneconcern/datamon/pkg/errors"
)

var (
	// ErrInterrupted signals that the current background processing has been interrupted
	ErrInterrupted = errors.New("background processing interrupted")
	// ErrNotFound indicates an object was not found
	ErrNotFound = errors.New("not found")
)
