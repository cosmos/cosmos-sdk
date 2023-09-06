package testnet_test

import (
	"fmt"
	"os"
	"path/filepath"

	cmtcfg "github.com/cometbft/cometbft/config"
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testnet"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Example_basicUsage() {
	const nVals = 2

	// Set up new private keys for the set of validators.
	valPKs := testnet.NewValidatorPrivKeys(nVals)

	// Comet-style validators.
	cmtVals := valPKs.CometGenesisValidators()

	// Cosmos SDK staking validators for genesis.
	stakingVals := cmtVals.StakingValidators()

	const chainID = "example-basic"

	// Create a genesis builder that only requires validators,
	// without any separate delegator accounts.
	//
	// If you need further customization, start with testnet.NewGenesisBuilder().
	b := testnet.DefaultGenesisBuilderOnlyValidators(
		chainID,
		stakingVals,
		// The amount to use in each validator's account during gentx.
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.DefaultPowerReduction),
	)

	// JSON-formatted genesis.
	jGenesis := b.Encode()

	// In this example, we have an outer root directory for the validators.
	// Use t.TempDir() in tests.
	rootDir, err := os.MkdirTemp("", "testnet-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(rootDir)

	// In tests, you probably want to use log.NewTestLoggerInfo(t).
	logger := log.NewNopLogger()

	// The NewNetwork function creates a network of validators.
	// We have to provide a callback to return CometStarter instances.
	// NewNetwork will start all the comet instances concurrently
	// and join the nodes together.
	nodes, err := testnet.NewNetwork(nVals, func(idx int) *testnet.CometStarter {
		// Make a new directory for the validator being created.
		// In tests, this would be a simpler call to t.TempDir().
		dir := filepath.Join(rootDir, fmt.Sprintf("val-%d", idx))
		if err := os.Mkdir(dir, 0o755); err != nil {
			panic(err)
		}

		// TODO: use a different minimal app for this.
		app := simapp.NewSimApp(
			logger.With("instance", idx),
			dbm.NewMemDB(),
			nil,
			true,
			simtestutil.NewAppOptionsWithFlagHome(rootDir),
			baseapp.SetChainID(chainID),
		)

		// Each CometStarter instance must be associated with
		// a distinct comet Config object,
		// as the CometStarter will automatically modify some fields,
		// including P2P.ListenAddress.
		cfg := cmtcfg.DefaultConfig()

		// No need to persist comet's DB to disk in this example.
		cfg.BaseConfig.DBBackend = "memdb"

		return testnet.NewCometStarter(
			app,
			cfg,
			valPKs[idx].Val, // Validator private key for this comet instance.
			jGenesis,        // Raw bytes of genesis file.
			dir,             // Where to put files on disk.
		).Logger(logger.With("root_module", fmt.Sprintf("comet_%d", idx)))
	})
	if err != nil {
		panic(err)
	}

	// StopAndWait must be deferred before the error check,"os"
	// as the nodes value may contain some successfully started instances.
	defer func() {
		err := nodes.StopAndWait()
		if err != nil {
			panic(err)
		}
	}()
	// Now you can begin interacting with the nodes.
	// For the sake of this example, we'll just check
	// a couple simple properties of one node.
	fmt.Println(nodes[0].IsListening())
	fmt.Println(nodes[0].GenesisDoc().ChainID)

	// Output:
	// true
	// example-basic
}
