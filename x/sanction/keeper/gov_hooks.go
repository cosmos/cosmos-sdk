package keeper

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

var _ govtypes.GovHooks = Keeper{}

// AfterProposalSubmission is called after proposal is submitted.
// If there's enough deposit, temporary entries are created.
func (k Keeper) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	k.proposalGovHook(ctx, proposalID)
}

// AfterProposalDeposit is called after a deposit is made.
// If there's enough deposit, temporary entries are created.
func (k Keeper) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, _ sdk.AccAddress) {
	k.proposalGovHook(ctx, proposalID)
}

// AfterProposalVote is called after a vote on a proposal is cast. This one does nothing.
func (k Keeper) AfterProposalVote(_ sdk.Context, _ uint64, _ sdk.AccAddress) {}

// AfterProposalFailedMinDeposit is called when proposal fails to reach min deposit.
// Cleans up any possible temporary entries.
func (k Keeper) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {
	k.proposalGovHook(ctx, proposalID)
}

// AfterProposalVotingPeriodEnded is called when proposal's finishes it's voting period.
// Cleans up temporary entries.
func (k Keeper) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	k.proposalGovHook(ctx, proposalID)
}

const (
	// propStatusNotFound is a governance module ProposalStatus (an enum) used in here to indicate that a proposal wasn't found.
	propStatusNotFound = govv1.ProposalStatus(-100)
)

// proposalGovHook does what needs to be done in here with the proposal in question.
// What needs to be done always depends on the status of the proposal.
// So while some hooks are probably only called when a proposal has a certain status, it's safer to just always do this.
func (k Keeper) proposalGovHook(ctx sdk.Context, proposalID uint64) {
	// A proposal can sometimes be deleted. In such cases, we still need to do some stuff.
	propStatus := propStatusNotFound
	proposal, found := k.govKeeper.GetProposal(ctx, proposalID)
	if found {
		propStatus = proposal.Status
	}

	switch propStatus {
	case govv1.StatusDepositPeriod, govv1.StatusVotingPeriod:
		for _, msg := range proposal.Messages {
			if k.isModuleGovHooksMsgURL(msg.TypeUrl) {
				// If the deposit is over the (non-zero) minimum, add temporary entries for the addrs.
				makeTemps := false
				minDeposit := k.getImmediateMinDeposit(ctx, msg)
				if !minDeposit.IsZero() {
					deposit := sdk.Coins(proposal.TotalDeposit)
					_, hasNeg := deposit.SafeSub(minDeposit...)
					if !hasNeg {
						makeTemps = true
					}
				}
				if makeTemps {
					addrs := k.getMsgAddresses(msg)
					var err error
					switch msg.TypeUrl {
					case k.msgSanctionTypeURL:
						err = k.AddTemporarySanction(ctx, proposalID, addrs...)
					case k.msgUnsanctionTypeURL:
						err = k.AddTemporaryUnsanction(ctx, proposalID, addrs...)
					}
					if err != nil {
						panic(err)
					}
				}
			}
		}
	case govv1.StatusRejected, govv1.StatusFailed, propStatusNotFound:
		// Delete only the temporary entries that were associated with this proposal.
		// We do this for all proposals, regardless of whether they have a sanction or unsanction message.
		// A) When the proposal isn't found, there's no way we can know what messages it had.
		// B) This code is simpler than trying to not call DeleteGovPropTempEntries when not needed.
		// C) The extra processing from calling DeleteGovPropTempEntries is probably on par with what's needed
		//    to not always call it.
		// D) There's no risk of this deleting anything that shouldn't be deleted.
		k.DeleteGovPropTempEntries(ctx, proposalID)
	case govv1.StatusPassed:
		// Nothing to do. The processing of the proposal message does everything that's needed.
	default:
		panic(fmt.Errorf("invalid governance proposal status: [%s]", proposal.Status))
	}
}

// isModuleGovHooksMsgURL returns true if the provided URL is one that these gov hooks care about.
func (k Keeper) isModuleGovHooksMsgURL(url string) bool {
	// Note: We don't need to care about one of ours being wrapped in a MsgExecLegacyContent,
	// because ours don't implement the old v1beta1.Content interface, and thus can't be wrapped as such.
	return url == k.msgSanctionTypeURL || url == k.msgUnsanctionTypeURL
}

// getMsgAddresses gets the list of addresses from the provided message if it's one we care about.
// If it's a type we don't care about, returns nil.
func (k Keeper) getMsgAddresses(msg *codectypes.Any) []sdk.AccAddress {
	if msg == nil {
		return nil
	}
	switch msg.TypeUrl {
	case k.msgSanctionTypeURL:
		var msgSanction *sanction.MsgSanction
		if err := k.cdc.UnpackAny(msg, &msgSanction); err != nil {
			panic(err)
		}
		addrs, err := toAccAddrs(msgSanction.Addresses)
		if err != nil {
			panic(err)
		}
		return addrs
	case k.msgUnsanctionTypeURL:
		var msgUnsanction *sanction.MsgUnsanction
		if err := k.cdc.UnpackAny(msg, &msgUnsanction); err != nil {
			panic(err)
		}
		addrs, err := toAccAddrs(msgUnsanction.Addresses)
		if err != nil {
			panic(err)
		}
		return addrs
	}
	return nil
}

// getImmediateMinDeposit gets the minimum deposit for immediate action to be taken on the proposal msg.
// If the msg isn't of a type we care about, returns empty coins.
func (k Keeper) getImmediateMinDeposit(ctx sdk.Context, msg *codectypes.Any) sdk.Coins {
	if msg != nil {
		switch msg.TypeUrl {
		case k.msgSanctionTypeURL:
			return k.GetImmediateSanctionMinDeposit(ctx)
		case k.msgUnsanctionTypeURL:
			return k.GetImmediateUnsanctionMinDeposit(ctx)
		}
	}
	return sdk.Coins{}
}
