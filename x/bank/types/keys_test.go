package types

import (
	"testing"

	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestBalanceValueCodec(t *testing.T) {
	c := NewBalanceCompatValueCodec()
	t.Run("value codec implementation", func(t *testing.T) {
		colltest.TestValueCodec(t, c, math.NewInt(100))
	})

	t.Run("legacy coin", func(t *testing.T) {
		coin := sdk.NewInt64Coin("coin", 1000)
		b, err := coin.Marshal()
		require.NoError(t, err)
		amt, err := c.Decode(b)
		require.NoError(t, err)
		require.Equal(t, coin.Amount, amt)
	})
}
