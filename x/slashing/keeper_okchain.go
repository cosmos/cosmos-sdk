package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (k Keeper) modifyValidatorStatus(ctx sdk.Context, consAddress sdk.ConsAddress, status types.ValStatus) {
	signingInfo, found := k.getValidatorSigningInfo(ctx, consAddress)
	if found {
		//update validator status to Created
		signingInfo.ValidatorStatus = status
		k.SetValidatorSigningInfo(ctx, consAddress, signingInfo)
	}
}
