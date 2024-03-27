package codec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeKey(t *testing.T) {
	key := NewTimeKey[time.Time]()
	buffer := make([]byte, key.Size(time.Time{}))
	//T := time.UnixMilli(time.Now().UTC().UnixMilli()).UTC()
	T := time.UnixMilli(time.Now().UnixMilli()).UTC()
	t.Log(T)
	bz, err := key.Encode(buffer, T)
	require.NoError(t, err)
	require.Equal(t, len(buffer), bz, "the length of the buffer and the written bytes do not match")
	read, t2, err := key.Decode(buffer)
	require.NoError(t, err)
	require.Equal(t, len(buffer), read, "encoded key and read bytes must have same size")
	require.Equal(t, T, t2, "encoding and decoding produces different keys")
}
