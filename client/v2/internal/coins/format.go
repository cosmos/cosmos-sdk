package coins

import (
	"errors"
	"regexp"
	"strings"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
)

// Amount can be a whole number or a decimal number. Denominations can be 3 ~ 128
// characters long and support letters, followed by either a letter, a number or
// a separator ('/', ':', '.', '_' or '-').
var coinRegex = regexp.MustCompile(`^(\d+(\.\d+)?)([a-zA-Z][a-zA-Z0-9\/\:\._\-]{2,127})$`)

// ParseCoin parses a coin from a string. The string must be in the format
// <amount><denom>, where <amount> is a number and <denom> is a valid denom.
func ParseCoin(input string) (*basev1beta1.Coin, error) {
	amount, denom, err := parseCoin(input)
	if err != nil {
		return nil, err
	}

	return &basev1beta1.Coin{
		Amount: amount,
		Denom:  denom,
	}, nil
}

// ParseDecCoin parses a decCoin from a string. The string must be in the format
// <amount><denom>, where <amount> is a number and <denom> is a valid denom.
func ParseDecCoin(input string) (*basev1beta1.DecCoin, error) {
	amount, denom, err := parseCoin(input)
	if err != nil {
		return nil, err
	}

	return &basev1beta1.DecCoin{
		Amount: amount,
		Denom:  denom,
	}, nil
}

// parseCoin parses a coin string into its amount and denom components.
// The input string must be in the format <amount><denom>.
// It returns the amount string, denom string, and any error encountered.
// Returns an error if the input is empty or doesn't match the expected format.
func parseCoin(input string) (amount, denom string, err error) {
	input = strings.TrimSpace(input)

	if input == "" {
		return "", "", errors.New("empty input when parsing coin")
	}

	matches := coinRegex.FindStringSubmatch(input)

	if len(matches) == 0 {
		return "", "", errors.New("invalid input format")
	}

	return matches[1], matches[3], nil
}
