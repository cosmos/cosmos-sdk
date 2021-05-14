package errors

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
