package testnet_test

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/simapp"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/rpc/client/http"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testnet"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

// A single comet server in a network runs an RPC server successfully.
func TestCometRPC_SingleRPCServer(t *testing.T) {
	const nVals = 2

	valPKs := testnet.NewValidatorPrivKeys(nVals)
	cmtVals := valPKs.CometGenesisValidators()
	stakingVals := cmtVals.StakingValidators()

	const chainID = "comet-rpc-singleton"

	b := testnet.DefaultGenesisBuilderOnlyValidators(
		chainID,
		stakingVals,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.DefaultPowerReduction),
	)

	jGenesis := b.Encode()

	// Logs shouldn't be necessary here because we are exercising CometStarter,
	// and only doing a very basic check that the RPC talks to the app.
	logger := log.NewNopLogger()

	nodes, err := testnet.NewNetwork(nVals, func(idx int) *testnet.CometStarter {
		rootDir := t.TempDir()

		app := simapp.NewSimApp(
			logger,
			dbm.NewMemDB(),
			nil,
			true,
			simtestutil.NewAppOptionsWithFlagHome(rootDir),
			baseapp.SetChainID(chainID),
		)

		cfg := cmtcfg.DefaultConfig()
		cfg.BaseConfig.DBBackend = "memdb"

		cs := testnet.NewCometStarter(
			app,
			cfg,
			valPKs[idx].Val,
			jGenesis,
			rootDir,
		)

		// Only enable the RPC on the first service.
		if idx == 0 {
			cs = cs.RPCListen()
		}

		return cs
	})
	defer nodes.StopAndWait()
	require.NoError(t, err)

	// Once HTTP client to be shared across the following subtests.
	c, err := http.New(nodes[0].Config().RPC.ListenAddress, "/websocket")
	require.NoError(t, err)

	t.Run("status query", func(t *testing.T) {
		ctx := context.Background()
		st, err := c.Status(ctx)
		require.NoError(t, err)

		// Simple assertion to ensure we have a functioning RPC.
		require.Equal(t, chainID, st.NodeInfo.Network)
	})

	// Block until reported height is at least 1,
	// otherwise we can't make transactions.
	require.NoError(t, testnet.WaitForNodeHeight(nodes[0], 1, 10*time.Second))

	t.Run("simple abci query", func(t *testing.T) {
		res, err := c.ABCIQuery(
			context.Background(),
			"/cosmos.bank.v1beta1.Query/TotalSupply",
			nil,
		)
		require.NoError(t, err)

		registry := codectypes.NewInterfaceRegistry()
		cdc := codec.NewProtoCodec(registry)

		var tsResp banktypes.QueryTotalSupplyResponse
		require.NoError(t, cdc.Unmarshal(res.Response.Value, &tsResp))

		// Just check that something is reported in the supply.
		require.NotEmpty(t, tsResp.Supply)
	})
}

// Starting two comet instances with an RPC server,
// fails with a predictable error.
func TestCometRPC_MultipleRPCError(t *testing.T) {
	const nVals = 2

	valPKs := testnet.NewValidatorPrivKeys(nVals)
	cmtVals := valPKs.CometGenesisValidators()
	stakingVals := cmtVals.StakingValidators()

	const chainID = "comet-rpc-multiple"

	b := testnet.DefaultGenesisBuilderOnlyValidators(
		chainID,
		stakingVals,
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.DefaultPowerReduction),
	)

	jGenesis := b.Encode()

	// Logs shouldn't be necessary here because we are exercising CometStarter.
	logger := log.NewNopLogger()

	nodes, err := testnet.NewNetwork(nVals, func(idx int) *testnet.CometStarter {
		rootDir := t.TempDir()

		app := simapp.NewSimApp(
			logger,
			dbm.NewMemDB(),
			nil,
			true,
			simtestutil.NewAppOptionsWithFlagHome(rootDir),
			baseapp.SetChainID(chainID),
		)

		cfg := cmtcfg.DefaultConfig()
		cfg.BaseConfig.DBBackend = "memdb"

		return testnet.NewCometStarter(
			app,
			cfg,
			valPKs[idx].Val,
			jGenesis,
			rootDir,
		).RPCListen() // Every node has RPCListen enabled, which will cause a failure.
	})
	defer nodes.StopAndWait()

	// Returned error is convertible to CometRPCInUseError.
	// We can't test the exact value because it includes a stack trace.
	require.Error(t, err)
	require.ErrorAs(t, err, new(testnet.CometRPCInUseError))
}
