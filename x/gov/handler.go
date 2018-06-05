package gov

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
		return sdk.Result{}
	}

	initDeposit := Deposit{
		Depositer: msg.Proposer,
		Amount:    msg.InitialDeposit,
	}

	keeper.NewProposal(ctx, msg.Title, msg.Description, msg.ProposalType, initDeposit)

	initialDeposit := Deposit{
		Depositer: msg.Proposer,
		Amount:    msg.InitialDeposit,
	}

	proposal := &Proposal{
		ProposalID:       keeper.getNewProposalID(ctx),
		Title:            msg.Title,
		Description:      msg.Description,
		ProposalType:     msg.ProposalType,
		TotalDeposit:     initialDeposit.Amount,
		Deposits:         []Deposit{initialDeposit},
		SubmitBlock:      ctx.BlockHeight(),
		VotingStartBlock: -1, // TODO: Make Time
	}

	if proposal.TotalDeposit.IsGTE(keeper.GetDepositProcedure().MinDeposit) {
		keeper.activateVotingPeriod(ctx, proposal)
	}

	keeper.SetProposal(ctx, proposal)

	tags := sdk.NewTags("action", []byte("submitProposal"), "proposer", msg.Proposer.Bytes(), "proposalId", []byte{byte(proposal.ProposalID)})

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

	if proposal.TotalDeposit.IsGTE(keeper.GetDepositProcedure().MinDeposit) {
		keeper.activateVotingPeriod(ctx, proposal)
	}

	keeper.SetProposal(ctx, proposal)

	tags := sdk.NewTags("action", []byte("deposit"), "depositer", msg.Depositer.Bytes(), "proposalId", []byte{byte(proposal.ProposalID)})
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

	if !proposal.isActive() || ctx.BlockHeight() > proposal.VotingStartBlock+keeper.GetVotingProcedure().VotingPeriod {
		return ErrInactiveProposal(msg.ProposalID).Result()
	}

	if ctx.IsCheckTx() {
		return sdk.Result{} // TODO
	}

	vote := Vote{
		ProposalID: proposal.ProposalID,
		Voter:      msg.Voter,
		Option:     msg.Option,
	}

	keeper.setVote(ctx, proposal.ProposalID, msg.Voter, vote)

	tags := sdk.NewTags("action", []byte("vote"), "voter", msg.Voter.Bytes(), "proposalId", []byte{byte(proposal.ProposalID)})
	return sdk.Result{
		Tags: tags,
	}
}
