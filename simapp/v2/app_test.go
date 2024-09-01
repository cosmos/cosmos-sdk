package simapp

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"testing"
	"time"

	"github.com/cometbft/cometbft/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/comet"
	context2 "cosmossdk.io/core/context"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	serverv2 "cosmossdk.io/server/v2"
	comettypes "cosmossdk.io/server/v2/cometbft/types"
	"cosmossdk.io/store/v2/db"
	authtypes "cosmossdk.io/x/auth/types"
	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewTestApp(t *testing.T) (*SimApp[transaction.Tx], context.Context) {
	t.Helper()

	logger := log.NewTestLogger(t)

	vp := viper.New()
	vp.Set("store.app-db-backend", string(db.DBTypeGoLevelDB))
	vp.Set(serverv2.FlagHome, t.TempDir())

	app := NewSimApp[transaction.Tx](logger, vp)
	genesis := app.ModuleManager().DefaultGenesis()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := types.NewValidator(pubKey, 1)
	valSet := types.NewValidatorSet([]*types.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000000000))),
	}

	genesis, err = simtestutil.GenesisStateWithValSet(
		app.AppCodec(),
		genesis,
		valSet,
		[]authtypes.GenesisAccount{acc},
		balance,
	)
	require.NoError(t, err)

	genesisBytes, err := json.Marshal(genesis)
	require.NoError(t, err)

	st := app.GetStore().(comettypes.Store)
	ci, err := st.LastCommitID()
	require.NoError(t, err)

	bz := sha256.Sum256([]byte{})

	ctx := context.Background()

	_, newState, err := app.InitGenesis(
		ctx,
		&server.BlockRequest[transaction.Tx]{
			Time:      time.Now(),
			Hash:      bz[:],
			ChainId:   "theChain",
			AppHash:   ci.Hash,
			IsGenesis: true,
		},
		genesisBytes,
		nil,
	)
	require.NoError(t, err)

	changes, err := newState.GetStateChanges()
	require.NoError(t, err)

	_, err = st.Commit(&store.Changeset{Changes: changes})
	require.NoError(t, err)

	return app, ctx
}

func MoveNextBlock(t *testing.T, app *SimApp[transaction.Tx], ctx context.Context) {
	t.Helper()

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
		&server.BlockRequest[transaction.Tx]{
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

	_, err := app.ExportAppStateAndValidators(nil)
	require.NoError(t, err)
}

func TestSimAppExportAndBlockedAddrs_NoBlocksProduced(t *testing.T) {
	app, _ := NewTestApp(t)

	_, err := app.ExportAppStateAndValidators(nil)
	require.NoError(t, err)
}
