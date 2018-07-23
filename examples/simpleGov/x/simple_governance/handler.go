package simpleGovernance

import (
	"encoding/binary"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/tendermint/abci/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Minimum proposal deposit
var minDeposit = sdk.NewInt(int64(100))

const votingPeriod = 1209600

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

// NewEndBlocker checks proposals and generates a EndBlocker
func NewEndBlocker(k Keeper) sdk.EndBlocker {
	return sdk.EndBlocker(func(ctx sdk.Context, req types.RequestEndBlock) types.ResponseEndBlock {
		newTags := sdk.NewTags()
		tags, err := checkProposal(ctx, k, newTags)
		if err != nil {
			panic(err)
		}
		return abci.ResponseEndBlock{
			Tags: tags,
		}
	})
}

// checkProposal checks if the proposal reached the end of the voting period
// and handles the logic of closing it
func checkProposal(ctx sdk.Context, k Keeper, tags sdk.Tags) (sdk.Tags, sdk.Error) {
	proposal, err := k.ProposalQueueHead(ctx)
	if err != nil {
		return nil, err
	}
	// Proposal reached the end of the voting period
	if ctx.BlockHeight() >= proposal.SubmitBlock+votingPeriod && proposal.IsOpen() {
		k.ProposalQueuePop(ctx)
		proposalIDBytes := k.cdc.MustMarshalBinaryBare(proposal.ID)
		nonAbstainTotal := proposal.YesVotes + proposal.NoVotes
		if float64(proposal.YesVotes)/float64(nonAbstainTotal) > float64(0.5) { // TODO: Deal with decimals

			// Refund deposit
			_, _, err := k.ck.AddCoins(ctx, proposal.Submitter, proposal.Deposit)
			if err != nil {
				return nil, err
			}
			proposal.State = "Accepted"
			tags.AppendTag("action", []byte("proposalPassed"))
			tags.AppendTag("proposalId", proposalIDBytes)
		} else {
			proposal.State = "Rejected"
			tags.AppendTag("action", []byte("proposalRejected"))
			tags.AppendTag("proposalId", proposalIDBytes)
		}
		return checkProposal(ctx, k, tags)
	}
	return tags, nil
}

// handleVoteMsg handles the logic of a SubmitProposalMsg
func handleSubmitProposalMsg(ctx sdk.Context, k Keeper, msg SubmitProposalMsg) sdk.Result {
	err := msg.ValidateBasic()
	if err != nil {
		return err.Result()
	}

	// Subtract coins from the submitter balance and updates it
	_, _, err = k.ck.SubtractCoins(ctx, msg.Submitter, msg.Deposit)
	if err != nil {
		return err.Result()
	}

	if msg.Deposit.AmountOf("Atom").GT(minDeposit) ||
		msg.Deposit.AmountOf("Atom").Equal(minDeposit) {
		proposalID := k.NewProposalID(ctx)
		proposal := NewProposal(
			proposalID,
			msg.Title,
			msg.Description,
			msg.Submitter,
			ctx.BlockHeight(),
			msg.Deposit)
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

// handleVoteMsg handles the logic of a VoteMsg
func handleVoteMsg(ctx sdk.Context, k Keeper, msg VoteMsg) sdk.Result {
	err := msg.ValidateBasic()
	if err != nil {
		return err.Result()
	}

	proposal, err := k.GetProposal(ctx, msg.ProposalID)
	if err != nil {
		return err.Result()
	}

	if ctx.BlockHeight() > proposal.SubmitBlock+votingPeriod ||
		!proposal.IsOpen() {
		return ErrVotingPeriodClosed().Result()
	}

	delegatedTo := k.sm.GetDelegations(ctx, msg.Voter, 10)

	if len(delegatedTo) <= 0 {
		return stake.ErrNoDelegatorForAddress(stake.DefaultCodespace).Result()
	}
	// Check if address already voted
	voterOption, err := k.GetVote(ctx, msg.ProposalID, msg.Voter)
	if voterOption == "" && err != nil {
		// voter has not voted yet
		for _, delegation := range delegatedTo {
			bondShares := delegation.GetBondShares().EvaluateBig().Int64()
			err = proposal.updateTally(msg.Option, bondShares)
			if err != nil {
				return err.Result()
			}
		}
	} else {
		// voter has already voted
		for _, delegation := range delegatedTo {
			bondShares := delegation.GetBondShares().EvaluateBig().Int64()
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

	k.SetVote(ctx, msg.ProposalID, msg.Voter, msg.Option)
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
