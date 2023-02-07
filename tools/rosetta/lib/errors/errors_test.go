package errors

import (
	"testing"

	cmttypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	"github.com/stretchr/testify/assert"
)

func TestRegisterError(t *testing.T) {
	var error *Error
	// this is the number of errors registered by default in errors.go
	registeredErrorsCount := 16
	assert.Equal(t, len(registry.list()), registeredErrorsCount)
	assert.ElementsMatch(t, registry.list(), ListErrors())
	// add a new Error
	error = RegisterError(69, "nice!", false, "nice!")
	assert.NotNil(t, error)
	// now we have a new error
	registeredErrorsCount++
	assert.Equal(t, len(ListErrors()), registeredErrorsCount)
	// re-register an error should not change anything
	error = RegisterError(69, "nice!", false, "nice!")
	assert.Equal(t, len(ListErrors()), registeredErrorsCount)

	// test sealing
	assert.Equal(t, registry.sealed, false)
	errors := SealAndListErrors()
	assert.Equal(t, registry.sealed, true)
	assert.Equal(t, len(errors), registeredErrorsCount)
	// add a new error on a sealed registry
	error = RegisterError(1024, "bytes", false, "bytes")
	assert.NotNil(t, error)
}

func TestError_Error(t *testing.T) {
	var error *Error
	// nil cases
	assert.False(t, ErrOffline.Is(error))
	error = &Error{}
	assert.False(t, ErrOffline.Is(error))
	// wrong type
	assert.False(t, ErrOffline.Is(&MyError{}))
	// test with wrapping an error
	error = WrapError(ErrOffline, "offline")
	assert.True(t, ErrOffline.Is(error))

	// test equality
	assert.False(t, ErrOffline.Is(ErrBadGateway))
	assert.True(t, ErrBadGateway.Is(ErrBadGateway))
}

func TestToRosetta(t *testing.T) {
	var error *Error
	// nil case
	assert.NotNil(t, ToRosetta(error))
	// wrong type
	assert.NotNil(t, ToRosetta(&MyError{}))

	tmErr := &cmttypes.RPCError{}
	// RpcError case
	assert.NotNil(t, ToRosetta(tmErr))
}

type MyError struct{}

func (e *MyError) Error() string {
	return ""
}

func (e *MyError) Is(err error) bool {
	return true
}
