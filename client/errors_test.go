package client

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorNoData(t *testing.T) {
	e1 := ErrNoData()
	e1.Error()
	assert.True(t, IsNoDataErr(e1))

	e2 := errors.New("foobar")
	assert.False(t, IsNoDataErr(e2))
	assert.False(t, IsNoDataErr(nil))
}
