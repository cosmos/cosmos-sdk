package rootmulti_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func TestRollback(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	db := dbm.NewMemDB()
	app := simapp.NewSimApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, simapp.DefaultNodeHome, 0, encCfg, simapp.EmptyAppOptions{})
	genesisState := simapp.NewDefaultGenesisState(encCfg.Marshaler)
	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	app.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	app.Commit()

	ver0 := app.LastBlockHeight()
	// commit 10 blocks
	for i := int64(1); i <= 10; i++ {
		header := tmproto.Header{
			Height:  ver0 + i,
			AppHash: app.LastCommitID().Hash,
		}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})
		ctx := app.NewContext(false, header)
		store := ctx.KVStore(app.GetKey("bank"))
		store.Set([]byte("key"), []byte(fmt.Sprintf("value%d", i)))
		app.Commit()
	}

	require.Equal(t, ver0+10, app.LastBlockHeight())
	store := app.NewContext(true, tmproto.Header{}).KVStore(app.GetKey("bank"))
	require.Equal(t, []byte("value10"), store.Get([]byte("key")))

	// rollback 5 blocks
	target := ver0 + 5
	require.NoError(t, app.CommitMultiStore().RollbackToVersion(target))
	require.Equal(t, target, app.LastBlockHeight())

	// recreate app to have clean check state
	app = simapp.NewSimApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, simapp.DefaultNodeHome, 0, encCfg, simapp.EmptyAppOptions{})
	store = app.NewContext(true, tmproto.Header{}).KVStore(app.GetKey("bank"))
	require.Equal(t, []byte("value5"), store.Get([]byte("key")))

	// commit another 5 blocks with different values
	for i := int64(6); i <= 10; i++ {
		header := tmproto.Header{
			Height:  ver0 + i,
			AppHash: app.LastCommitID().Hash,
		}
		app.BeginBlock(abci.RequestBeginBlock{Header: header})
		ctx := app.NewContext(false, header)
		store := ctx.KVStore(app.GetKey("bank"))
		store.Set([]byte("key"), []byte(fmt.Sprintf("VALUE%d", i)))
		app.Commit()
	}

	require.Equal(t, ver0+10, app.LastBlockHeight())
	store = app.NewContext(true, tmproto.Header{}).KVStore(app.GetKey("bank"))
	require.Equal(t, []byte("VALUE10"), store.Get([]byte("key")))
}
