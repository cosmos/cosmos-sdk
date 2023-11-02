package types

import (
	"testing"

	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/math"
)

func TestIntValue(t *testing.T) {
	colltest.TestValueCodec(t, IntValue, math.NewInt(10005994859))
}
