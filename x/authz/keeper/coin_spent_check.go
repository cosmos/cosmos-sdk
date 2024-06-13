package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CoinsSpentEvents(evts sdk.Events, spender string) (sdk.Coins, error) {
	coinsSpent := sdk.Coins{}

	for _, evt := range evts {
		if evt.Type == "coin_spent" {
			for i := 0; i < len(evt.Attributes); i++ {
				attr := evt.Attributes[i]
				if attr.Key == "spender" && attr.Value == spender {
					attrAmountIdx := i + 1
					if attrAmountIdx < len(evt.Attributes) {
						attrNext := evt.Attributes[attrAmountIdx]
						if attrNext.Key == "amount" {
							commaSeperatedCoins := attrNext.Value
							currentCoins := strings.Split(commaSeperatedCoins, ",")
							for _, coin := range currentCoins {
								if coin != "" {
									parsedCoin, err := sdk.ParseCoinNormalized(coin)
									if err != nil {
										return sdk.Coins{}, err
									}

									coinsSpent = append(coinsSpent, parsedCoin)
								}
							}
						}
					}
				}
			}
		}
	}

	return coinsSpent, nil
}
