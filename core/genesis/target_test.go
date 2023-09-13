package genesis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTarget(t *testing.T) {
	target := &RawJSONTarget{}

	w, err := target.Target()("foo")
	require.NoError(t, err)
	_, err = w.Write([]byte("1"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	w, err = target.Target()("bar")
	require.NoError(t, err)
	_, err = w.Write([]byte(`"abc"`))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	bz, err := target.JSON()
	require.NoError(t, err)

	// test that it's correct by reading back with a source
	source, err := SourceFromRawJSON(bz)
	require.NoError(t, err)

	expectJSON(t, source, "foo", "1")
	expectJSON(t, source, "bar", `"abc"`)
}
