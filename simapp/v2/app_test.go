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
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"
	serverv2store "cosmossdk.io/server/v2/store"
	"cosmossdk.io/store/v2/db"
	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func NewTestApp(t *testing.T) (*SimApp[transaction.Tx], context.Context) {
	t.Helper()

	logger := log.NewTestLogger(t)

	vp := viper.New()
	vp.Set(serverv2store.FlagAppDBBackend, string(db.DBTypeGoLevelDB))
	vp.Set(serverv2.FlagHome, t.TempDir())

	app, err := NewSimApp[transaction.Tx](depinject.Configs(
		depinject.Supply(logger, runtime.GlobalConfig(vp.AllSettings()))),
	)
	require.NoError(t, err)

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
	accAddr, err := app.txConfig.SigningContext().AddressCodec().BytesToString(acc.GetAddress())
	require.NoError(t, err)
	balance := banktypes.Balance{
		Address: accAddr,
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

	st := app.Store()
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
			Height:    1,
		},
		genesisBytes,
		nil,
	)
	require.NoError(t, err)

	changes, err := newState.GetStateChanges()
	require.NoError(t, err)

	_, err = st.Commit(&store.Changeset{Version: 1, Changes: changes})
	require.NoError(t, err)

	return app, ctx
}

func MoveNextBlock(t *testing.T, app *SimApp[transaction.Tx], ctx context.Context) {
	t.Helper()

	bz := sha256.Sum256([]byte{})

	st := app.Store()
	ci, err := st.LastCommitID()
	require.NoError(t, err)

	height, err := app.LoadLatestHeight()
	height++
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
			Height:  height,
			Time:    time.Now(),
			Hash:    bz[:],
			AppHash: ci.Hash,
		})
	require.NoError(t, err)

	changes, err := newState.GetStateChanges()
	require.NoError(t, err)

	_, err = st.Commit(&store.Changeset{Version: height, Changes: changes})
	require.NoError(t, err)
}

func TestSimAppExportAndBlockedAddrs_WithOneBlockProduced(t *testing.T) {
	app, ctx := NewTestApp(t)

	MoveNextBlock(t, app, ctx)

	_, err := app.ExportAppStateAndValidators(false, nil)
	require.NoError(t, err)
}

func TestSimAppExportAndBlockedAddrs_NoBlocksProduced(t *testing.T) {
	app, _ := NewTestApp(t)

	_, err := app.ExportAppStateAndValidators(false, nil)
	require.NoError(t, err)
}
