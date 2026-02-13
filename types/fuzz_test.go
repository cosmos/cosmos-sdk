package types

import (
	"sort"
	"testing"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
)

func FuzzCoinUnmarshalJSON(f *testing.F) {
	if testing.Short() {
		f.Skip()
	}

	cdc := codec.NewLegacyAmino()
	f.Add(`{"denom":"atom","amount":"1000"}`)
	f.Add(`{"denom":"atom","amount":"-1000"}`)
	f.Add(`{"denom":"uatom","amount":"1000111111111111111111111"}`)
	f.Add(`{"denom":"mu","amount":"0"}`)

	f.Fuzz(func(t *testing.T, jsonBlob string) {
		var c Coin
		_ = cdc.UnmarshalJSON([]byte(jsonBlob), &c)
	})
}

// FuzzCoinsAdd tests the Coins.Add method by comparing it against a reference
// implementation using a map-based approach.
func FuzzCoinsAdd(f *testing.F) {
	if testing.Short() {
		f.Skip()
	}

	// Seed corpus with interesting cases
	f.Add(uint8(2), uint8(2), uint8(3))    // small coins sets
	f.Add(uint8(0), uint8(0), uint8(5))    // empty + non-empty
	f.Add(uint8(5), uint8(5), uint8(2))    // larger sets, few denoms
	f.Add(uint8(10), uint8(10), uint8(10)) // many denoms

	f.Fuzz(func(t *testing.T, lenA, lenB, numDenoms uint8) {
		// Limit sizes to prevent excessive memory/time
		if lenA > 20 || lenB > 20 || numDenoms > 20 || numDenoms == 0 {
			return
		}

		// Generate denominations
		denoms := make([]string, numDenoms)
		for i := range denoms {
			denoms[i] = string(rune('a' + i))
		}

		// Generate random coins for set A
		coinsA := make(Coins, lenA)
		for i := range coinsA {
			coinsA[i] = Coin{
				Denom:  denoms[i%len(denoms)],
				Amount: math.NewInt(int64(i + 1)),
			}
		}

		// Generate random coins for set B
		coinsB := make(Coins, lenB)
		for i := range coinsB {
			coinsB[i] = Coin{
				Denom:  denoms[i%len(denoms)],
				Amount: math.NewInt(int64(i + 1)),
			}
		}

		// Sort both (required by safeAdd)
		sort.Sort(coinsA)
		sort.Sort(coinsB)

		// Run the actual implementation
		result := coinsA.Add(coinsB...)

		// Reference implementation using map
		expected := referenceCoinsAdd(coinsA, coinsB)

		// Compare results
		if !result.Equal(expected) {
			t.Errorf("Add mismatch:\n  A: %v\n  B: %v\n  got:  %v\n  want: %v",
				coinsA, coinsB, result, expected)
		}

		// Verify result is sorted
		if !result.isSorted() {
			t.Errorf("result not sorted: %v", result)
		}

		// Verify result has no zero coins
		for _, c := range result {
			if c.IsZero() {
				t.Errorf("result contains zero coin: %v", result)
			}
		}
	})
}

// referenceCoinsAdd is a simple map-based implementation for comparison
func referenceCoinsAdd(a, b Coins) Coins {
	m := make(map[string]math.Int)

	for _, c := range a {
		if existing, ok := m[c.Denom]; ok {
			m[c.Denom] = existing.Add(c.Amount)
		} else {
			m[c.Denom] = c.Amount
		}
	}

	for _, c := range b {
		if existing, ok := m[c.Denom]; ok {
			m[c.Denom] = existing.Add(c.Amount)
		} else {
			m[c.Denom] = c.Amount
		}
	}

	result := make(Coins, 0, len(m))
	for denom, amount := range m {
		if !amount.IsZero() {
			result = append(result, Coin{Denom: denom, Amount: amount})
		}
	}

	sort.Sort(result)
	return result
}
