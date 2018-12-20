package mock

import (
	"testing"

	"github.com/tendermint/tendermint/types"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
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
	appState, err := AppGenState(nil, types.GenesisDoc{}, nil)
	require.NoError(t, err)

	//TODO test validators in the init chain?
	req := abci.RequestInitChain{
		AppStateBytes: appState,
	}
	app.InitChain(req)
	app.Commit()

	// make sure we can query these values
	query := abci.RequestQuery{
		Path: "/store/main/key",
		Data: []byte("foo"),
	}
	qres := app.Query(query)
	require.Equal(t, uint32(0), qres.Code, qres.Log)
	require.Equal(t, []byte("bar"), qres.Value)
}

// TextDeliverTx ensures we can write a tx
func TestDeliverTx(t *testing.T) {
	// set up an app
	app, closer, err := SetupApp()
	// closer may need to be run, even when error in later stage
	if closer != nil {
		defer closer()
	}
	require.NoError(t, err)

	key := "my-special-key"
	value := "top-secret-data!!"
	tx := NewTx(key, value)
	txBytes := tx.GetSignBytes()

	header := abci.Header{
		AppHash: []byte("apphash"),
		Height:  1,
	}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})
	dres := app.DeliverTx(txBytes)
	require.Equal(t, uint32(0), dres.Code, dres.Log)
	app.EndBlock(abci.RequestEndBlock{})
	cres := app.Commit()
	require.NotEmpty(t, cres.Data)

	// make sure we can query these values
	query := abci.RequestQuery{
		Path: "/store/main/key",
		Data: []byte(key),
	}
	qres := app.Query(query)
	require.Equal(t, uint32(0), qres.Code, qres.Log)
	require.Equal(t, []byte(value), qres.Value)
}
