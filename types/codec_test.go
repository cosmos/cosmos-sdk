package types

import (
	"testing"

	"cosmossdk.io/collections/colltest"
)

func TestIntValue(t *testing.T) {
	colltest.TestValueCodec(t, IntValue, NewInt(10005994859))
}
