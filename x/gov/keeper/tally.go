package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// Tally iterates over the votes and updates the tally of a proposal based on the voting power of the
// voters
// CONTRACT: passes is always false when err!=nil
func (keeper Keeper) Tally(ctx context.Context, proposal v1.Proposal) (passes bool, burnDeposits bool, participation math.LegacyDec, tallyResults v1.TallyResult, err error) {
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

	/* DISABLED on AtomOne SDK - no possible increase of computation speed by
	 iterating over validators since vote inheritance is disabled.
	 Keeping as comment because this should be adapted with governors loop

	// we check first if voting power of validators alone is enough to pass quorum
	// and if so, we return true skipping the iteration over all votes
	// can speed up computation in case quorum is already reached by validator votes alone
	approxTotalVotingPower := math.LegacyZeroDec()
	for _, val := range currValidators {
		_, ok := keeper.GetVote(ctx, proposal.Id, sdk.AccAddress(val.GetOperator()))
		if ok {
			approxTotalVotingPower = approxTotalVotingPower.Add(math.LegacyNewDecFromInt(val.GetBondedTokens()))
		}
	}
	// check and return whether or not the proposal has reached quorum
	approxPercentVoting := approxTotalVotingPower.Quo(math.LegacyNewDecFromInt(totalBonded))
	if approxPercentVoting.GTE(quorum) {
		return true, nil
	}
	*/

	// voting power of validators does not reach quorum, let's tally all votes
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
	err := keeper.sk.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
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
	totalVotingPower = math.LegacyZeroDec()
	if isFinal {
		results = make(map[v1.VoteOption]math.LegacyDec)
		results[v1.OptionYes] = math.LegacyZeroDec()
		results[v1.OptionAbstain] = math.LegacyZeroDec()
		results[v1.OptionNo] = math.LegacyZeroDec()
	}

	rng := collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposal.Id)
	err = keeper.Votes.Walk(ctx, rng, func(_ collections.Pair[uint64, sdk.AccAddress], vote v1.Vote) (bool, error) {
		// if validator, just record it in the map
		voter, err := keeper.authKeeper.AddressCodec().StringToBytes(vote.Voter)
		if err != nil {
			return false, err
		}

		// iterate over all delegations from voter, deduct from any delegated-to validators
		err = keeper.sk.IterateDelegations(ctx, voter, func(index int64, delegation stakingtypes.DelegationI) (stop bool) {
			valAddrStr := delegation.GetValidatorAddr()

			if val, ok := currValidators[valAddrStr]; ok {
				// delegation shares * bonded / total shares
				votingPower := delegation.GetShares().MulInt(val.GetBondedTokens()).Quo(val.GetDelegatorShares())

				if isFinal {
					for _, option := range vote.Options {
						weight, _ := math.LegacyNewDecFromStr(option.Weight)
						subPower := votingPower.Mul(weight)
						results[option.Option] = results[option.Option].Add(subPower)
					}
				}
				totalVotingPower = totalVotingPower.Add(votingPower)
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

	/* DISABLED on AtomOne SDK - Voting can only be done with your own stake
	// iterate over the validators again to tally their voting power
	for _, val := range currValidators {
		if len(val.Vote) == 0 {
			continue
		}

		sharesAfterDeductions := val.DelegatorShares.Sub(val.DelegatorDeductions)
		votingPower := sharesAfterDeductions.MulInt(val.BondedTokens).Quo(val.DelegatorShares)

		for _, option := range val.Vote {
			weight, _ := math.LegacyNewDecFromStr(option.Weight)
			subPower := votingPower.Mul(weight)
			results[option.Option] = results[option.Option].Add(subPower)
		}
		totalVotingPower = totalVotingPower.Add(votingPower)
	}
	*/

	return totalVotingPower, results, nil
}

// getQuorumAndThreshold returns the appropriate quorum and threshold according
// to proposal kind. If the proposal contains multiple kinds, the highest
// quorum and threshold is returned.
func (keeper Keeper) getQuorumAndThreshold(ctx context.Context, proposal v1.Proposal) (quorum math.LegacyDec, threshold math.LegacyDec) {
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
