package gov

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// InitGenesis - store genesis parameters
func InitGenesis(ctx sdk.Context, ak types.AccountKeeper, bk types.BankKeeper, k *keeper.Keeper, data *v1.GenesisState) {
	err := k.ProposalID.Set(ctx, data.StartingProposalId)
	if err != nil {
		panic(err)
	}

	err = k.Params.Set(ctx, *data.Params)
	if err != nil {
		panic(err)
	}

	err = k.Constitution.Set(ctx, data.Constitution)
	if err != nil {
		panic(err)
	}

	// Use default values for participation EMAs if not provided
	participationEmaStr := data.ParticipationEma
	if participationEmaStr == "" {
		participationEmaStr = v1.DefaultParticipationEma
	}
	participationEma, err := math.LegacyNewDecFromStr(participationEmaStr)
	if err != nil {
		panic(fmt.Sprintf("invalid value for participationEma %s: %v", participationEmaStr, err))
	}
	if err := k.ParticipationEMA.Set(ctx, participationEma); err != nil {
		panic(err)
	}

	constitutionAmendmentParticipationEmaStr := data.ConstitutionAmendmentParticipationEma
	if constitutionAmendmentParticipationEmaStr == "" {
		constitutionAmendmentParticipationEmaStr = v1.DefaultParticipationEma
	}
	constitutionAmendmentparticipationEma, err := math.LegacyNewDecFromStr(constitutionAmendmentParticipationEmaStr)
	if err != nil {
		panic(fmt.Sprintf("invalid value for constitutionAmendmentparticipationEma %s: %v", constitutionAmendmentParticipationEmaStr, err))
	}
	if err := k.ConstitutionAmendmentParticipationEMA.Set(ctx, constitutionAmendmentparticipationEma); err != nil {
		panic(err)
	}

	lawParticipationEmaStr := data.LawParticipationEma
	if lawParticipationEmaStr == "" {
		lawParticipationEmaStr = v1.DefaultParticipationEma
	}
	lawParticipationEma, err := math.LegacyNewDecFromStr(lawParticipationEmaStr)
	if err != nil {
		panic(fmt.Sprintf("invalid value for lawParticipationEma %s: %v", lawParticipationEmaStr, err))
	}
	if err := k.LawParticipationEMA.Set(ctx, lawParticipationEma); err != nil {
		panic(err)
	}

	// check if the deposits pool account exists
	moduleAcc := k.GetGovernanceAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	var totalDeposits sdk.Coins
	for _, deposit := range data.Deposits {
		err := k.SetDeposit(ctx, *deposit)
		if err != nil {
			panic(err)
		}
		totalDeposits = totalDeposits.Add(deposit.Amount...)
	}

	for _, vote := range data.Votes {
		addr, err := ak.AddressCodec().StringToBytes(vote.Voter)
		if err != nil {
			panic(err)
		}
		err = k.Votes.Set(ctx, collections.Join(vote.ProposalId, sdk.AccAddress(addr)), *vote)
		if err != nil {
			panic(err)
		}
	}

	for _, proposal := range data.Proposals {
		switch proposal.Status {
		case v1.StatusDepositPeriod:
			err := k.InactiveProposalsQueue.Set(ctx, collections.Join(*proposal.DepositEndTime, proposal.Id), proposal.Id)
			if err != nil {
				panic(err)
			}
		case v1.StatusVotingPeriod:
			err := k.ActiveProposalsQueue.Set(ctx, collections.Join(*proposal.VotingEndTime, proposal.Id), proposal.Id)
			if err != nil {
				panic(err)
			}
		}
		err := k.SetProposal(ctx, *proposal)
		if err != nil {
			panic(err)
		}

		if data.Params.QuorumCheckCount > 0 && proposal.Status == v1.StatusVotingPeriod {
			quorumTimeoutTime := proposal.VotingStartTime.Add(*data.Params.QuorumTimeout)
			quorumCheckEntry := v1.NewQuorumCheckQueueEntry(quorumTimeoutTime, data.Params.QuorumCheckCount)
			quorum := false
			if ctx.BlockTime().After(quorumTimeoutTime) {
				quorum, err = k.HasReachedQuorum(ctx, *proposal)
				if err != nil {
					panic(fmt.Sprintf("HasReachedQuorum returned an error: %v", err))
				}
				if !quorum {
					// since we don't export the state of the quorum check queue, we can't know how many checks were actually
					// done. However, in order to trigger a vote time extension, it is enough to have QuorumChecksDone > 0 to
					// trigger a vote time extension, so we set it to 1
					quorumCheckEntry.QuorumChecksDone = 1
				}
			}
			if !quorum {
				if err := k.QuorumCheckQueue.Set(ctx, collections.Join(quorumTimeoutTime, proposal.Id), quorumCheckEntry); err != nil {
					panic(fmt.Sprintf("QuorumCheckQueue.Set returned an error: %v", err))
				}
			}
		}
	}

	// if account has zero balance it probably means it's not set, so we set it
	balance := bk.GetAllBalances(ctx, moduleAcc.GetAddress())
	if balance.IsZero() {
		ak.SetModuleAccount(ctx, moduleAcc)
	}

	// check if total deposits equals balance, if it doesn't panic because there were export/import errors
	if !balance.Equal(totalDeposits) {
		panic(fmt.Sprintf("expected module account was %s but we got %s", balance.String(), totalDeposits.String()))
	}

	t := ctx.BlockTime()
	if data.LastMinDeposit != nil {
		if err := k.LastMinDeposit.Set(ctx, *data.LastMinDeposit); err != nil {
			panic(err)
		}
	} else {
		if err := k.LastMinDeposit.Set(ctx, v1.LastMinDeposit{
			Value: data.Params.MinDepositThrottler.FloorValue,
			Time:  &t,
		}); err != nil {
			panic(err)
		}
	}

	if data.LastMinInitialDeposit != nil {
		if err := k.LastMinInitialDeposit.Set(ctx, *data.LastMinInitialDeposit); err != nil {
			panic(err)
		}
	} else {
		if err := k.LastMinInitialDeposit.Set(ctx, v1.LastMinDeposit{
			Value: data.Params.MinInitialDepositThrottler.FloorValue,
			Time:  &t,
		}); err != nil {
			panic(err)
		}
	}

	// set governors
	for _, governor := range data.Governors {
		// check that base account exists
		accAddr := sdk.AccAddress(governor.GetAddress())
		acc := ak.GetAccount(ctx, accAddr)
		if acc == nil {
			panic(fmt.Sprintf("account %s does not exist", accAddr.String()))
		}

		k.Governors.Set(ctx, governor.GetAddress(), *governor)
		if governor.IsActive() {
			err := k.DelegateToGovernor(ctx, accAddr, governor.GetAddress())
			if err != nil {
				panic(fmt.Sprintf("failed to delegate to governor %s: %v", governor.GetAddress().String(), err))
			}
		}
	}
	// set governance delegations
	for _, delegation := range data.GovernanceDelegations {
		delAddr := sdk.MustAccAddressFromBech32(delegation.DelegatorAddress)
		govAddr := types.MustGovernorAddressFromBech32(delegation.GovernorAddress)
		// check delegator exists
		acc := ak.GetAccount(ctx, delAddr)
		if acc == nil {
			panic(fmt.Sprintf("account %s does not exist", delAddr.String()))
		}
		// check governor exists
		_, err := k.Governors.Get(ctx, govAddr)
		if err != nil {
			panic(fmt.Sprintf("error getting governor %s: %v", govAddr.String(), err))
		}

		// if account is active governor and delegation is not to self, error
		delGovAddr := types.GovernorAddress(delAddr)
		if _, err = k.Governors.Get(ctx, delGovAddr); err != nil && !delGovAddr.Equals(govAddr) {
			panic(fmt.Sprintf("account %s is an active governor and cannot delegate", delAddr.String()))
		}

		err = k.DelegateToGovernor(ctx, delAddr, govAddr)
		if err != nil {
			panic(fmt.Sprintf("failed to delegate to governor %s: %v", govAddr.String(), err))
		}
	}
}

// ExportGenesis - output genesis parameters
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) (*v1.GenesisState, error) {
	startingProposalID, err := k.ProposalID.Peek(ctx)
	if err != nil {
		return nil, err
	}

	var proposals v1.Proposals
	err = k.Proposals.Walk(ctx, nil, func(_ uint64, value v1.Proposal) (stop bool, err error) {
		proposals = append(proposals, &value)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	constitution, err := k.Constitution.Get(ctx)
	if err != nil {
		return nil, err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	var proposalsDeposits v1.Deposits
	err = k.Deposits.Walk(ctx, nil, func(_ collections.Pair[uint64, sdk.AccAddress], value v1.Deposit) (stop bool, err error) {
		proposalsDeposits = append(proposalsDeposits, &value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	// export proposals votes
	var proposalsVotes v1.Votes
	err = k.Votes.Walk(ctx, nil, func(_ collections.Pair[uint64, sdk.AccAddress], value v1.Vote) (stop bool, err error) {
		proposalsVotes = append(proposalsVotes, &value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	blockTime := ctx.BlockTime()
	lastMinDeposit := v1.LastMinDeposit{
		Value: k.GetMinDeposit(ctx),
		Time:  &blockTime,
	}

	lastMinInitialDeposit := v1.LastMinDeposit{
		Value: k.GetMinInitialDeposit(ctx),
		Time:  &blockTime,
	}

	participationEma, err := k.ParticipationEMA.Get(ctx)
	if err != nil {
		panic(err)
	}

	constitutionAmendmentParticipationEma, err := k.ConstitutionAmendmentParticipationEMA.Get(ctx)
	if err != nil {
		panic(err)
	}

	lawParticipationEma, err := k.LawParticipationEMA.Get(ctx)
	if err != nil {
		panic(err)
	}

	var governors []*v1.Governor
	err = k.Governors.Walk(ctx, nil, func(_ types.GovernorAddress, value v1.Governor) (stop bool, err error) {
		governors = append(governors, &value)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	var governanceDelegations []*v1.GovernanceDelegation
	for _, g := range governors {
		var delegations []*v1.GovernanceDelegation
		k.GovernanceDelegationsByGovernor.Walk(ctx, collections.NewPrefixedPairRange[types.GovernorAddress, sdk.AccAddress](g.GetAddress()), func(_ collections.Pair[types.GovernorAddress, sdk.AccAddress], delegation v1.GovernanceDelegation) (stop bool, err error) {
			delegations = append(delegations, &delegation)
			return false, nil
		})
		governanceDelegations = append(governanceDelegations, delegations...)
	}

	return &v1.GenesisState{
		StartingProposalId:                    startingProposalID,
		Deposits:                              proposalsDeposits,
		Votes:                                 proposalsVotes,
		Proposals:                             proposals,
		Params:                                &params,
		Constitution:                          constitution,
		LastMinDeposit:                        &lastMinDeposit,
		LastMinInitialDeposit:                 &lastMinInitialDeposit,
		ParticipationEma:                      participationEma.String(),
		ConstitutionAmendmentParticipationEma: constitutionAmendmentParticipationEma.String(),
		LawParticipationEma:                   lawParticipationEma.String(),
		Governors:                             governors,
		GovernanceDelegations:                 governanceDelegations,
	}, nil
}
