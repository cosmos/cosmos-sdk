package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

var _ banktypes.SendRestrictionFn = Keeper{}.SendRestrictionFn

func (k Keeper) SendRestrictionFn(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.AccAddress, error) {
	// bypass if the context says to.
	if quarantine.HasBypass(ctx) {
		return toAddr, nil
	}
	// bypass if the fromAddr is either the toAddr or the funds holder.
	fundsHolder := k.GetFundsHolder()
	if fromAddr.Equals(toAddr) || fromAddr.Equals(fundsHolder) {
		return toAddr, nil
	}
	// Nothing to do if they're not quarantined or if they are, but have auto-accept enabled for the fromAddr.
	if !k.IsQuarantinedAddr(ctx, toAddr) || k.IsAutoAccept(ctx, toAddr, fromAddr) {
		return toAddr, nil
	}
	// Make sure there's a funds holder defined since we need it now.
	// This should not be possible since NewKeeper makes sure it always has a value.
	// But it would be really bad if it somehow happened.
	if fundsHolder.Empty() {
		return nil, sdkerrors.ErrUnknownAddress.Wrapf("no quarantine funds holder account defined")
	}
	// Record the quarantined funds and return the funds holder as the new toAddr.
	err := k.AddQuarantinedCoins(ctx, amt, toAddr, fromAddr)
	if err != nil {
		return nil, err
	}
	return fundsHolder, nil
}
