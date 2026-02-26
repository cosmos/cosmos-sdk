package cli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func TestAddGenesisAccountCmd(t *testing.T) {
	_, _, addr1 := testdata.KeyTestPubAddr()
	tests := []struct {
		name        string
		addr        string
		denom       string
		withKeyring bool
		expectErr   bool
	}{
		{
			name:        "invalid address",
			addr:        "",
			denom:       "1000atom",
			withKeyring: false,
			expectErr:   true,
		},
		{
			name:        "valid address",
			addr:        addr1.String(),
			denom:       "1000atom",
			withKeyring: false,
			expectErr:   false,
		},
		{
			name:        "multiple denoms",
			addr:        addr1.String(),
			denom:       "1000atom, 2000stake",
			withKeyring: false,
			expectErr:   false,
		},
		{
			name:        "with keyring",
			addr:        "ser",
			denom:       "1000atom",
			withKeyring: true,
			expectErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			logger := log.NewNopLogger()
			cfg, err := genutiltest.CreateDefaultCometConfig(home)
			require.NoError(t, err)

			appCodec := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}).Codec
			err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err)

			serverCtx := server.NewContext(viper.New(), cfg, logger)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home)

			if tc.withKeyring {
				path := hd.CreateHDPath(118, 0, 0).String()
				kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendMemory, home, nil, appCodec)
				require.NoError(t, err)
				_, _, err = kr.NewMnemonic(tc.addr, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
				require.NoError(t, err)
				clientCtx = clientCtx.WithKeyring(kr)
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			cmd := genutilcli.AddGenesisAccountCmd(home, addresscodec.NewBech32Codec("cosmos"))
			cmd.SetArgs([]string{
				tc.addr,
				tc.denom,
				fmt.Sprintf("--%s=home", flags.FlagHome),
			})

			if tc.expectErr {
				require.Error(t, cmd.ExecuteContext(ctx))
			} else {
				require.NoError(t, cmd.ExecuteContext(ctx))
			}
		})
	}
}

func TestBulkAddGenesisAccountCmd(t *testing.T) {
	_, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	_, _, addr3 := testdata.KeyTestPubAddr()
	addr1Str := addr1.String()
	addr2Str := addr2.String()
	addr3Str := addr3.String()

	tests := []struct {
		name       string
		state      [][]genutil.GenesisAccount
		expected   map[string]sdk.Coins
		appendFlag bool
		expectErr  bool
	}{
		{
			name: "invalid address",
			state: [][]genutil.GenesisAccount{
				{
					{
						Address: "invalid",
						Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
					},
				},
			},
			expectErr: true,
		},
		{
			name: "no append flag for multiple account adds",
			state: [][]genutil.GenesisAccount{
				{
					{
						Address: addr1Str,
						Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
					},
				},
				{
					{
						Address: addr1Str,
						Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 2)),
					},
				},
			},
			appendFlag: false,
			expectErr:  true,
		},

		{
			name: "multiple additions with append",
			state: [][]genutil.GenesisAccount{
				{
					{
						Address: addr1Str,
						Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
					},
					{
						Address: addr2Str,
						Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
					},
				},
				{
					{
						Address: addr1Str,
						Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 2)),
					},
					{
						Address: addr2Str,
						Coins:   sdk.NewCoins(sdk.NewInt64Coin("stake", 1)),
					},
					{
						Address: addr3Str,
						Coins:   sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
					},
				},
			},
			expected: map[string]sdk.Coins{
				addr1Str: sdk.NewCoins(sdk.NewInt64Coin("test", 3)),
				addr2Str: sdk.NewCoins(sdk.NewInt64Coin("test", 1), sdk.NewInt64Coin("stake", 1)),
				addr3Str: sdk.NewCoins(sdk.NewInt64Coin("test", 1)),
			},
			appendFlag: true,
			expectErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			logger := log.NewNopLogger()
			cfg, err := genutiltest.CreateDefaultCometConfig(home)
			require.NoError(t, err)

			appCodec := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}).Codec
			err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err)

			serverCtx := server.NewContext(viper.New(), cfg, logger)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			// The first iteration (pre-append) may not error.
			// Check if any errors after all state transitions to genesis.
			doesErr := false

			// apply multiple state iterations if applicable (e.g. --append)
			for _, state := range tc.state {
				bz, err := json.Marshal(state)
				require.NoError(t, err)

				filePath := path.Join(home, "accounts.json")
				err = os.WriteFile(filePath, bz, 0o600)
				require.NoError(t, err)

				cmd := genutilcli.AddBulkGenesisAccountCmd(home, addresscodec.NewBech32Codec("cosmos"))
				args := []string{filePath}
				if tc.appendFlag {
					args = append(args, "--append")
				}
				cmd.SetArgs(args)

				err = cmd.ExecuteContext(ctx)
				if err != nil {
					doesErr = true
				}
			}
			require.Equal(t, tc.expectErr, doesErr)

			// an error already occurred, no need to check the state
			if doesErr {
				return
			}

			appState, _, err := genutiltypes.GenesisStateFromGenFile(path.Join(home, "config", "genesis.json"))
			require.NoError(t, err)

			bankState := banktypes.GetGenesisStateFromAppState(appCodec, appState)

			require.EqualValues(t, len(tc.expected), len(bankState.Balances))
			for _, acc := range bankState.Balances {
				require.True(t, tc.expected[acc.Address].Equal(acc.Coins), "expected: %v, got: %v", tc.expected[acc.Address], acc.Coins)
			}

			expectedSupply := sdk.NewCoins()
			for _, coins := range tc.expected {
				expectedSupply = expectedSupply.Add(coins...)
			}
			require.Equal(t, expectedSupply, bankState.Supply)
		})
	}
}
