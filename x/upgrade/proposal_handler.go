package upgrade

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	RouterKey = ModuleName
)

func NewSoftwareUpgradeProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) sdk.Error {
		switch c := content.(type) {
		case SoftwareUpgradeProposal:
			return handleSoftwareUpgradeProposal(ctx, k, c)
		case CancelSoftwareUpgradeProposal:
			return handleCancelSoftwareUpgradeProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized software upgrade proposal content type: %T", c)
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}

func handleSoftwareUpgradeProposal(ctx sdk.Context, k Keeper, p SoftwareUpgradeProposal) sdk.Error {
	return k.ScheduleUpgrade(ctx, p.Plan)
}

func handleCancelSoftwareUpgradeProposal(ctx sdk.Context, k Keeper, p CancelSoftwareUpgradeProposal) sdk.Error {
	k.ClearUpgradePlan(ctx)
	return nil
}
