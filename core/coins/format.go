package coins

import (
	"fmt"
	"sort"
	"strings"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/math"
)

// formatCoin formats a sdk.Coin into a value-rendered string, using the
// given metadata about the denom. It returns the formatted coin string, the
// display denom, and an optional error.
func formatCoin(coin *basev1beta1.Coin, metadata *bankv1beta1.Metadata) (string, error) {
	coinDenom := coin.Denom

	// Return early if no display denom or display denom is the current coin denom.
	if metadata == nil || metadata.Display == "" || coinDenom == metadata.Display {
		vr, err := math.FormatDec(coin.Amount)
		return vr + " " + coin.Denom, err
	}

	dispDenom := metadata.Display

	// Find exponents of both denoms.
	var coinExp, dispExp uint32
	foundCoinExp, foundDispExp := false, false
	for _, unit := range metadata.DenomUnits {
		if coinDenom == unit.Denom {
			coinExp = unit.Exponent
			foundCoinExp = true
		}
		if dispDenom == unit.Denom {
			dispExp = unit.Exponent
			foundDispExp = true
		}
	}

	// If we didn't find either exponent, then we return early.
	if !foundCoinExp || !foundDispExp {
		vr, err := math.FormatInt(coin.Amount)
		return vr + " " + coin.Denom, err
	}

	dispAmount, err := math.LegacyNewDecFromStr(coin.Amount)
	if err != nil {
		return "", err
	}

	if coinExp > dispExp {
		dispAmount = dispAmount.Mul(math.LegacyNewDec(10).Power(uint64(coinExp - dispExp)))
	} else {
		dispAmount = dispAmount.Quo(math.LegacyNewDec(10).Power(uint64(dispExp - coinExp)))
	}

	vr, err := math.FormatDec(dispAmount.String())
	return vr + " " + dispDenom, err
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

func ParseCoins(coins []string, metadata []*bankv1beta1.Metadata) ([]*basev1beta1.Coin, error) {
	if len(coins) != len(metadata) {
		return []*basev1beta1.Coin{}, fmt.Errorf("formatCoins expect one metadata for each coin; expected %d, got %d", len(coins), len(metadata))
	}

	parsedCoins := make([]*basev1beta1.Coin, len(coins), 0)
	for i, coinStr := range coins {
		coin, err := parseCoin(coinStr, metadata[i])
		if err != nil {
			return []*basev1beta1.Coin{}, err
		}
		parsedCoins = append(parsedCoins, coin)
	}

	return parsedCoins, nil
}

func parseCoin(coinStr string, metadata *bankv1beta1.Metadata) (*basev1beta1.Coin, error) {
	coinArr := strings.Split(coinStr, " ")
	coinDenom := coinArr[1]

	if metadata == nil || metadata.Display == "" || coinArr[1] == metadata.Display {
		dec, err := math.ParseDec(coinArr[0])
		return &basev1beta1.Coin{
			Amount: dec,
			Denom:  coinDenom,
		}, err
	}

	baseDenom := metadata.Display

	// Find exponents of both denoms.
	foundCoinExp, foundBaseExp := false, false
	var coinExp, baseExp uint32
	for _, unit := range metadata.DenomUnits {
		if coinDenom == unit.Denom {
			coinExp = unit.Exponent
			foundCoinExp = true
		}
		if baseDenom == unit.Denom {
			baseExp = unit.Exponent
			foundBaseExp = true
		}
	}

	// If we didn't find either exponent, then we return early.
	if !foundCoinExp || !foundBaseExp {
		amt, err := math.ParseInt(coinArr[0])
		return &basev1beta1.Coin{
			Amount: amt.String(),
			Denom:  coinDenom,
		}, err
	}

	amt, err := math.LegacyNewDecFromStr(coinArr[0])
	if err != nil {
		return &basev1beta1.Coin{}, err
	}

	if coinExp > baseExp {
		amt = amt.Mul(math.LegacyNewDec(10).Power(uint64(coinExp - baseExp)))
	} else {
		amt = amt.Quo(math.LegacyNewDec(10).Power(uint64(baseExp - coinExp)))
	}

	return &basev1beta1.Coin{
		Amount: amt.String(),
		Denom:  coinDenom,
	}, nil
}
