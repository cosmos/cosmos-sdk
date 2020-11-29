package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
)

var coinStrs = []string{
	"2000ATM",
	"5000AMX",
	"192XXX",
	"1e9BTC",
}

func BenchmarkParseCoin(b *testing.B) {
	var blankCoin types.Coin
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, coinStr := range coinStrs {
			coin, err := types.ParseCoinNormalized(coinStr)
			if err != nil {
				b.Fatal(err)
			}
			if coin == blankCoin {
				b.Fatal("Unexpectedly returned a blank coin")
			}
		}
	}
}
