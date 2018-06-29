package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorHeight(t *testing.T) {
	e1 := ErrHeightMismatch(2, 3)
	e1.Error()
	assert.True(t, IsHeightMismatchErr(e1))

	e2 := errors.New("foobar")
	assert.False(t, IsHeightMismatchErr(e2))
	assert.False(t, IsHeightMismatchErr(nil))
}
