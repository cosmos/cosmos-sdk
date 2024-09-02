package cli_test

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	corectx "cosmossdk.io/core/context"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/cosmos-sdk/client"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			logger := log.NewNopLogger()
			v := viper.New()

			encodingConfig := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{})
			appCodec := encodingConfig.Codec
			txConfig := encodingConfig.TxConfig
			err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err)

			err := writeAndTrackDefaultConfig(v, home)
			require.NoError(t, err)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home).
				WithAddressCodec(ac).WithTxConfig(txConfig)

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
