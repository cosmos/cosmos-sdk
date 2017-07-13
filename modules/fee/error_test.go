package fee

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	assert := assert.New(t)

	e := ErrInsufficientFees()
	assert.True(IsInsufficientFeesErr(e))
	assert.False(IsWrongFeeDenomErr(e))

	e2 := ErrWrongFeeDenom("atom")
	assert.False(IsInsufficientFeesErr(e2))
	assert.True(IsWrongFeeDenomErr(e2))
}
