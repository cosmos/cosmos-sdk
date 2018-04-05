package gov

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Handle all "gov" type messages.
func NewHandler(gm governanceMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case DepositMsg:
			return handleDepositMsg(ctx, gm, msg)
		case SubmitProposalMsg:
			return handleSubmitProposalMsg(ctx, gm, msg)
		case VoteMsg:
			return handleVoteMsg(ctx, gm, msg)
		default:
			errMsg := "Unrecognized gov Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle SubmitProposalMsg.
func handleSubmitProposalMsg(ctx sdk.Context, gm governanceMapper, msg SubmitProposalMsg) sdk.Result {

	_, err := gm.ck.SubtractCoins(ctx, msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	if !gm.GetActiveProcedure().validProposalType(msg.ProposalType) {
		return ErrInvalidProposalType(msg.ProposalType).Result()
	}

	if ctx.IsCheckTx() {
		return sdk.Result{} // TODO
	}

	initialDeposit := Deposit{
		Depositer: msg.Proposer,
		Amount:    msg.InitialDeposit,
	}

	proposal := Proposal{
		ProposalID:       gm.getNewProposalID(ctx),
		Title:            msg.Title,
		Description:      msg.Description,
		ProposalType:     msg.ProposalType,
		TotalDeposit:     initialDeposit.Amount,
		Deposits:         []Deposit{initialDeposit},
		SubmitBlock:      ctx.BlockHeight(),
		VotingStartBlock: -1, // TODO: Make Time
		TotalVotingPower: 0,
		Procedure:        *(gm.GetActiveProcedure()), // TODO: Get cloned active Procedure from params kvstore
		YesVotes:         0,
		NoVotes:          0,
		NoWithVetoVotes:  0,
		AbstainVotes:     0,
	}

	if proposal.TotalDeposit.IsGTE(proposal.Procedure.MinDeposit) {
		activateVotingPeriod(ctx, gm, &proposal)
	}

	gm.SetProposal(ctx, proposal)

	return sdk.Result{} // TODO
}

// Handle DepositMsg.
func handleDepositMsg(ctx sdk.Context, gm governanceMapper, msg DepositMsg) sdk.Result {

	_, err := gm.ck.SubtractCoins(ctx, msg.Depositer, msg.Amount)
	if err != nil {
		return err.Result()
	}

	proposal := gm.GetProposal(ctx, msg.ProposalID)

	if proposal == nil {
		return ErrUnknownProposal(msg.ProposalID).Result()
	}

	if proposal.isActive() {
		return ErrAlreadyActiveProposal(msg.ProposalID).Result()
	}

	if ctx.IsCheckTx() {
		return sdk.Result{} // TODO
	}

	deposit := Deposit{
		Depositer: msg.Depositer,
		Amount:    msg.Amount,
	}

	proposal.TotalDeposit = proposal.TotalDeposit.Plus(deposit.Amount)
	proposal.Deposits = append(proposal.Deposits, deposit)

	if proposal.TotalDeposit.IsGTE(proposal.Procedure.MinDeposit) {
		activateVotingPeriod(ctx, gm, proposal)
	}

	gm.SetProposal(ctx, *proposal)

	return sdk.Result{} // TODO
}

// Handle SendMsg.
func handleVoteMsg(ctx sdk.Context, gm governanceMapper, msg VoteMsg) sdk.Result {

	proposal := gm.GetProposal(ctx, msg.ProposalID)
	if proposal == nil {
		return ErrUnknownProposal(msg.ProposalID).Result()
	}

	if !proposal.isActive() || ctx.BlockHeight() > proposal.VotingStartBlock+proposal.Procedure.VotingPeriod {
		return ErrInactiveProposal(msg.ProposalID).Result()
	}

	validatorGovInfo := proposal.getValidatorGovInfo(msg.Voter)

	// Need to finalize interface to staking mapper for delegatedTo. Makes assumption from here on out.
	delegatedTo := gm.sm.LoadDelegatorCandidates(ctx, msg.Voter) // TODO: Finalize with staking store

	if validatorGovInfo == nil && len(delegatedTo) == 0 {
		return ErrAddressNotStaked(msg.Voter).Result() // TODO: Return proper Error
	}

	if proposal.VotingStartBlock <= gm.sm.getLastDelationChangeBlock(msg.Voter) { // TODO: Get last block in which voter bonded or unbonded
		return ErrAddressChangedDelegation(msg.Voter).Result() // TODO: Return proper Error
	}

	if ctx.IsCheckTx() {
		return sdk.Result{} // TODO
	}

	existingVote := proposal.getVote(msg.Voter)

	if existingVote == nil {
		proposal.VoteList = append(proposal.VoteList, Vote{Voter: msg.Voter, ProposalID: msg.ProposalID, Option: msg.Option})

		if validatorGovInfo != nil {
			voteWeight := validatorGovInfo.InitVotingPower - validatorGovInfo.Minus
			proposal.updateTally(msg.Option, voteWeight)
			validatorGovInfo.LastVoteWeight = voteWeight
		}

		for index, delegation := range delegatedTo {
			proposal.updateTally(msg.Option, delegation.amount)
			delegatedValidatorGovInfo := proposal.getValidatorGovInfo(delegation.validator)
			delegatedValidatorGovInfo.Minus += delegation.amount

			delegatedValidatorVote := proposal.getVote(delegation.validator)

			if delegatedValidatorVote != nil {
				proposal.updateTally(delegatedValidatorVote.Option, -delegation.amount)
			}
		}

	} else {
		if validatorGovInfo != nil {
			proposal.updateTally(existingVote.Option, -(validatorGovInfo.LastVoteWeight))
			voteWeight := validatorGovInfo.InitVotingPower - validatorGovInfo.Minus
			proposal.updateTally(msg.Option, voteWeight)
			validatorGovInfo.LastVoteWeight = voteWeight
		}

		for index, delegation := range delegatedTo {
			proposal.updateTally(existingVote.Option, -delegation.amount)
			proposal.updateTally(msg.Option, delegation.amount)
		}

		existingVote.Option = msg.Option
	}

	gm.SetProposal(ctx, *proposal)

	return sdk.Result{} // TODO
}

func activateVotingPeriod(ctx sdk.Context, gm governanceMapper, proposal *Proposal) {
	proposal.VotingStartBlock = ctx.BlockHeight()

	// TODO: Can we get this directly from stakeState
	// stakeState := gm.sm.loadGlobalState()
	// proposal.TotalVotingPower = stakeState.TotalSupply

	proposal.TotalVotingPower = 0

	validatorList := gm.sm.GetValidators(100) // TODO: Finalize with staking module

	for index, validator := range validatorList {
		validatorGovInfo := ValidatorGovInfo{
			ProposalID:      proposal.ProposalID,
			ValidatorAddr:   validator.Address,     // TODO: Finalize with staking module
			InitVotingPower: validator.VotingPower, // TODO: Finalize with staking module
			Minus:           0,
			LastVoteWeight:  -1,
		}

		proposal.ValidatorGovInfos = append(proposal.ValidatorGovInfos, validatorGovInfo)
	}

	gm.ProposalQueuePush(ctx, *proposal)
}
