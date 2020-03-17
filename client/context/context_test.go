package context

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

func TestCLIContext_WithOffline(t *testing.T) {
	viper.Set(flags.FlagOffline, true)
	viper.Set(flags.FlagNode, "tcp://localhost:26657")

	ctx := NewCLIContext()
	require.True(t, ctx.Offline)
	require.Nil(t, ctx.Client)

	viper.Reset()

	viper.Set(flags.FlagOffline, false)
	viper.Set(flags.FlagNode, "tcp://localhost:26657")

	ctx = NewCLIContext()
	require.False(t, ctx.Offline)
	require.NotNil(t, ctx.Client)
}
