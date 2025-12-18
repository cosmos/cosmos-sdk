package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Tally iterates over the votes and updates the tally of a proposal based on the voting power of the
// voters
// CONTRACT: passes is always false when err!=nil
func (keeper Keeper) Tally(ctx context.Context, proposal v1.Proposal) (passes, burnDeposits bool, participation math.LegacyDec, tallyResults v1.TallyResult, err error) {
	// fetch all the bonded validators
	currValidators, err := keeper.getBondedValidatorsByAddress(ctx)
	if err != nil {
		return false, false, math.LegacyZeroDec(), tallyResults, err
	}
	totalVotingPower, results, err := keeper.tallyVotes(ctx, proposal, currValidators, true)
	if err != nil {
		return false, false, math.LegacyZeroDec(), tallyResults, err
	}

	params, err := keeper.Params.Get(ctx)
	if err != nil {
		return false, false, math.LegacyZeroDec(), tallyResults, fmt.Errorf("failed to get params: %w", err)
	}

	tallyResults = v1.NewTallyResultFromMap(results)

	// If there is no staked coins, the proposal fails
	totalBonded, err := keeper.sk.TotalBondedTokens(ctx)
	if err != nil {
		return false, false, math.LegacyZeroDec(), tallyResults, err
	}

	if totalBonded.IsZero() {
		return false, false, math.LegacyZeroDec(), tallyResults, nil
	}

	// If there is not enough quorum of votes, the proposal fails
	percentVoting := totalVotingPower.Quo(math.LegacyNewDecFromInt(totalBonded))
	quorum, threshold := keeper.getQuorumAndThreshold(ctx, proposal)
	if percentVoting.LT(quorum) {
		return false, params.BurnVoteQuorum, percentVoting, tallyResults, nil
	}

	// Compute non-abstaining voting power, aka active voting power
	activeVotingPower := totalVotingPower.Sub(results[v1.OptionAbstain])

	// If no one votes (everyone abstains), proposal fails
	if activeVotingPower.IsZero() {
		return false, false, percentVoting, tallyResults, nil
	}

	// If more than `threshold` of non-abstaining voters vote Yes, proposal passes.
	yesPercent := results[v1.OptionYes].Quo(activeVotingPower)
	if yesPercent.GT(threshold) {
		return true, false, percentVoting, tallyResults, nil
	}

	// If more than `burnDepositNoThreshold` of non-abstaining voters vote No,
	// proposal is rejected and deposit is burned.
	burnDepositNoThreshold := math.LegacyMustNewDecFromStr(params.BurnDepositNoThreshold)
	noPercent := results[v1.OptionNo].Quo(activeVotingPower)
	if noPercent.GT(burnDepositNoThreshold) {
		return false, true, percentVoting, tallyResults, nil
	}

	// If less than `burnDepositNoThreshold` of non-abstaining voters vote No,
	// proposal is rejected but deposit is not burned.
	return false, false, percentVoting, tallyResults, nil
}

// HasReachedQuorum returns whether or not a proposal has reached quorum
// this is just a stripped down version of the Tally function above
func (keeper Keeper) HasReachedQuorum(ctx sdk.Context, proposal v1.Proposal) (bool, error) {
	// If there is no staked coins, the proposal has not reached quorum
	totalBonded, err := keeper.sk.TotalBondedTokens(ctx)
	if err != nil {
		return false, err
	}

	if totalBonded.IsZero() {
		return false, nil
	}

	currValidators, err := keeper.getBondedValidatorsByAddress(ctx)
	if err != nil {
		return false, err
	}
	totalVotingPower, _, err := keeper.tallyVotes(ctx, proposal, currValidators, false)
	if err != nil {
		return false, err
	}

	// check and return whether or not the proposal has reached quorum
	percentVoting := totalVotingPower.Quo(math.LegacyNewDecFromInt(totalBonded))
	quorum, _ := keeper.getQuorumAndThreshold(ctx, proposal)
	return percentVoting.GTE(quorum), nil
}

// getBondedValidatorsByAddress fetches all the bonded validators and return
// them in map using their operator address as the key.
func (keeper Keeper) getBondedValidatorsByAddress(ctx context.Context) (map[string]stakingtypes.ValidatorI, error) {
	vals := make(map[string]stakingtypes.ValidatorI)
	err := keeper.sk.IterateBondedValidatorsByPower(ctx, func(_ int64, validator stakingtypes.ValidatorI) (stop bool) {
		vals[validator.GetOperator()] = validator
		return false
	})
	return vals, err
}

// tallyVotes returns the total voting power and tally results of the votes
// on a proposal. If `isFinal` is true, results will be stored in `results`
// map and votes will be deleted. Otherwise, only the total voting power
// will be returned and `results` will be nil.
func (keeper Keeper) tallyVotes(
	ctx context.Context, proposal v1.Proposal,
	currValidators map[string]stakingtypes.ValidatorI, isFinal bool,
) (totalVotingPower math.LegacyDec, results map[v1.VoteOption]math.LegacyDec, err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	totalVotingPower = math.LegacyZeroDec()
	// keeps track of governors that voted or have delegators that voted
	allGovernors := make(map[string]v1.GovernorGovInfo)

	if isFinal {
		results = make(map[v1.VoteOption]math.LegacyDec)
		results[v1.OptionYes] = math.LegacyZeroDec()
		results[v1.OptionAbstain] = math.LegacyZeroDec()
		results[v1.OptionNo] = math.LegacyZeroDec()
	}

	rng := collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposal.Id)
	err = keeper.Votes.Walk(ctx, rng, func(_ collections.Pair[uint64, sdk.AccAddress], vote v1.Vote) (bool, error) {
		var governor v1.GovernorGovInfo
		voter, err := keeper.authKeeper.AddressCodec().StringToBytes(vote.Voter)
		if err != nil {
			return false, err
		}

		gd, err := keeper.GovernanceDelegations.Get(sdkCtx, voter)
		hasGovernor := err == nil
		if hasGovernor {
			if gi, ok := allGovernors[gd.GovernorAddress]; ok {
				governor = gi
			} else {
				govAddr := types.MustGovernorAddressFromBech32(gd.GovernorAddress)

				var shares []v1.GovernorValShares
				err := keeper.ValidatorSharesByGovernor.Walk(ctx, collections.NewPrefixedPairRange[types.GovernorAddress, sdk.ValAddress](govAddr), func(_ collections.Pair[types.GovernorAddress, sdk.ValAddress], s v1.GovernorValShares) (stop bool, err error) {
					shares = append(shares, s)
					return false, nil
				})
				if err != nil {
					return false, err
				}
				governor = v1.NewGovernorGovInfo(
					govAddr,
					shares,
					v1.WeightedVoteOptions{},
				)
			}
			if gd.GovernorAddress == types.GovernorAddress(voter).String() {
				// voter and governor are the same account, record his vote
				governor.Vote = vote.Options
			}
			// Ensure allGovernors contains the updated governor
			allGovernors[gd.GovernorAddress] = governor
		}

		// iterate over all delegations from voter, deduct from any delegated-to validators
		err = keeper.sk.IterateDelegations(ctx, voter, func(_ int64, delegation stakingtypes.DelegationI) (stop bool) {
			valAddrStr := delegation.GetValidatorAddr()
			votingPower := math.LegacyZeroDec()

			if val, ok := currValidators[valAddrStr]; ok {
				// delegation shares * bonded / total shares
				votingPower = votingPower.Add(delegation.GetShares().MulInt(val.GetBondedTokens()).Quo(val.GetDelegatorShares()))

				// remove the delegation shares from the governor
				if hasGovernor {
					governor.ValSharesDeductions[valAddrStr] = governor.ValSharesDeductions[valAddrStr].Add(delegation.GetShares())
				}
			}

			totalVotingPower = totalVotingPower.Add(votingPower)
			if isFinal {
				for _, option := range vote.Options {
					subPower := option.Power(votingPower)
					results[option.Option] = results[option.Option].Add(subPower)
				}
			}

			return false
		})
		if err != nil {
			return true, err
		}

		if isFinal {
			if err := keeper.Votes.Remove(ctx, collections.Join(vote.ProposalId, sdk.AccAddress(voter))); err != nil {
				return false, err
			}
		}

		return false, nil
	})
	if err != nil {
		return totalVotingPower, results, err
	}

	// get only the voting governors that are active and have the niminum self-delegation requirement met.
	currGovernors := keeper.getCurrGovernors(sdkCtx, allGovernors)

	// iterate over the governors again to tally their voting power
	// As active governor are simply voters that need to have 100% of their bonded tokens
	// delegated to them and their shares were deducted when iterating over votes
	// we don't need to handle special cases.
	for _, gov := range currGovernors {
		votingPower := getGovernorVotingPower(gov, currValidators)
		if isFinal {
			for _, option := range gov.Vote {
				subPower := option.Power(votingPower)
				results[option.Option] = results[option.Option].Add(subPower)
			}
		}
		totalVotingPower = totalVotingPower.Add(votingPower)
	}

	return totalVotingPower, results, nil
}

// getQuorumAndThreshold returns the appropriate quorum and threshold according
// to proposal kind. If the proposal contains multiple kinds, the highest
// quorum and threshold is returned.
func (keeper Keeper) getQuorumAndThreshold(ctx context.Context, proposal v1.Proposal) (quorum, threshold math.LegacyDec) {
	params, err := keeper.Params.Get(ctx)
	if err != nil {
		return math.LegacyZeroDec(), math.LegacyZeroDec()
	}

	kinds := keeper.ProposalKinds(proposal)

	// start with the default quorum and threshold
	quorum = keeper.GetQuorum(ctx)
	threshold = math.LegacyMustNewDecFromStr(params.Threshold)

	// Check for Constitution Amendment and update if higher
	if kinds.HasKindConstitutionAmendment() {
		constitutionQuorum := keeper.GetConstitutionAmendmentQuorum(ctx)
		constitutionThreshold := math.LegacyMustNewDecFromStr(params.ConstitutionAmendmentThreshold)

		if constitutionQuorum.GT(quorum) {
			quorum = constitutionQuorum
		}
		if constitutionThreshold.GT(threshold) {
			threshold = constitutionThreshold
		}
	}

	// Check for Law and update if higher
	if kinds.HasKindLaw() {
		lawQuorum := keeper.GetLawQuorum(ctx)
		lawThreshold := math.LegacyMustNewDecFromStr(params.LawThreshold)
		if proposal.Endorsed {
			// If the proposal is endorsed, we use threshold for the generic kind
			lawThreshold = math.LegacyMustNewDecFromStr(params.Threshold)
		}

		if lawQuorum.GT(quorum) {
			quorum = lawQuorum
		}
		if lawThreshold.GT(threshold) {
			threshold = lawThreshold
		}
	}

	return quorum, threshold
}

// getCurrGovernors returns the governors that voted, are active and meet the minimum self-delegation requirement
func (keeper Keeper) getCurrGovernors(ctx sdk.Context, allGovernors map[string]v1.GovernorGovInfo) (governors []v1.GovernorGovInfo) {
	governorsInfos := make([]v1.GovernorGovInfo, 0)
	for _, govInfo := range allGovernors {
		governor, _ := keeper.Governors.Get(ctx, govInfo.Address)

		if keeper.ValidateGovernorMinSelfDelegation(ctx, governor) && len(govInfo.Vote) > 0 {
			governorsInfos = append(governorsInfos, govInfo)
		}
	}

	return governorsInfos
}

func getGovernorVotingPower(governor v1.GovernorGovInfo, currValidators map[string]stakingtypes.ValidatorI) (votingPower math.LegacyDec) {
	votingPower = math.LegacyZeroDec()
	for valAddrStr, shares := range governor.ValShares {
		if val, ok := currValidators[valAddrStr]; ok {
			sharesAfterDeductions := shares.Sub(governor.ValSharesDeductions[valAddrStr])
			votingPower = votingPower.Add(sharesAfterDeductions.MulInt(val.GetBondedTokens()).Quo(val.GetDelegatorShares()))
		}
	}
	return votingPower
}
