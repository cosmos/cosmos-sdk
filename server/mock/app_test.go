package mock

import (
	"math/rand"
	"testing"
	"time"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// SetupApp initializes a new application,
// failing t if initialization fails.
func SetupApp(t *testing.T) abci.Application {
	t.Helper()

	logger := log.NewTestLogger(t)

	rootDir := t.TempDir()

	app, err := NewApp(rootDir, logger)
	require.NoError(t, err)

	return app
}

func TestInitApp(t *testing.T) {
	app := SetupApp(t)

	appState, err := AppGenState(nil, genutiltypes.AppGenesis{}, nil)
	require.NoError(t, err)

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

func TestDeliverTx(t *testing.T) {
	app := SetupApp(t)

	key := "my-special-key"
	value := "top-secret-data!!"

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomAccounts := simtypes.RandomAccounts(r, 1)

	tx := NewTx(key, value, randomAccounts[0].Address)
	txBytes := tx.GetSignBytes()

	app.BeginBlock(abci.RequestBeginBlock{Header: cmtproto.Header{
		AppHash: []byte("apphash"),
		Height:  1,
	}})

	dres := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
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
