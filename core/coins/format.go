package coins

import (
	"fmt"
	"sort"
	"strings"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/math"
)

const EmptyCoins = "empty coins"

// ConvertAmt converts given coin to `toDenom` base coin using metadata, and returns dec as string.
func ConvertAmt(coin *basev1beta1.Coin, metadata *bankv1beta1.Metadata, toDenom string) (string, string, error) {
	amt1 := coin.Amount
	coinDenom := coin.Denom

	// Return early if no display denom or display denom is the current coin denom.
	if metadata == nil || toDenom == "" || coinDenom == toDenom {
		dec, err := math.FormatDec(amt1)
		return dec, coinDenom, err
	}

	// Find exponents of both denoms.
	foundCoinExp, foundToDenomExp := false, false
	var coinExp, ToDenomExp uint32
	for _, unit := range metadata.DenomUnits {
		if coinDenom == unit.Denom {
			coinExp = unit.Exponent
			foundCoinExp = true
		}
		if toDenom == unit.Denom {
			ToDenomExp = unit.Exponent
			foundToDenomExp = true
		}
	}

	// If we didn't find either exponent, then we return early.
	if !foundCoinExp || !foundToDenomExp {
		dec, err := math.FormatDec(amt1)
		return dec, coinDenom, err
	}

	amount, err := math.LegacyNewDecFromStr(coin.Amount)
	if err != nil {
		return "", "", err
	}

	if coinExp > ToDenomExp {
		amount = amount.Mul(math.LegacyNewDec(10).Power(uint64(coinExp - ToDenomExp)))
	} else {
		amount = amount.Quo(math.LegacyNewDec(10).Power(uint64(ToDenomExp - coinExp)))
	}

	amt, err := math.FormatDec(amount.String())
	return amt, toDenom, err
}

// formatCoin formats a sdk.Coin into a value-rendered string, using the
// given metadata about the denom. It returns the formatted coin string, the
// display denom, and an optional error.
func formatCoin(coin *basev1beta1.Coin, metadata *bankv1beta1.Metadata) (string, error) {
	var display = coin.Denom
	if metadata != nil {
		display = metadata.Display
	}

	vr, denom, err := ConvertAmt(coin, metadata, display)

	// vr, err := math.FormatDec(dispAmount.String())
	return vr + " " + denom, err
}

// formatCoins formats Coins into a value-rendered string, which uses
// `formatCoin` separated by ", " (a comma and a space), and sorted
// alphabetically by value-rendered denoms. It expects an array of metadata
// (optionally nil), where each metadata at index `i` MUST match the coin denom
// at the same index.
func FormatCoins(coins []*basev1beta1.Coin, metadata []*bankv1beta1.Metadata) (string, error) {
	if len(coins) != len(metadata) {
		return "", fmt.Errorf("formatCoins expect one metadata for each coin; expected %d, got %d", len(coins), len(metadata))
	}

	if len(coins) == 0 {
		return EmptyCoins, nil
	}

	formatted := make([]string, len(coins))
	for i, coin := range coins {
		var err error
		formatted[i], err = formatCoin(coin, metadata[i])
		if err != nil {
			return "", err
		}
	}

	// Sort the formatted coins by display denom.
	sort.SliceStable(formatted, func(i, j int) bool {
		denomI := strings.Split(formatted[i], " ")[1]
		denomJ := strings.Split(formatted[j], " ")[1]

		return denomI < denomJ
	})

	return strings.Join(formatted, ", "), nil
}
