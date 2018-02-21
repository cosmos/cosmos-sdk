package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
)

// TestInitApp makes sure we can initialize this thing without an error
func TestInitApp(t *testing.T) {
	// set up an app
	app, closer, err := SetupApp()
	// closer may need to be run, even when error in later stage
	if closer != nil {
		defer closer()
	}
	require.NoError(t, err)

	// initialize it future-way
	opts, err := GenInitOptions(nil)
	require.NoError(t, err)
	req := abci.RequestInitChain{AppStateBytes: opts}
	app.InitChain(req)

	// make sure we can query these values
	query := abci.RequestQuery{
		Path: "/main/key",
		Data: []byte("foo"),
	}
	qres := app.Query(query)
	require.Equal(t, uint32(0), qres.Code, qres.Log)
	assert.Equal(t, []byte("bar"), qres.Value)
}
