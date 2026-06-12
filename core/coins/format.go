package coins

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"cosmossdk.io/math"
)

const emptyCoins = "zero"

// Amount can be a whole number or a decimal number. Denominations can be 3 ~ 128
// characters long and support letters, followed by either a letter, a number or
// a separator ('/', ':', '.', '_' or '-').
var coinRegex = regexp.MustCompile(`^(\d+(\.\d+)?)([a-zA-Z][a-zA-Z0-9\/\:\._\-]{2,127})$`)

// Coin is a denomination and amount pair used for coin formatting.
// It mirrors cosmos.base.v1beta1.Coin's Denom/Amount string fields.
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// DecCoin is a denomination and decimal amount pair.
type DecCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// DenomUnit defines a denomination unit with an exponent.
type DenomUnit struct {
	Denom    string   `json:"denom"`
	Exponent uint32   `json:"exponent"`
	Aliases  []string `json:"aliases,omitempty"`
}

// Metadata holds denomination display metadata.
type Metadata struct {
	Display    string       `json:"display"`
	Base       string       `json:"base,omitempty"`
	DenomUnits []*DenomUnit `json:"denom_units,omitempty"`
}

// formatCoin formats a Coin into a value-rendered string, using the
// given metadata about the denom.
func formatCoin(coin *Coin, metadata *Metadata) (string, error) {
	coinDenom := coin.Denom

	if metadata == nil || metadata.Display == "" || coinDenom == metadata.Display {
		vr, err := math.FormatDec(coin.Amount)
		return vr + " " + coin.Denom, err
	}

	dispDenom := metadata.Display

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

// FormatCoins formats Coins into a value-rendered string, sorted alphabetically.
// Each metadata at index i MUST match the coin denom at the same index.
func FormatCoins(coins []*Coin, metadata []*Metadata) (string, error) {
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

		if strings.Contains(formatted[i], ",") {
			return "", fmt.Errorf("coin %s contains a comma", formatted[i])
		}
	}

	if len(coins) == 0 {
		return emptyCoins, nil
	}

	sort.SliceStable(formatted, func(i, j int) bool {
		denomI := strings.Split(formatted[i], " ")[1]
		denomJ := strings.Split(formatted[j], " ")[1]
		return denomI < denomJ
	})

	return strings.Join(formatted, ", "), nil
}

// ParseCoin parses a coin from a string in the format <amount><denom>.
func ParseCoin(input string) (*Coin, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input when parsing coin")
	}

	matches := coinRegex.FindStringSubmatch(input)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid input format")
	}

	return &Coin{Amount: matches[1], Denom: matches[3]}, nil
}

// ParseDecCoin parses a decimal coin from a string in the format <amount><denom>.
func ParseDecCoin(input string) (*DecCoin, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input when parsing dec coin")
	}

	matches := coinRegex.FindStringSubmatch(input)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid input format")
	}

	return &DecCoin{Amount: matches[1], Denom: matches[3]}, nil
}
