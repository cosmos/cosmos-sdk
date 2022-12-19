package types_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/types"
)

func TestPrefixEndBytes(t *testing.T) {
	t.Parallel()
	bs1 := []byte{0x23, 0xA5, 0x06}
	require.True(t, bytes.Equal([]byte{0x23, 0xA5, 0x07}, types.PrefixEndBytes(bs1)))
	bs2 := []byte{0x23, 0xA5, 0xFF}
	require.True(t, bytes.Equal([]byte{0x23, 0xA6}, types.PrefixEndBytes(bs2)))
	require.Nil(t, types.PrefixEndBytes([]byte{0xFF}))
	require.Nil(t, types.PrefixEndBytes(nil))
}

func TestInclusiveEndBytes(t *testing.T) {
	t.Parallel()
	require.True(t, bytes.Equal([]byte{0x00}, types.InclusiveEndBytes(nil)))
	bs := []byte("test")
	require.True(t, bytes.Equal(append(bs, byte(0x00)), types.InclusiveEndBytes(bs)))
}
