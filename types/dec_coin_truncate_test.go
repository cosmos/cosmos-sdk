package types_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDecCoinsTruncateDecimalCompatibility(t *testing.T) {
	coin := func(denom, amount string) sdk.DecCoin {
		return sdk.NewDecCoinFromDec(denom, math.LegacyMustNewDecFromStr(amount))
	}
	for _, tc := range []struct {
		name  string
		input sdk.DecCoins
	}{
		{"nil", nil},
		{"empty", sdk.DecCoins{}},
		{"zero", sdk.DecCoins{coin("foo", "0")}},
		{"integers", sdk.DecCoins{coin("bar", "1"), coin("foo", "2")}},
		{"fractions", sdk.DecCoins{coin("bar", "0.000000000000000001"), coin("foo", "0.99")}},
		{"mixed", sdk.DecCoins{coin("bar", "1.25"), coin("baz", "0"), coin("foo", "0.5")}},
		{"unsorted", sdk.DecCoins{coin("foo", "2.75"), coin("bar", "1.5"), coin("baz", "3.25")}},
		// Truncate each coin before combining duplicates: two 1.75foo coins
		// produce 2foo and 1.5foo, rather than 3foo and 0.5foo.
		{"duplicates", sdk.DecCoins{coin("foo", "1.75"), coin("foo", "1.75")}},
		{"nonadjacent duplicates", sdk.DecCoins{coin("foo", "1.75"), coin("bar", "0.5"), coin("foo", "1.75")}},
		{"different output denoms", sdk.DecCoins{coin("foo", "2"), coin("bar", "0.5"), coin("baz", "3.25")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := tc.input.String()
			wantCoins, wantChange := truncateDecimalByAddition(tc.input)
			coins, change := tc.input.TruncateDecimal()
			require.Equal(t, wantCoins, coins)
			require.Equal(t, wantChange, change)
			require.Equal(t, before, tc.input.String())
			// Both outputs must own their amounts, even when no addition was needed.
			for _, c := range coins {
				c.Amount.BigIntMut().SetInt64(0)
			}
			for _, c := range change {
				c.Amount.BigIntMut().SetInt64(0)
			}
			require.Equal(t, before, tc.input.String())
		})
	}
}

func TestDecCoinsTruncateDecimalInvalid(t *testing.T) {
	maxInt := math.NewIntFromBigInt(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)))
	for _, tc := range []struct {
		name  string
		input sdk.DecCoins
	}{
		{"invalid denom", sdk.DecCoins{{Denom: "!", Amount: math.LegacyOneDec()}}},
		{"negative integer", sdk.DecCoins{{Denom: "foo", Amount: math.LegacyNewDec(-1)}}},
		{"negative fraction", sdk.DecCoins{{Denom: "foo", Amount: math.LegacyNewDecWithPrec(-1, 1)}}},
		{"nil amount", sdk.DecCoins{{Denom: "foo"}}},
		{"duplicate overflow", sdk.DecCoins{sdk.NewDecCoin("foo", maxInt), sdk.NewInt64DecCoin("foo", 1)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Panics(t, func() { truncateDecimalByAddition(tc.input) })
			require.Panics(t, func() { tc.input.TruncateDecimal() })
		})
	}
}

// Keep the previous accumulation algorithm as a reference for ordering,
// duplicate denominations, zero filtering and per-coin rounding.
func truncateDecimalByAddition(input sdk.DecCoins) (coins sdk.Coins, change sdk.DecCoins) {
	for _, coin := range input {
		integer, fraction := coin.TruncateDecimal()
		if !integer.IsZero() {
			coins = coins.Add(integer)
		}
		if !fraction.IsZero() {
			change = change.Add(fraction)
		}
	}
	return coins, change
}

func FuzzDecCoinsTruncateDecimal(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 0, 1, 10, 2, 19})
	f.Add([]byte{2, 17, 0, 5, 2, 17, 1, 20})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 200 {
			data = data[:200]
		}
		var input sdk.DecCoins
		for i := 0; i+1 < len(data); i += 2 {
			input = append(input, sdk.NewDecCoinFromDec(
				fmt.Sprintf("token%02d", data[i]%16), math.LegacyNewDecWithPrec(int64(data[i+1]), 1),
			))
		}
		wantCoins, wantChange := truncateDecimalByAddition(input)
		coins, change := input.TruncateDecimal()
		require.Equal(t, wantCoins, coins)
		require.Equal(t, wantChange, change)
	})
}

func BenchmarkDecCoinsTruncateDecimal(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			coins := make(sdk.DecCoins, n)
			for i := range coins {
				coins[i] = sdk.NewDecCoinFromDec(fmt.Sprintf("token%03d", i), math.LegacyNewDecWithPrec(125, 2))
			}
			b.ReportAllocs()
			for b.Loop() {
				coins.TruncateDecimal()
			}
		})
	}
}
