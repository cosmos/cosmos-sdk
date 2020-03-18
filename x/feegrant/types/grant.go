package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewFeeAllowanceGrantBase(granter sdk.AccAddress, grantee sdk.AccAddress) FeeAllowanceGrantBase {
	return FeeAllowanceGrantBase{
		Granter: granter,
		Grantee: grantee,
	}
}
