package v6

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type paramsKeeper interface {
	GetParams(context.Context) (types.Params, error)
	SetParams(context.Context, types.Params) error
}

// Migrate adds the KeyRotationFee param. Existing chains inherit the default
// key rotation fee amount denominated in their configured bond denom when the
// field is absent.
func Migrate(ctx context.Context, keeper paramsKeeper) error {
	params, err := keeper.GetParams(ctx)
	if err != nil {
		return err
	}

	if isMissingKeyRotationFee(params.KeyRotationFee) {
		params.KeyRotationFee = sdk.Coin{
			Denom:  params.BondDenom,
			Amount: types.DefaultKeyRotationFee.Amount,
		}
	}
	if err := params.Validate(); err != nil {
		return err
	}

	return keeper.SetParams(ctx, params)
}

func isMissingKeyRotationFee(fee sdk.Coin) bool {
	return fee.Denom == "" && fee.Amount.IsNil()
}
