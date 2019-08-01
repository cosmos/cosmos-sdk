package state

import (
	"fmt"
)

type GetSafeErrorType byte

const (
	ErrTypeEmptyValue GetSafeErrorType = iota
	ErrTypeUnmarshal
)

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

type GetSafeError struct {
	ty    GetSafeErrorType
	inner error
}

var _ error = (*GetSafeError)(nil) // TODO: sdk.Error

func (err *GetSafeError) Error() string {
	if err.inner == nil {
		return err.ty.Format("")
	}
	return err.ty.Format(err.inner.Error())
}

func ErrEmptyValue() *GetSafeError {
	return &GetSafeError{
		ty: ErrTypeEmptyValue,
	}
}

func ErrUnmarshal(err error) *GetSafeError {
	return &GetSafeError{
		ty:    ErrTypeUnmarshal,
		inner: err,
	}
}
