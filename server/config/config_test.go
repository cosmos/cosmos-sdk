package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.True(t, cfg.MinimumFees().IsZero())
}

func TestSetMinimumFees(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SetMinimumFees(sdk.Coins{sdk.NewCoin("foo", sdk.NewInt(100))})
	require.Equal(t, "100foo", cfg.MinFees)
}
