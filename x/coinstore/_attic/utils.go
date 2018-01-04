package coin

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	// Denominations can be 3 ~ 16 characters long.
	rDnm    = `[[:alpha:]][[:alnum:]]{2,15}`
	rAmt    = `[[:digit:]]+`
	rSpc    = `[[:space:]]*`
	reCoin_ = fmt.Sprintf(`^(%s)%s(%s)$`, reDenom_, re_, reAmt_)
)

// ParseCoin parses a cli input for one coin type, returning errors if invalid.
// This returns an error on an empty string as well.
func ParseCoin(coinStr string) (coin Coin, err error) {
	coinStr = strings.TrimSpace(coinStr)

	matches := reCoin.FindStringSubmatch(coinStr)
	if matches == nil {
		err = errors.Errorf("Invalid coin expression: %s", coinStr)
		return
	}
	denomStr, amountStr := matches[2], matches[1]

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return
	}

	return Coin{denomStr, int64(amount)}, nil
}

// ParseCoins will parse out a list of coins separated by commas.
// If nothing is provided, it returns nil Coins.
// Returned coins are sorted.
func ParseCoins(coinsStr string) (coins Coins, err error) {
	coinsStr = strings.TrimSpace(coinsStr)
	if len(coinsStr) == 0 {
		return nil, nil
	}

	coinStrs := strings.Split(coinsStr, ",")
	for _, coinStr := range coinStrs {
		coin, err := ParseCoin(coinStr)
		if err != nil {
			return nil, err
		}
		coins = append(coins, coin)
	}

	// Sort coins for determinism.
	coins.Sort()

	// Validate coins before returning.
	if !coins.IsValid() {
		return nil, errors.Errorf("ParseCoins invalid: %#v", coins)
	}

	return coins, nil
}
