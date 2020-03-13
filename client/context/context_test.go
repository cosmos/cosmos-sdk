package context_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	crkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
)

func TestCLIContext_WithKeybase(t *testing.T) {
	ctx := context.CLIContext{}
	kb := keys.NewInMemoryKeyBase()
	require.Nil(t, ctx.Keybase)
	ctx = ctx.WithKeybase(kb)
	require.NotNil(t, ctx.Keybase)
}

func TestGetFromFields(t *testing.T) {
	kb := keys.NewInMemoryKeyBase()
	info, _, err := kb.CreateMnemonic("test", crkeys.English, "", crkeys.Secp256k1)
	require.NoError(t, err)

	got1, got2, err := context.GetFromFields("test", kb)
	require.True(t, info.GetAddress().Equals(got1))
	require.Equal(t, "test", got2)
	require.NoError(t, err)

	bech32fied := info.GetAddress().String()
	got1, got2, err = context.GetFromFields(bech32fied, nil)
	require.True(t, info.GetAddress().Equals(got1))
	require.Empty(t, got2)
	require.NoError(t, err)

	got1, got2, err = context.GetFromFields("invalid", nil)
	require.Nil(t, got1)
	require.Empty(t, got2)
	require.Error(t, err)
}
