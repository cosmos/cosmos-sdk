package client_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func TestMain(m *testing.M) {
	viper.Set(flags.FlagKeyringBackend, keyring.BackendMemory)
	os.Exit(m.Run())
}

func TestContext_PrintObject(t *testing.T) {
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
	ctx = ctx.WithCodec(codec.NewProtoCodec(registry))

	// json
	buf := &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = "json"
	err = ctx.PrintProto(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`{"animal":{"@type":"/testdata.Dog","size":"big","name":"Spot"},"x":"10"}
`, buf.String())

	// yaml
	buf = &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = "text"
	err = ctx.PrintProto(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`animal:
  '@type': /testdata.Dog
  name: Spot
  size: big
x: "10"
`, buf.String())

	//
	// amino
	//
	amino := testdata.NewTestAmino()
	ctx = ctx.WithLegacyAmino(&codec.LegacyAmino{Amino: amino})

	// json
	buf = &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = "json"
	err = ctx.PrintObjectLegacy(hasAnimal)
	require.NoError(t, err)
	require.Equal(t,
		`{"type":"testdata/HasAnimal","value":{"animal":{"type":"testdata/Dog","value":{"size":"big","name":"Spot"}},"x":"10"}}
`, buf.String())

	// yaml
	buf = &bytes.Buffer{}
	ctx = ctx.WithOutput(buf)
	ctx.OutputFormat = "text"
	err = ctx.PrintObjectLegacy(hasAnimal)
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
`, buf.String())
}

func TestCLIQueryConn(t *testing.T) {
	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	n := network.New(t, cfg)
	defer n.Cleanup()

	testClient := testdata.NewQueryClient(n.Validators[0].ClientCtx)
	res, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", res.Message)
}
