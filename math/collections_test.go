package math

import (
	"cosmossdk.io/collections/colltest"
	"testing"
)

func TestIntValue(t *testing.T) {
	colltest.TestValueCodec(t, IntValue, NewInt(10005994859))
}
