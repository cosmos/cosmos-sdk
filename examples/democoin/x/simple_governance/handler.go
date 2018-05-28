package module_tutorial

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stake "github.coms/cosmos/cosmos-sdk/x/stake"
	"reflect"
)

// Handle all "simple_gov" type messages.
func NewHandler(sgm SimpleGovernanceKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case SubmitProposalMsg:
			return handleSubmitProposalMsg(ctx, sgm, msg)
		case VoteMsg:
			return handleVoteMsg(ctx, sgm, msg)
		default:
			errMsg := "Unrecognized gov Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func NewBeginBlocker(sgm SimpleGovernanceKeeper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		checkProposal(ctx, sgm)
		return abci.ResponseBeginBlock{}
	}
}

func checkProposal(ctx sdk.Context, sgm simpleGovernanceKeeper) {
	proposal := sgm.ProposalQueuePeek(ctx)
	if proposal == nil {
		return
	}

	// Proposal reached the end of the voting period
	if ctx.BlockHeight() == proposal.SubmitBlock+1209600 {
		sgm.ProposalQueuePop(ctx)

		nonAbstainTotal := proposal.Votes.YesVotes + proposal.Votes.NoVotes
		if proposal.YesVotes/nonAbstainTotal > 0.5 { // TODO: Deal with decimals

			// Refund deposit
			_, err := sgm.ck.AddCoins(ctx, proposal.Submitter, proposal.Deposit.AmountOf("Atom"))
			if err != nil {
				panic("Should not be possible")
			}

			proposal.State = "Accepted"

		} else {

			proposal.State = "Rejected"

		}

		return checkProposal()
	}
}

func handleSubmitProposalMsg(ctx sdk.Context, sgm SimpleGovernanceKeeper, msg sdk.Msg) sdk.Result {
	_, err := sgm.ck.SubstractCoins(ctx, msg.Submitter, msg.Deposit)

	if err != nil {
		return err.Result()
	}

	var minDeposit int64 = 100

	if msg.Deposit.AmountOf("Atom") >= minDeposit {
		proposal := Proposal{
			Title:        msg.Title,
			Description:  msg.Description,
			Submitter:    msg.Submitter,
			SubmitBlock:  ctx.BlockHeight(),
			State:        "Open",
			Deposit:      msg.Deposit,
			YesVotes:     0,
			NoVotes:      0,
			AbstainVotes: 0,
		}

		sgm.SetProposal(ctx, sgm.NewProposalID, proposal)
	}

	return sdk.Result{} // return proper result
}

func handleVoteMsg(ctx sdk.Context, sgm SimpleGovernanceKeeper, msg sdk.Msg) sdk.Result {
	proposal := sgm.GetProposal(ctx, msg.ProposalID)

	if proposal == nil {
		return ErrInvalidProposalID().Result()
	}

	if ctx.BlockHeight() > proposal.SubmitBlock+1209600 {
		return ErrVotingPeriodClosed().Result()
	}

	delegatedTo := sgm.sm.getValidators(msg.Voter)
	if len(delegatedTo) <= 0 {
		return stake.ErrNoDelegatorForAddress().Result()
	}

	key := append(msg.ProposalID, msg.Voter...)
	voterOption := sgm.GetOption(ctx, key)
	if voterOption == nil {
		// voter has not voted yet

		for _, delegation := range delegatedTo {
			proposal.updateTally(msg.Option, delegation.amount)
		}
	} else {
		// voter has already voted

		for _, delegation := range delegatedTo {
			proposal.updateTally(voterOption, -delegation.amount)
			proposal.updateTally(msg.Option, delegation.amount)
		}
	}

	sgm.SetOption(ctx, key, msg.Option)
	sgm.SetProposal(ctx, msg.ProposalID, proposal)

	return sdk.Result{} // return proper result

}
