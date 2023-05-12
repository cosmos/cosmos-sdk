package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/errors"
)

var _ banktypes.SendRestrictionFn = Keeper{}.SendRestrictionFn

func (k Keeper) SendRestrictionFn(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, _ sdk.Coins) (sdk.AccAddress, error) {
	if !sanction.HasBypass(ctx) && k.IsSanctionedAddr(ctx, fromAddr) {
		return nil, errors.ErrSanctionedAccount.Wrapf("cannot send from %s", fromAddr.String())
	}
	return toAddr, nil
}
