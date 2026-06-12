package coins

import (
	"errors"
	"regexp"
	"strings"

	coinscore "cosmossdk.io/core/coins"
)

// Amount can be a whole number or a decimal number. Denominations can be 3 ~ 128
// characters long and support letters, followed by either a letter, a number or
// a separator ('/', ':', '.', '_' or '-').
var coinRegex = regexp.MustCompile(`^(\d+(\.\d+)?)([a-zA-Z][a-zA-Z0-9\/\:\._\-]{2,127})$`)

// ParseCoin parses a coin from a string in the format <amount><denom>.
func ParseCoin(input string) (*coinscore.Coin, error) {
	amount, denom, err := parseCoin(input)
	if err != nil {
		return nil, err
	}
	return &coinscore.Coin{Amount: amount, Denom: denom}, nil
}

// ParseDecCoin parses a dec coin from a string in the format <amount><denom>.
func ParseDecCoin(input string) (*coinscore.DecCoin, error) {
	amount, denom, err := parseCoin(input)
	if err != nil {
		return nil, err
	}
	return &coinscore.DecCoin{Amount: amount, Denom: denom}, nil
}

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
