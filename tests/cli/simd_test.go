// +build cli_test

package cli_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestCLISimdCollectGentxs(t *testing.T) {
	t.SkipNow() // TODO: Bring back once viper is refactored.
	t.Parallel()
	var customMaxBytes, customMaxGas int64 = 99999999, 1234567
	f := cli.NewFixtures(t)

	// Initialise temporary directories
	gentxDir, err := ioutil.TempDir("", "")
	gentxDoc := filepath.Join(gentxDir, "gentx.json")
	require.NoError(t, err)

	// Reset testing path
	f.UnsafeResetAll()

	// Initialize keys
	f.KeysAdd(cli.KeyFoo)

	// Run init
	f.SDInit(cli.KeyFoo)

	// Customise genesis.json

	genFile := f.GenesisFile()
	genDoc, err := tmtypes.GenesisDocFromFile(genFile)
	require.NoError(t, err)
	genDoc.ConsensusParams.Block.MaxBytes = customMaxBytes
	genDoc.ConsensusParams.Block.MaxGas = customMaxGas
	genDoc.SaveAs(genFile)

	// Add account to genesis.json
	f.AddGenesisAccount(f.KeyAddress(cli.KeyFoo), cli.StartCoins)

	// Write gentx file
	f.GenTx(cli.KeyFoo, fmt.Sprintf("--output-document=%s", gentxDoc))

	// Collect gentxs from a custom directory
	f.CollectGenTxs(fmt.Sprintf("--gentx-dir=%s", gentxDir))

	genDoc, err = tmtypes.GenesisDocFromFile(genFile)
	require.NoError(t, err)
	require.Equal(t, genDoc.ConsensusParams.Block.MaxBytes, customMaxBytes)
	require.Equal(t, genDoc.ConsensusParams.Block.MaxGas, customMaxGas)

	f.Cleanup(gentxDir)
}

func TestCLISimdAddGenesisAccount(t *testing.T) {
	t.Parallel()
	f := cli.NewFixtures(t)

	// Reset testing path
	f.UnsafeResetAll()

	// Initialize keys
	f.KeysDelete(cli.KeyFoo)
	f.KeysDelete(cli.KeyBar)
	f.KeysDelete(cli.KeyBaz)
	f.KeysAdd(cli.KeyFoo)
	f.KeysAdd(cli.KeyBar)
	f.KeysAdd(cli.KeyBaz)

	// Run init
	f.SDInit(cli.KeyFoo)

	// Add account to genesis.json
	bazCoins := sdk.Coins{
		sdk.NewInt64Coin("acoin", 1000000),
		sdk.NewInt64Coin("bcoin", 1000000),
	}

	f.AddGenesisAccount(f.KeyAddress(cli.KeyFoo), cli.StartCoins)
	f.AddGenesisAccount(f.KeyAddress(cli.KeyBar), bazCoins)

	genesisState := f.GenesisState()

	appCodec := f.EncodingConfig.Marshaler

	accounts := authtypes.GetGenesisStateFromAppState(appCodec, genesisState).Accounts
	balances := banktypes.GetGenesisStateFromAppState(f.Cdc, genesisState).Balances
	balancesSet := make(map[string]sdk.Coins)

	for _, b := range balances {
		balancesSet[b.GetAddress().String()] = b.Coins
	}

	require.Equal(t, accounts[0].GetAddress(), f.KeyAddress(cli.KeyFoo))
	require.Equal(t, accounts[1].GetAddress(), f.KeyAddress(cli.KeyBar))
	require.True(t, balancesSet[accounts[0].GetAddress().String()].IsEqual(cli.StartCoins))
	require.True(t, balancesSet[accounts[1].GetAddress().String()].IsEqual(bazCoins))

	// Cleanup testing directories
	f.Cleanup()
}

func TestCLIValidateGenesis(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	f.ValidateGenesis()

	// Cleanup testing directories
	f.Cleanup()
}
