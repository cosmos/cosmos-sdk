package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/math"
)

func TestIntValue(t *testing.T) {
	colltest.TestValueCodec(t, IntValue, math.NewInt(10005994859))

	original := math.NewInt(132005994859)

	// Stringify Test
	str := IntValue.Stringify(original)
	require.Equal(t, original.String(), str)

	// ValueType Test
	require.Equal(t, IntValue.ValueType(), "math.Int")
}

func TestUintValue(t *testing.T) {
	colltest.TestValueCodec(t, UintValue, math.NewUint(1337))
	colltest.TestValueCodec(t, UintValue, math.ZeroUint())
	colltest.TestValueCodec(t, UintValue, math.NewUintFromString("1000000000000000000"))

	original := math.NewUint(1234567890)

	// Stringify Test
	str := UintValue.Stringify(original)
	require.Equal(t, original.String(), str)

	// ValueType Test
	require.Equal(t, UintValue.ValueType(), "math.Uint")
}
