package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func TestMain(m *testing.M) {
	viper.Set(flags.FlagKeyringBackend, keyring.BackendMemory)
	os.Exit(m.Run())
}

func TestContext_PrintProto(t *testing.T) {
	ctx := client.Context{}

	animal := &testdata.Dog{
		Size_: "big",
		Name:  "Spot",
	}
	anyAnimal, err := types.NewAnyWithValue(animal)
	require.NoError(t, err)
	hasAnimal := &testdata.HasAnimal{
		Animal: anyAnimal,
		X:      10,
	}

	// proto
	registry := testdata.NewTestInterfaceRegistry()
	ctx = ctx.WithCodec(codec.NewProtoCodec(registry))

	// json
	buf := &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = flags.OutputFormatJSON
	err = ctx.PrintProto(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`{"animal":{"@type":"/testpb.Dog","size":"big","name":"Spot"},"x":"10"}
`, buf.String())

	// yaml
	buf = &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = flags.OutputFormatText
	err = ctx.PrintProto(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`animal:
  '@type': /testpb.Dog
  name: Spot
  size: big
x: "10"
`, buf.String())
}

func TestContext_PrintObjectLegacy(t *testing.T) {
	ctx := client.Context{}

	animal := &testdata.Dog{
		Size_: "big",
		Name:  "Spot",
	}
	anyAnimal, err := types.NewAnyWithValue(animal)
	require.NoError(t, err)
	hasAnimal := &testdata.HasAnimal{
		Animal: anyAnimal,
		X:      10,
	}

	// amino
	amino := testdata.NewTestAmino()
	ctx = ctx.WithLegacyAmino(&codec.LegacyAmino{Amino: amino})

	// json
	buf := &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = flags.OutputFormatJSON
	err = ctx.PrintObjectLegacy(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`{"type":"testpb/HasAnimal","value":{"animal":{"type":"testpb/Dog","value":{"size":"big","name":"Spot"}},"x":"10"}}
`, buf.String())

	// yaml
	buf = &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = flags.OutputFormatText
	err = ctx.PrintObjectLegacy(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`type: testpb/HasAnimal
value:
  animal:
    type: testpb/Dog
    value:
      name: Spot
      size: big
  x: "10"
`, buf.String())
}

func TestContext_PrintRaw(t *testing.T) {
	ctx := client.Context{}
	hasAnimal := json.RawMessage(`{"animal":{"@type":"/testpb.Dog","size":"big","name":"Spot"},"x":"10"}`)

	// json
	buf := &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = flags.OutputFormatJSON
	err := ctx.PrintRaw(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`{"animal":{"@type":"/testpb.Dog","size":"big","name":"Spot"},"x":"10"}
`, buf.String())

	// yaml
	buf = &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = flags.OutputFormatText
	err = ctx.PrintRaw(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`animal:
  '@type': /testpb.Dog
  name: Spot
  size: big
x: "10"
`, buf.String())
}

func TestGetFromFields(t *testing.T) {
	cfg := testutil.MakeTestEncodingConfig()
	path := hd.CreateHDPath(118, 0, 0).String()

	testCases := []struct {
		clientCtx   client.Context
		keyring     func() keyring.Keyring
		from        string
		expectedErr string
	}{
		{
			keyring: func() keyring.Keyring {
				kb := keyring.NewInMemory(cfg.Codec)

				_, _, err := kb.NewMnemonic("alice", keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
				require.NoError(t, err)

				return kb
			},
			from: "alice",
		},
		{
			keyring: func() keyring.Keyring {
				kb, err := keyring.New(t.Name(), keyring.BackendTest, t.TempDir(), nil, cfg.Codec)
				require.NoError(t, err)

				_, _, err = kb.NewMnemonic("alice", keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
				require.NoError(t, err)

				return kb
			},
			from: "alice",
		},
		{
			keyring: func() keyring.Keyring {
				return keyring.NewInMemory(cfg.Codec)
			},
			from:        "cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5",
			expectedErr: "key with address cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5 not found: key not found",
		},
		{
			keyring: func() keyring.Keyring {
				kb, err := keyring.New(t.Name(), keyring.BackendTest, t.TempDir(), nil, cfg.Codec)
				require.NoError(t, err)
				return kb
			},
			from:        "alice",
			expectedErr: "alice.info: key not found",
		},
		{
			keyring: func() keyring.Keyring {
				return keyring.NewInMemory(cfg.Codec)
			},
			from:      "cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5",
			clientCtx: client.Context{}.WithSimulation(true),
		},
		{
			keyring: func() keyring.Keyring {
				return keyring.NewInMemory(cfg.Codec)
			},
			from:        "alice",
			clientCtx:   client.Context{}.WithSimulation(true),
			expectedErr: "a valid bech32 address must be provided in simulation mode",
		},
		{
			keyring: func() keyring.Keyring {
				return keyring.NewInMemory(cfg.Codec)
			},
			from:      "cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5",
			clientCtx: client.Context{}.WithGenerateOnly(true),
		},
		{
			keyring: func() keyring.Keyring {
				return keyring.NewInMemory(cfg.Codec)
			},
			from:        "alice",
			clientCtx:   client.Context{}.WithGenerateOnly(true),
			expectedErr: "alice.info: key not found",
		},
		{
			keyring: func() keyring.Keyring {
				kb, err := keyring.New(t.Name(), keyring.BackendTest, t.TempDir(), nil, cfg.Codec)
				require.NoError(t, err)

				_, _, err = kb.NewMnemonic("alice", keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
				require.NoError(t, err)

				return kb
			},
			clientCtx: client.Context{}.WithGenerateOnly(true),
			from:      "alice",
		},
	}

	for _, tc := range testCases {
		_, _, _, err := client.GetFromFields(tc.clientCtx, tc.keyring(), tc.from)
		if tc.expectedErr == "" {
			require.NoError(t, err)
		} else {
			require.True(t, strings.HasPrefix(err.Error(), tc.expectedErr))
		}
	}
}

func TestGetFromFields_SignerAddress(t *testing.T) {
	cfg := testutil.MakeTestEncodingConfig()
	path := hd.CreateHDPath(118, 0, 0).String()
	kb := keyring.NewInMemory(cfg.Codec)
	rec, _, err := kb.NewMnemonic("alice", keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	keyAddr, err := rec.GetAddress()
	require.NoError(t, err)

	accountAddr := sdk.AccAddress("rekeyed_account_addr")
	require.NotEqual(t, keyAddr, accountAddr)

	// Without an override, the key's own address is returned.
	addr, name, _, err := client.GetFromFields(client.Context{}, kb, "alice")
	require.NoError(t, err)
	require.Equal(t, keyAddr, addr)
	require.Equal(t, "alice", name)

	// With an override, the key still signs but the account address is the
	// override.
	ctx := client.Context{}.WithSignerAddress(accountAddr)
	for _, from := range []string{"alice", keyAddr.String()} {
		addr, name, _, err = client.GetFromFields(ctx, kb, from)
		require.NoError(t, err)
		require.Equal(t, accountAddr, addr)
		require.Equal(t, "alice", name)
	}

	// The override also applies in generate-only mode.
	addr, name, _, err = client.GetFromFields(ctx.WithGenerateOnly(true), kb, "alice")
	require.NoError(t, err)
	require.Equal(t, accountAddr, addr)
	require.Equal(t, "alice", name)
}

// TestGetClientTxContext_SignerAddressWithPresetFrom checks that
// --signer-address replaces the from address even when the context already
// has a from key, so msgs and signing name the same account.
func TestGetClientTxContext_SignerAddressWithPresetFrom(t *testing.T) {
	cfg := testutil.MakeTestEncodingConfig()
	path := hd.CreateHDPath(118, 0, 0).String()
	kb := keyring.NewInMemory(cfg.Codec)
	rec, _, err := kb.NewMnemonic("alice", keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	keyAddr, err := rec.GetAddress()
	require.NoError(t, err)
	accountAddr := sdk.AccAddress("rekeyed_account_addr")

	initCtx := client.Context{}.WithKeyring(kb).WithFrom("alice").WithFromAddress(keyAddr).WithFromName("alice")

	var got client.Context
	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, _ []string) error {
			got, err = client.GetClientTxContext(cmd)
			return err
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.SetArgs([]string{"--" + flags.FlagSignerAddress + "=" + accountAddr.String()})
	require.NoError(t, cmd.ExecuteContext(context.WithValue(context.Background(), client.ClientContextKey, &initCtx)))

	require.Equal(t, accountAddr, got.SignerAddress)
	require.Equal(t, accountAddr, got.GetFromAddress())
	require.Equal(t, "alice", got.GetFromName())
}
