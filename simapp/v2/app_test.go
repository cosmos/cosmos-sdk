package simapp

import (
	"context"
	app2 "cosmossdk.io/core/app"
	"cosmossdk.io/core/comet"
	context2 "cosmossdk.io/core/context"
	serverv2 "cosmossdk.io/server/v2"
	"crypto/sha256"
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	comettypes "cosmossdk.io/server/v2/cometbft/types"
	"cosmossdk.io/store/v2/db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func NewTestApp(t *testing.T) (*SimApp[transaction.Tx], context.Context) {
	logger := log.NewTestLogger(t)

	vp := viper.New()
	vp.Set("store.app-db-backend", string(db.DBTypeGoLevelDB))
	vp.Set(serverv2.FlagHome, t.TempDir())

	app := NewSimApp[transaction.Tx](logger, vp)
	genesis := app.ModuleManager().DefaultGenesis()
	genesisBytes, err := json.Marshal(genesis)
	require.NoError(t, err)

	st := app.GetStore().(comettypes.Store)
	ci, err := st.LastCommitID()
	require.NoError(t, err)

	bz := sha256.Sum256([]byte{})

	ctx := context.Background()

	_, newState, err := app.InitGenesis(
		ctx,
		&app2.BlockRequest[transaction.Tx]{
			0,
			time.Now(),
			bz[:],
			"theChain",
			ci.Hash,
			nil,
			nil,
			true,
		},
		genesisBytes,
		nil,
	)
	require.NoError(t, err)

	changes, err := newState.GetStateChanges()
	require.NoError(t, err)

	_, err = st.Commit(&store.Changeset{changes})
	require.NoError(t, err)

	return app, ctx
}

func MoveNextBlock(t *testing.T, app *SimApp[transaction.Tx], ctx context.Context) {
	bz := sha256.Sum256([]byte{})

	st := app.GetStore().(comettypes.Store)
	ci, err := st.LastCommitID()
	require.NoError(t, err)

	height, err := app.LoadLatestHeight()
	require.NoError(t, err)

	// TODO: this is a hack to set the comet info in the context for distribution module dependency.
	ctx = context.WithValue(ctx, context2.CometInfoKey, comet.Info{
		Evidence:        nil,
		ValidatorsHash:  nil,
		ProposerAddress: nil,
		LastCommit:      comet.CommitInfo{},
	})

	_, newState, err := app.DeliverBlock(
		ctx,
		&app2.BlockRequest[transaction.Tx]{
			Height:  height + 1,
			Time:    time.Now(),
			Hash:    bz[:],
			AppHash: ci.Hash,
		})
	require.NoError(t, err)

	changes, err := newState.GetStateChanges()
	require.NoError(t, err)

	_, err = st.Commit(&store.Changeset{Changes: changes})
	require.NoError(t, err)
}

func TestSimAppExportAndBlockedAddrs_WithOneBlockProduced(t *testing.T) {
	app, ctx := NewTestApp(t)

	MoveNextBlock(t, app, ctx)

	_, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
}

func TestSimAppExportAndBlockedAddrs_NoBlocksProduced(t *testing.T) {
	app, _ := NewTestApp(t)

	_, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
}
