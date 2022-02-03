package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var _ types.MsgServer = Keeper{}

func (k Keeper) SoftwareUpgrade(goCtx context.Context, req *types.MsgSoftwareUpgrade) (*types.MsgSoftwareUpgradeResponse, error) {
	govAcct := k.authKeeper.GetModuleAddress(gov.ModuleName).String()
	if govAcct != req.Authority {
		return nil, sdkerrors.Wrapf(types.ErrInvalidSigner, "expected %s got %s", govAcct, msg.Authority)
	}
}
