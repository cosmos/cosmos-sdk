package keyerror

import (
	"fmt"
)

const (
	codeKeyNotFound   = 1
	codeWrongPassword = 2
)

type keybaseError interface {
	error
	Code() int
}

type errKeyNotFound struct {
	code int
	name string
}

func (e errKeyNotFound) Code() int {
	return e.code
}

func (e errKeyNotFound) Error() string {
	return fmt.Sprintf("Key %s not found", e.name)
}

// NewErrKeyNotFound returns a standardized error reflecting that the specified key doesn't exist
func NewErrKeyNotFound(name string) error {
	return errKeyNotFound{
		code: codeKeyNotFound,
		name: name,
	}
}

// IsErrKeyNotFound returns true if the given error is errKeyNotFound
func IsErrKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	if keyErr, ok := err.(keybaseError); ok {
		if keyErr.Code() == codeKeyNotFound {
			return true
		}
	}
	return false
}

type errWrongPassword struct {
	code int
}

func (e errWrongPassword) Code() int {
	return e.code
}

func (e errWrongPassword) Error() string {
	return "invalid account password"
}

// NewErrWrongPassword returns a standardized error reflecting that the specified password is wrong
func NewErrWrongPassword() error {
	return errWrongPassword{
		code: codeWrongPassword,
	}
}

// IsErrWrongPassword returns true if the given error is errWrongPassword
func IsErrWrongPassword(err error) bool {
	if err == nil {
		return false
	}
	if keyErr, ok := err.(keybaseError); ok {
		if keyErr.Code() == codeWrongPassword {
			return true
		}
	}
	return false
}
