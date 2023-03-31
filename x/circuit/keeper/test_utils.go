package keeper

import (
	"testing"
)

type baseFixture struct {
	t   *testing.T
	err error
}

func initFixture(t *testing.T) *baseFixture {
	s := &baseFixture{t: t}

	return s
}
