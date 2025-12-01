package v1

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetNewMinDeposit(minDepositFloor, lastMinDeposit sdk.Coins, percChange math.LegacyDec) sdk.Coins {
	newMinDeposit := sdk.Coins{}
	minDepositFloorDenomsSeen := make(map[string]bool)
	for _, lastMinDepositCoin := range lastMinDeposit {
		minDepositFloorCoinAmt := minDepositFloor.AmountOf(lastMinDepositCoin.Denom)
		if minDepositFloorCoinAmt.IsZero() {
			// minDepositFloor was changed since last update,
			// and this coin was removed.
			// reflect this also in the current min initial deposit,
			// i.e. remove this coin
			continue
		}
		minDepositFloorDenomsSeen[lastMinDepositCoin.Denom] = true
		minDepositCoinAmt := lastMinDepositCoin.Amount.ToLegacyDec().Mul(percChange).TruncateInt()
		if minDepositCoinAmt.LT(minDepositFloorCoinAmt) {
			newMinDeposit = append(newMinDeposit, sdk.NewCoin(lastMinDepositCoin.Denom, minDepositFloorCoinAmt))
		} else {
			newMinDeposit = append(newMinDeposit, sdk.NewCoin(lastMinDepositCoin.Denom, minDepositCoinAmt))
		}
	}

	// make sure any new denoms in minDepositFloor are added to minDeposit
	for _, minDepositFloorCoin := range minDepositFloor {
		if _, seen := minDepositFloorDenomsSeen[minDepositFloorCoin.Denom]; !seen {
			minDepositCoinAmt := minDepositFloorCoin.Amount.ToLegacyDec().Mul(percChange).TruncateInt()
			if minDepositCoinAmt.LT(minDepositFloorCoin.Amount) {
				newMinDeposit = append(newMinDeposit, minDepositFloorCoin)
			} else {
				newMinDeposit = append(newMinDeposit, sdk.NewCoin(minDepositFloorCoin.Denom, minDepositCoinAmt))
			}
		}
	}

	return newMinDeposit
}
