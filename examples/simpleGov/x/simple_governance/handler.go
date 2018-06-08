package simpleGovernance

import (
	"encoding/binary"
	"reflect"

	abci "github.com/tendermint/abci/types"
	// stake "github.com/cosmos/cosmos-sdk/examples/simpleGov/x/simplestake"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// Minimum proposal deposit
const minDeposit = 100

func int64ToBytes(i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}

// NewHandler creates a new handler for all simple_gov type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case SubmitProposalMsg:
			return handleSubmitProposalMsg(ctx, k, msg)
		case VoteMsg:
			return handleVoteMsg(ctx, k, msg)
		default:
			errMsg := "Unrecognized gov Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// NewBeginBlocker checks proposal and creates a BeginBlock
func NewBeginBlocker(k Keeper) sdk.BeginBlocker {
	// TODO cannot use func literal (type func("github.com/cosmos/cosmos-sdk/types".Context, "github.com/tendermint/abci/types".RequestBeginBlock) "github.com/tendermint/abci/types".ResponseBeginBlock) as type "github.com/cosmos/cosmos-sdk/types".BeginBlocker in return argument
	return func(ctx sdk.Context, req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
		err := checkProposal(ctx, k)
		if err != nil {
			panic(err)
		}
		return
	}
}

func checkProposal(ctx sdk.Context, k Keeper) sdk.Error {
	proposal, err := k.ProposalQueueHead(ctx)
	if err != nil {
		return err
	}

	// Proposal reached the end of the voting period
	if ctx.BlockHeight() >= proposal.SubmitBlock+proposal.BlockLimit &&
		proposal.IsOpen() {
		k.ProposalQueuePop(ctx)

		nonAbstainTotal := proposal.YesVotes + proposal.NoVotes
		if float64(proposal.YesVotes)/float64(nonAbstainTotal) > float64(0.5) { // TODO: Deal with decimals

			// Refund deposit
			_, _, err := k.ck.AddCoins(ctx, proposal.Submitter, proposal.Deposit)
			if err != nil {
				return err
			}
			proposal.State = "Accepted"
		} else {
			proposal.State = "Rejected"
		}
		return checkProposal(ctx, k)
	}
	return nil
}

func handleSubmitProposalMsg(ctx sdk.Context, k Keeper, msg SubmitProposalMsg) sdk.Result {
	// TODO check if have to set coins
	_, _, err := k.ck.SubtractCoins(ctx, msg.Submitter, msg.Deposit)
	if err != nil {
		return err.Result()
	}

	if msg.Deposit.AmountOf("Atom") >= minDeposit {
		proposal := NewProposal(
			msg.Title,
			msg.Description,
			msg.Submitter,
			ctx.BlockHeight(),
			msg.VotingWindow,
			msg.Deposit)
		proposalID := k.NewProposalID(ctx)
		k.SetProposal(ctx, proposalID, proposal)
		return sdk.Result{
			Tags: sdk.NewTags(
				"action", []byte("propose"),
				"proposal", int64ToBytes(proposalID),
				"submitter", msg.Submitter.Bytes(),
			),
		}
	}
	return ErrMinimumDeposit().Result()
}

func handleVoteMsg(ctx sdk.Context, k Keeper, msg VoteMsg) sdk.Result {
	proposal, err := k.GetProposal(ctx, msg.ProposalID)
	if err != nil {
		return err.Result()
	}

	if ctx.BlockHeight() > proposal.SubmitBlock+proposal.BlockLimit ||
		!proposal.IsOpen() {
		return ErrVotingPeriodClosed().Result()
	}

	delegatedTo := k.sm.GetDelegations(ctx, msg.Voter, 10)

	if len(delegatedTo) == 0 {
		return stake.ErrNoDelegatorForAddress(DefaultCodespace).Result()
	}
	// Check if address already voted
	voterOption, err := k.GetOption(ctx, msg.ProposalID, msg.Voter)
	if voterOption == "" && err != nil {
		// voter has not voted yet
		for _, delegation := range delegatedTo {
			bondShares := delegation.GetBondShares().Evaluate()
			err = proposal.updateTally(msg.Option, bondShares)
			if err != nil {
				return err.Result()
			}
		}
	} else {
		// voter has already voted
		for _, delegation := range delegatedTo {
			bondShares := delegation.GetBondShares().Evaluate()
			// update previous vote with new one
			err = proposal.updateTally(voterOption, -bondShares)
			if err != nil {
				return err.Result()
			}
			err = proposal.updateTally(msg.Option, bondShares)
			if err != nil {
				return err.Result()
			}
		}
	}

	k.SetOption(ctx, msg.ProposalID, msg.Voter, msg.Option)
	k.SetProposal(ctx, msg.ProposalID, proposal)

	return sdk.Result{
		Tags: sdk.NewTags(
			"action", []byte("vote"),
			"proposal", int64ToBytes(msg.ProposalID),
			"voter", msg.Voter.Bytes(),
			"option", []byte(msg.Option),
		),
	}

}
