package keeper

import (
	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/errors"
)

var _ banktypes.SendRestrictionFn = Keeper{}.SendRestrictionFn

func (k Keeper) SendRestrictionFn(ctx types.Context, fromAddr, toAddr types.AccAddress, _ types.Coins) (types.AccAddress, error) {
	if !sanction.HasSanctionBypass(ctx) && k.IsSanctionedAddr(ctx, fromAddr) {
		return nil, errors.ErrSanctionedAccount.Wrap(fromAddr.String())
	}
	return toAddr, nil
}
