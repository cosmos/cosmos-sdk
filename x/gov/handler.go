package gov

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bech32cosmos "github.com/cosmos/bech32cosmos/go"
	"strconv"
)

// Handle all "gov" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgDeposit:
			return handleMsgDeposit(ctx, keeper, msg)
		case MsgSubmitProposal:
			return handleMsgSubmitProposal(ctx, keeper, msg)
		case MsgVote:
			return handleMsgVote(ctx, keeper, msg)
		default:
			errMsg := "Unrecognized gov Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgSubmitProposal.
func handleMsgSubmitProposal(ctx sdk.Context, keeper Keeper, msg MsgSubmitProposal) sdk.Result {

	_, _, err := keeper.ck.SubtractCoins(ctx, msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	if ctx.IsCheckTx() {
		if !keeper.GetActiveProcedure().validProposalType(msg.ProposalType) {
			return ErrInvalidProposalType(msg.ProposalType).Result()
		}

		return sdk.Result{}
	}

	initDeposit := Deposit{
		Depositer: msg.Proposer,
		Amount:    msg.InitialDeposit,
	}

	keeper.NewProposal(ctx, msg.Title, msg.Description, msg.ProposalType, initDeposit)

	if !keeper.GetActiveProcedure().validProposalType(msg.ProposalType) {
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
		ProposalID:       keeper.getNewProposalID(ctx),
		Title:            msg.Title,
		Description:      msg.Description,
		ProposalType:     msg.ProposalType,
		TotalDeposit:     initialDeposit.Amount,
		Deposits:         []Deposit{initialDeposit},
		SubmitBlock:      ctx.BlockHeight(),
		VotingStartBlock: -1, // TODO: Make Time
		TotalVotingPower: 0,
		Procedure:        *(keeper.GetActiveProcedure()), // TODO: Get cloned active Procedure from params kvstore
		YesVotes:         0,
		NoVotes:          0,
		NoWithVetoVotes:  0,
		AbstainVotes:     0,
	}

	if proposal.TotalDeposit.IsGTE(proposal.Procedure.MinDeposit) {
		keeper.activateVotingPeriod(ctx, &proposal)
	}

	keeper.SetProposal(ctx, proposal)

	bech, _ := bech32cosmos.Encode("",[]byte(strconv.FormatInt(proposal.ProposalID,10)))

	tags := sdk.NewTags("action", []byte("submitProposal"), "proposer", msg.Proposer.Bytes(), "proposalId", []byte(bech)) //TODO bech32 use https://github.com/tendermint/tmlibs/tree/develop/bech32

	return sdk.Result{
		Data: []byte{byte(proposal.ProposalID)},
		Tags: tags,
	}
}

// Handle MsgDeposit.
func handleMsgDeposit(ctx sdk.Context, keeper Keeper, msg MsgDeposit) sdk.Result {

	_, _, err := keeper.ck.SubtractCoins(ctx, msg.Depositer, msg.Amount)
	if err != nil {
		return err.Result()
	}

	proposal := keeper.GetProposal(ctx, msg.ProposalID)

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
		keeper.activateVotingPeriod(ctx, proposal)
	}

	keeper.SetProposal(ctx, *proposal)

	bech, _ := bech32cosmos.Encode("",[]byte(strconv.FormatInt(proposal.ProposalID,10))) //TODO bech32 use https://github.com/tendermint/tmlibs/tree/develop/bech32

	tags := sdk.NewTags("action", []byte("deposit"), "depositer", msg.Depositer.Bytes(), "proposalId", []byte(bech))
	return sdk.Result{
		Tags: tags,
	}
}

// Handle SendMsg.
func handleMsgVote(ctx sdk.Context, keeper Keeper, msg MsgVote) sdk.Result {

	proposal := keeper.GetProposal(ctx, msg.ProposalID)
	if proposal == nil {
		return ErrUnknownProposal(msg.ProposalID).Result()
	}

	if !proposal.isActive() || ctx.BlockHeight() > proposal.VotingStartBlock+proposal.Procedure.VotingPeriod {
		return ErrInactiveProposal(msg.ProposalID).Result()
	}

	validatorGovInfo := proposal.getValidatorGovInfo(msg.Voter)

	// Need to finalize interface to staking mapper for delegatedTo. Makes assumption from here on out.
	delegatedTo := keeper.sm.LoadDelegatorCandidates(ctx, msg.Voter) // TODO: Finalize with staking store

	if validatorGovInfo == nil && len(delegatedTo) == 0 {
		return ErrAddressNotStaked(msg.Voter).Result() // TODO: Return proper Error
	}

	if proposal.VotingStartBlock <= keeper.sm.getLastDelationChangeBlock(msg.Voter) { // TODO: Get last block in which voter bonded or unbonded
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

	keeper.SetProposal(ctx, *proposal)

	bech, _ := bech32cosmos.Encode("",[]byte(strconv.FormatInt(proposal.ProposalID,10))) //TODO bech32 use https://github.com/tendermint/tmlibs/tree/develop/bech32

	tags := sdk.NewTags("action", []byte("vote"), "voter", msg.Voter.Bytes(), "proposalId", []byte(bech))
	return sdk.Result{
		Tags: tags,
	}
}
