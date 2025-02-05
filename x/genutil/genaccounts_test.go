package genutil_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/bank"
	banktypes "cosmossdk.io/x/bank/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

func TestAddGenesisAccounts(t *testing.T) {
	addressCodec := addresscodec.NewBech32Codec("cosmos")
	cdc := testutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{})

	addresses := []string{
		"cosmos1cxlt8kznps92fwu3j6npahx4mjfutydyene2qw",
		"cosmos1hd6fsrvnz6qkp87s3u86ludegq97agxsdkwzyh",
	}

	appState := map[string]json.RawMessage{
		"auth": cdc.Codec.MustMarshalJSON(&authtypes.GenesisState{}),
		"bank": cdc.Codec.MustMarshalJSON(banktypes.DefaultGenesisState()),
	}

	authState := authtypes.NewGenesisState(authtypes.DefaultParams(), []authtypes.GenesisAccount{
		&authtypes.BaseAccount{Address: addresses[0]},
		&authtypes.BaseAccount{Address: addresses[1]},
	})

	bankState := banktypes.DefaultGenesisState()
	bankState.Balances = []banktypes.Balance{
		{
			Address: addresses[0],
			Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
		},
		{
			Address: addresses[1],
			Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 4)),
		},
	}
	bankState.Supply = sdk.NewCoins(sdk.NewInt64Coin("test", 5))
	appStateWithAccounts := map[string]json.RawMessage{
		"auth": cdc.Codec.MustMarshalJSON(authState),
		"bank": cdc.Codec.MustMarshalJSON(bankState),
	}

	testCases := []struct {
		name           string
		appState       map[string]json.RawMessage
		genesisAccount []genutil.GenesisAccount
		expected       map[string]sdk.Coins
		expectedError  string
		appendFlag     bool
	}{
		{
			name:     "single addition without append",
			appState: appState,
			genesisAccount: []genutil.GenesisAccount{
				{
					Address: addresses[0],
					Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
				},
				{
					Address: addresses[1],
					Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 2)),
				},
			},
			expected: map[string]sdk.Coins{
				addresses[0]: sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
				addresses[1]: sdk.NewCoins(sdk.NewInt64Coin("test", 2)),
			},
		},
		{
			name:     "already existing account additions without append",
			appState: appStateWithAccounts,
			genesisAccount: []genutil.GenesisAccount{
				{
					Address: addresses[0],
					Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
				},
			},
			expectedError: "already exists",
		},
		{
			name:     "already existing account additions with append",
			appState: appStateWithAccounts,
			genesisAccount: []genutil.GenesisAccount{
				{
					Address: addresses[0],
					Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
				},
				{
					Address: addresses[1],
					Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1), sdk.NewInt64Coin("stake", 1)),
				},
			},
			expected: map[string]sdk.Coins{
				addresses[0]: sdk.NewCoins(sdk.NewInt64Coin("test", 2)),
				addresses[1]: sdk.NewCoins(sdk.NewInt64Coin("test", 5), sdk.NewInt64Coin("stake", 1)),
			},
			appendFlag: true,
		},
		{
			name:     "duplicate accounts",
			appState: appStateWithAccounts,
			genesisAccount: []genutil.GenesisAccount{
				{
					Address: addresses[0],
					Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
				},
				{
					Address: addresses[0],
					Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
				},
				{
					Address: addresses[1],
					Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 2)),
				},
				{
					Address: addresses[1],
					Coins:   sdk.NewCoins(sdk.NewInt64Coin("stake", 1)),
				},
				{
					Address: addresses[1],
					Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
				},
			},
			appendFlag:    true,
			expectedError: "duplicate account",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appState, err := genutil.AddGenesisAccountsWithGenesis(cdc.Codec, addressCodec, tc.genesisAccount, tc.appendFlag, tc.appState)
			if tc.expectedError != "" {
				require.ErrorContains(t, err, tc.expectedError)
				return
			} else {
				require.NoError(t, err)
			}

			authGenState := authtypes.GetGenesisStateFromAppState(cdc.Codec, appState)
			bankGenState := banktypes.GetGenesisStateFromAppState(cdc.Codec, appState)

			require.Len(t, authGenState.Accounts, len(tc.expected))
			for _, acc := range tc.genesisAccount {
				addr := acc.Address
				accounts, _ := authtypes.UnpackAccounts(authGenState.Accounts)
				found := false
				for _, a := range accounts {
					if a.GetAddress().String() == addr {
						found = true
						break
					}
				}
				require.True(t, found, "account %s not found", addr)
			}

			expectedSupply := sdk.NewCoins()
			for _, coins := range tc.expected {
				expectedSupply = expectedSupply.Add(coins...)
			}
			require.Equal(t, expectedSupply.String(), bankGenState.Supply.String())
		})
	}
}
