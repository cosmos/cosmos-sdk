package delegation

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgDelegateFeeAllowance:
			grant := FeeAllowanceGrant{Granter: msg.Granter, Grantee: msg.Grantee, Allowance: msg.Allowance}
			k.DelegateFeeAllowance(ctx, grant)
			return sdk.Result{}
		case MsgRevokeFeeAllowance:
			k.RevokeFeeAllowance(ctx, msg.Granter, msg.Grantee)
			return sdk.Result{}
		default:
			errMsg := fmt.Sprintf("Unrecognized data Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
