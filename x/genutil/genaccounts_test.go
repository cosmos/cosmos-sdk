package genutil_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func TestAddGenesisAccountsDuplicateAddress(t *testing.T) {
	_, _, addr := testdata.KeyTestPubAddr()
	addrStr := addr.String()
	accounts := []genutil.GenesisAccount{
		{Address: addrStr, Coins: sdk.NewCoins(sdk.NewInt64Coin("test", 1))},
		{Address: addrStr, Coins: sdk.NewCoins(sdk.NewInt64Coin("test", 2))},
	}

	for _, appendAcct := range []bool{false, true} {
		genFile := filepath.Join(t.TempDir(), "genesis.json")
		genesis := genutiltypes.NewAppGenesisWithVersion("test", json.RawMessage(`{}`))
		require.NoError(t, genutil.ExportGenesisFile(genesis, genFile))

		cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}).Codec
		err := genutil.AddGenesisAccounts(cdc, addresscodec.NewBech32Codec("cosmos"), accounts, appendAcct, genFile)
		if !appendAcct {
			require.ErrorContains(t, err, addrStr+" already exists")
			continue
		}
		require.NoError(t, err)

		appState, _, err := genutiltypes.GenesisStateFromGenFile(genFile)
		require.NoError(t, err)

		authState := authtypes.GetGenesisStateFromAppState(cdc, appState)
		require.Len(t, authState.Accounts, 1)

		bankState := banktypes.GetGenesisStateFromAppState(cdc, appState)
		require.Len(t, bankState.Balances, 1)
		require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("test", 3)), bankState.Balances[0].Coins)
		require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("test", 3)), bankState.Supply)
	}
}
