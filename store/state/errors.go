package state

import (
	"fmt"
)

// GetSafeErrorType is enum for indicating the type of error
type GetSafeErrorType byte

const (
	// ErrTypeEmptyValue is used for nil byteslice values
	ErrTypeEmptyValue GetSafeErrorType = iota
	// ErrTypeUnmarshal is used for undeserializable values
	ErrTypeUnmarshal
)

// Implements Formatter
func (ty GetSafeErrorType) Format(msg string) (res string) {
	switch ty {
	case ErrTypeEmptyValue:
		res = fmt.Sprintf("Empty Value found")
	case ErrTypeUnmarshal:
		res = fmt.Sprintf("Error while unmarshal")
	default:
		panic("Unknown error type")
	}

	if msg != "" {
		res = fmt.Sprintf("%s: %s", res, msg)
	}

	return
}

// GetSafeError is error type for GetSafe method
type GetSafeError struct {
	ty    GetSafeErrorType
	inner error
}

var _ error = GetSafeError{}

// Implements error
func (err GetSafeError) Error() string {
	if err.inner == nil {
		return err.ty.Format("")
	}
	return err.ty.Format(err.inner.Error())
}

// ErrEmptyValue constructs GetSafeError with ErrTypeEmptyValue
func ErrEmptyValue() GetSafeError {
	return GetSafeError{
		ty: ErrTypeEmptyValue,
	}
}

// ErrUnmarshal constructs GetSafeError with ErrTypeUnmarshal
func ErrUnmarshal(err error) GetSafeError {
	return GetSafeError{
		ty:    ErrTypeUnmarshal,
		inner: err,
	}
}
