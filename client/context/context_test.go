package context

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

func TestCLIContext_WithGenOnly(t *testing.T) {
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
			ctx := NewCLIContextWithFrom(tt.from)

			require.Equal(t, tt.expectedFromAddr, ctx.FromAddress)
			require.Equal(t, tt.expectedFromName, ctx.FromName)
		})
	}
}
