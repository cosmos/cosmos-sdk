package client_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/testdata"
	"github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

func TestContext_WithOffline(t *testing.T) {
	viper.Set(flags.FlagOffline, true)
	viper.Set(flags.FlagNode, "tcp://localhost:26657")

	ctx := client.NewContext()
	require.True(t, ctx.Offline)
	require.Nil(t, ctx.Client)
}

func TestContext_WithGenOnly(t *testing.T) {
	viper.Set(flags.FlagGenerateOnly, true)

	validFromAddr := "cosmos1q7380u26f7ntke3facjmynajs4umlr329vr4ja"
	fromAddr, err := sdk.AccAddressFromBech32(validFromAddr)
	require.NoError(t, err)

	tests := []struct {
		name             string
		from             string
		expectedFromAddr sdk.AccAddress
		expectedFromName string
	}{
		{
			name:             "valid from",
			from:             validFromAddr,
			expectedFromAddr: fromAddr,
			expectedFromName: "",
		},
		{
			name:             "empty from",
			from:             "",
			expectedFromAddr: nil,
			expectedFromName: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx := client.NewContextWithFrom(tt.from)

			require.Equal(t, tt.expectedFromAddr, ctx.FromAddress)
			require.Equal(t, tt.expectedFromName, ctx.FromName)
		})
	}
}

func TestContext_WithKeyring(t *testing.T) {
	viper.Set(flags.FlagGenerateOnly, true)
	ctx := client.NewContextWithFrom("cosmos1q7380u26f7ntke3facjmynajs4umlr329vr4ja")
	require.NotNil(t, ctx.Keyring)
	kr := ctx.Keyring
	ctx = ctx.WithKeyring(nil)
	require.Nil(t, ctx.Keyring)
	ctx = ctx.WithKeyring(kr)
	require.Equal(t, kr, ctx.Keyring)
}

func TestMain(m *testing.M) {
	viper.Set(flags.FlagKeyringBackend, keyring.BackendMemory)
	os.Exit(m.Run())
}

func TestContext_PrintOutput(t *testing.T) {
	ctx := client.Context{}

	animal := &testdata.Dog{
		Size_: "big",
		Name:  "Spot",
	}
	any, err := types.NewAnyWithValue(animal)
	require.NoError(t, err)
	hasAnimal := &testdata.HasAnimal{
		Animal: any,
		X:      10,
	}

	//
	// proto
	//
	registry := testdata.NewTestInterfaceRegistry()
	ctx = ctx.WithJSONMarshaler(codec.NewProtoCodec(registry))

	// json
	buf := &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = "json"
	err = ctx.PrintOutput(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`{"animal":{"@type":"/testdata.Dog","size":"big","name":"Spot"},"x":"10"}
`, string(buf.Bytes()))

	// yaml
	buf = &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = "text"
	err = ctx.PrintOutput(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`animal:
  '@type': /testdata.Dog
  name: Spot
  size: big
x: "10"
`, string(buf.Bytes()))

	//
	// amino
	//
	amino := testdata.NewTestAmino()
	ctx = ctx.WithJSONMarshaler(codec.NewAminoCodec(&codec.Codec{Amino: amino}))

	// json
	buf = &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = "json"
	err = ctx.PrintOutput(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`{"type":"testdata/HasAnimal","value":{"animal":{"type":"testdata/Dog","value":{"size":"big","name":"Spot"}},"x":"10"}}
`, string(buf.Bytes()))

	// yaml
	buf = &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = "text"
	err = ctx.PrintOutput(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`type: testdata/HasAnimal
value:
  animal:
    type: testdata/Dog
    value:
      name: Spot
      size: big
  x: "10"
`, string(buf.Bytes()))
}
