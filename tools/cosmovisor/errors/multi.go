package errors

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

// MultiError is an error combining multiple other errors.
// It will never have 0 or 1 errors. It will always have two or more.
type MultiError struct {
	errs []error
}

// FlattenErrors possibly creates a MultiError.
// Nil entries are ignored.
// If all provided errors are nil (or nothing is provided), nil is returned.
// If only one non-nil error is provided, it is returned unchanged.
// If two or more non-nil errors are provided, the returned error will be of type *MultiError
// and it will contain each non-nil error.
func FlattenErrors(errs ...error) error {
	rv := MultiError{}
	for _, err := range errs {
		if err != nil {
			if merr, isMerr := err.(*MultiError); isMerr {
				rv.errs = append(rv.errs, merr.errs...)
			} else {
				rv.errs = append(rv.errs, err)
			}
		}
	}
	switch rv.Len() {
	case 0:
		return nil
	case 1:
		return rv.errs[0]
	}
	return &rv
}

// GetErrors gets all the errors that make up this MultiError.
func (e MultiError) GetErrors() []error {
	// Return a copy of the errs slice to prevent alteration of the original slice.
	rv := make([]error, e.Len())
	copy(rv, e.errs)
	return rv
}

// Len gets the number of errors in this MultiError.
func (e MultiError) Len() int {
	return len(e.errs)
}

// Error implements the error interface for a MultiError.
func (e *MultiError) Error() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d errors: ", len(e.errs))
	for i, err := range e.errs {
		if i != 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%d: %v", i+1, err)
	}
	return sb.String()
}

// String implements the string interface for a MultiError.
func (e MultiError) String() string {
	return e.Error()
}

func LogErrors(logger *zerolog.Logger, msg string, err error) {
	switch err := err.(type) {
	case *MultiError:
		if msg != "" {
			logger.Error().Msg(msg)
		}
		for i, e := range err.GetErrors() {
			logger.Error().Err(e).Msg(fmt.Sprintf("  %d:", i+1))
		}
	default:
		logger.Error().Err(err).Msg(msg)
	}
}
