package strings

import (
	"fmt"
	"strings"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
)

func CoinsAsString(coins []*basev1beta1.Coin) string {
	if len(coins) == 0 {
		return ""
	} else if len(coins) == 1 {
		return CoinAsString(coins[0])
	}

	// Build the string with a string builder
	var out strings.Builder
	for _, coin := range coins[:len(coins)-1] {
		out.WriteString(coin.String())
		out.WriteByte(',')
	}
	out.WriteString(CoinAsString(coins[len(coins)-1]))
	return out.String()
}

func CoinAsString(coin *basev1beta1.Coin) string {
	return fmt.Sprintf("%v%s", coin.Amount, coin.Denom)
}
