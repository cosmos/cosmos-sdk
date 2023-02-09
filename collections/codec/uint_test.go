package codec

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUintJSON(t *testing.T) {
	var x uint64 = 3076
	bz, err := uintEncodeJSON(x)
	require.NoError(t, err)
	require.Equal(t, []byte(`"3076"`), bz)
	y, err := uintDecodeJSON(bz, 64)
	require.NoError(t, err)
	require.Equal(t, x, y)
}
