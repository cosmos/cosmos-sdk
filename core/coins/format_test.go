package coins_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/coins"
)

func TestDecodeCoin(t *testing.T) {
	encodedCoin := "1000000000foo"
	coin, err := coins.ParseCoin(encodedCoin)
	require.NoError(t, err)
	require.Equal(t, "1000000000", coin.Amount)
	require.Equal(t, "foo", coin.Denom)
}
