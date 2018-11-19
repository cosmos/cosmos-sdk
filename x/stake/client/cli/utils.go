package cli

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/pkg/errors"
)

func getShares(
	storeName string, cdc *codec.Codec, sharesAmountStr,
	sharesPercentStr string, delAddr sdk.AccAddress, valAddr sdk.ValAddress,
) (sharesAmount sdk.Dec, err error) {

	switch {
	case sharesAmountStr != "" && sharesPercentStr != "":
		return sharesAmount, errors.Errorf("can either specify the amount OR the percent of the shares, not both")

	case sharesAmountStr == "" && sharesPercentStr == "":
		return sharesAmount, errors.Errorf("can either specify the amount OR the percent of the shares, not both")

	case sharesAmountStr != "":
		sharesAmount, err = sdk.NewDecFromStr(sharesAmountStr)
		if err != nil {
			return sharesAmount, err
		}
		if !sharesAmount.GT(sdk.ZeroDec()) {
			return sharesAmount, errors.Errorf("shares amount must be positive number (ex. 123, 1.23456789)")
		}

	case sharesPercentStr != "":
		var sharesPercent sdk.Dec
		sharesPercent, err = sdk.NewDecFromStr(sharesPercentStr)
		if err != nil {
			return sharesAmount, err
		}
		if !sharesPercent.GT(sdk.ZeroDec()) || !sharesPercent.LTE(sdk.OneDec()) {
			return sharesAmount, errors.Errorf("shares percent must be >0 and <=1 (ex. 0.01, 0.75, 1)")
		}

		// make a query to get the existing delegation shares
		key := stake.GetDelegationKey(delAddr, valAddr)
		cliCtx := context.NewCLIContext().
			WithCodec(cdc).
			WithAccountDecoder(cdc)

		resQuery, err := cliCtx.QueryStore(key, storeName)
		if err != nil {
			return sharesAmount, errors.Errorf("cannot find delegation to determine percent Error: %v", err)
		}

		delegation, err := types.UnmarshalDelegation(cdc, key, resQuery)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		sharesAmount = sharesPercent.Mul(delegation.Shares)
	}

	return
}

func buildCommissionMsg(rateStr, maxRateStr, maxChangeRateStr string) (commission types.CommissionMsg, err error) {
	if rateStr == "" || maxRateStr == "" || maxChangeRateStr == "" {
		return commission, errors.Errorf("must specify all validator commission parameters")
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
