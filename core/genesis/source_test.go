package genesis

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/appmodule"
)

func TestSource(t *testing.T) {
	source, err := SourceFromRawJSON(json.RawMessage(testJSON))
	require.NoError(t, err)

	expectJSON(t, source, "foo", fooContents)
	expectJSON(t, source, "bar", barContents)

	// missing fields just return nil, nil
	r, err := source("baz")
	require.NoError(t, err)
	require.Nil(t, r)
}

func expectJSON(t *testing.T, source appmodule.GenesisSource, field, contents string) {
	t.Helper()
	r, err := source(field)
	require.NoError(t, err)
	bz, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, contents, string(bz))
}

const (
	testJSON = `
{
	"foo":{"x":1,"y":"abc"},
	"bar":[1,2,3,4]
}
`
	fooContents = `{"x":1,"y":"abc"}`
	barContents = `[1,2,3,4]`
)
