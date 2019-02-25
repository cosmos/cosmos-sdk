package cli

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func getShares(sharesAmountStr string, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sharesAmount sdk.Dec, err error) {
	sharesAmount, err = sdk.NewDecFromStr(sharesAmountStr)
	if err != nil {
		return sharesAmount, err
	}

	if !sharesAmount.GT(sdk.ZeroDec()) {
		return sharesAmount, errors.New("shares amount must be positive number (ex. 123, 1.23456789)")
	}

	return
}

func buildCommissionMsg(rateStr, maxRateStr, maxChangeRateStr string) (commission types.CommissionMsg, err error) {
	if rateStr == "" || maxRateStr == "" || maxChangeRateStr == "" {
		return commission, errors.New("must specify all validator commission parameters")
	}

	rate, err := sdk.NewDecFromStr(rateStr)
	if err != nil {
		return commission, err
	}

	maxRate, err := sdk.NewDecFromStr(maxRateStr)
	if err != nil {
		return commission, err
	}

	maxChangeRate, err := sdk.NewDecFromStr(maxChangeRateStr)
	if err != nil {
		return commission, err
	}

	commission = types.NewCommissionMsg(rate, maxRate, maxChangeRate)
	return commission, nil
}
