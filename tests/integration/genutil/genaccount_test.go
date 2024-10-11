package genutil

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	corectx "cosmossdk.io/core/context"
	"cosmossdk.io/log"
	banktypes "cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	genutilhelpers "github.com/cosmos/cosmos-sdk/testutil/x/genutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func TestAddGenesisAccountCmd(t *testing.T) {
	_, _, addr1 := testdata.KeyTestPubAddr()
	ac := codectestutil.CodecOptions{}.GetAddressCodec()
	addr1Str, err := ac.BytesToString(addr1)
	require.NoError(t, err)

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
			addr:        addr1Str,
			denom:       "1000atom",
			withKeyring: false,
			expectErr:   false,
		},
		{
			name:        "multiple denoms",
			addr:        addr1Str,
			denom:       "1000atom, 2000stake",
			withKeyring: false,
			expectErr:   false,
		},
		{
			name:        "with keyring",
			addr:        "set",
			denom:       "1000atom",
			withKeyring: true,
			expectErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			logger := log.NewNopLogger()
			v := viper.New()

			encodingConfig := moduletestutil.MakeTestEncodingConfig(
				codectestutil.CodecOptions{},
				auth.AppModule{},
			)
			appCodec := encodingConfig.Codec
			txConfig := encodingConfig.TxConfig
			err = genutilhelpers.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err)

			err := writeAndTrackDefaultConfig(v, home)
			require.NoError(t, err)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home).
				WithAddressCodec(ac).WithTxConfig(txConfig)

			if tc.withKeyring {
				path := hd.CreateHDPath(118, 0, 0).String()
				kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendMemory, home, nil, appCodec)
				require.NoError(t, err)
				_, _, err = kr.NewMnemonic(
					tc.addr,
					keyring.English,
					path,
					keyring.DefaultBIP39Passphrase,
					hd.Secp256k1,
				)
				require.NoError(t, err)
				clientCtx = clientCtx.WithKeyring(kr)
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, corectx.ViperContextKey, v)
			ctx = context.WithValue(ctx, corectx.LoggerContextKey, logger)

			cmd := genutilcli.AddGenesisAccountCmd()
			cmd.SetArgs([]string{
				tc.addr,
				tc.denom,
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
	ac := codectestutil.CodecOptions{}.GetAddressCodec()
	_, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	_, _, addr3 := testdata.KeyTestPubAddr()
	addr1Str, err := ac.BytesToString(addr1)
	require.NoError(t, err)
	addr2Str, err := ac.BytesToString(addr2)
	require.NoError(t, err)
	addr3Str, err := ac.BytesToString(addr3)
	require.NoError(t, err)

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
			v := viper.New()

			encodingConfig := moduletestutil.MakeTestEncodingConfig(
				codectestutil.CodecOptions{},
				auth.AppModule{},
			)
			appCodec := encodingConfig.Codec
			txConfig := encodingConfig.TxConfig
			err = genutilhelpers.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err)

			err = writeAndTrackDefaultConfig(v, home)
			require.NoError(t, err)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home).
				WithAddressCodec(ac).WithTxConfig(txConfig)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, corectx.ViperContextKey, v)
			ctx = context.WithValue(ctx, corectx.LoggerContextKey, logger)

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

				cmd := genutilcli.AddBulkGenesisAccountCmd()
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

			bankState := banktypes.GetGenesisStateFromAppState(encodingConfig.Codec, appState)

			require.EqualValues(t, len(tc.expected), len(bankState.Balances))
			for _, acc := range bankState.Balances {
				require.True(
					t,
					tc.expected[acc.Address].Equal(acc.Coins),
					"expected: %v, got: %v",
					tc.expected[acc.Address],
					acc.Coins,
				)
			}

			expectedSupply := sdk.NewCoins()
			for _, coins := range tc.expected {
				expectedSupply = expectedSupply.Add(coins...)
			}
			require.Equal(t, expectedSupply, bankState.Supply)
		})
	}
}
