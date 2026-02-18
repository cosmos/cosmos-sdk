package codec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeValue_BinaryRoundTrip(t *testing.T) {
	cdc := TimeValue()

	tt := time.Date(2026, 1, 25, 12, 34, 56, 123456789, time.UTC)

	b, err := cdc.Encode(tt)
	require.NoError(t, err)
	require.Len(t, b, 8)

	got, err := cdc.Decode(b)
	require.NoError(t, err)

	require.True(t, tt.Equal(got))
	require.Equal(t, time.UTC, got.Location())
}

func TestTimeValue_BinaryInvalidLen(t *testing.T) {
	cdc := TimeValue()

	_, err := cdc.Decode([]byte{1, 2, 3})
	require.Error(t, err)
}

func TestTimeValue_JSONRoundTrip(t *testing.T) {
	cdc := TimeValue()

	tt := time.Date(2026, 1, 25, 12, 34, 56, 123456789, time.UTC)

	b, err := cdc.EncodeJSON(tt)
	require.NoError(t, err)

	got, err := cdc.DecodeJSON(b)
	require.NoError(t, err)

	require.True(t, tt.Equal(got))
	require.Equal(t, time.UTC, got.Location())
}

func TestTimeValue_Stringify(t *testing.T) {
	cdc := TimeValue()

	tt := time.Date(2026, 1, 25, 12, 34, 56, 123456789, time.UTC)
	s := cdc.Stringify(tt)

	require.Contains(t, s, "2026-01-25T12:34:56")
	require.Contains(t, s, "Z")
}
