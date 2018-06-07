package simpleGovernance

import (
	"encoding/binary"
	"reflect"

	// stake "github.com/cosmos/cosmos-sdk/examples/simpleGov/x/simplestake"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

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

// NewBeginBlocker checks if the
// func NewBeginBlocker(k Keeper) (sdk.BeginBlocker, sdk.Error) {
// 	return func(ctx sdk.Context, req abci.RequestBeginBlock) (abci.ResponseBeginBlock, sdk.Error) {
// 		err := checkProposal(ctx, k)
// 		if err != nil {
// 			return abci.ResponseBeginBlock{}, err
// 		}
// 		return abci.ResponseBeginBlock{}, nil
// 	}
// }

// func checkProposal() {
//
// }
// 	proposalID = ProposalProcessingQueue.Peek()
// 	if (proposalID == nil)
// 		return
//
// 	proposal = load(Proposals, proposalID)
//
// 	if (proposal.Votes.YesVotes/proposal.InitTotalVotingPower > 2/3)
//
// 		// proposal accepted early by super-majority
// 		// no punishments; refund deposits
//
// 		ProposalProcessingQueue.pop()
//
// 		var newDeposits []Deposits
//
// 		// XXX: why do we need to reset deposits? cant we just clear it ?
// 		for each (amount, depositer) in proposal.Deposits
// 			newDeposits.append[{0, depositer}]
// 			depositer.AtomBalance += amount
//
// 		proposal.Deposits = newDeposits
// 		store(Proposals, proposalID, proposal)
//
// 		checkProposal()
//
// 	else if (CurrentBlock == proposal.VotingStartBlock + proposal.Procedure.VotingPeriod)
//
// 		ProposalProcessingQueue.pop()
// 		activeProcedure = load(params, 'ActiveProcedure')
//
// 		for each validator in CurrentBondedValidators
// 			validatorGovInfo = load(ValidatorGovInfos, <proposalID | validator.Address>)
//
// 			if (validatorGovInfo.InitVotingPower != nil)
// 				// validator was bonded when vote started
//
// 				validatorOption = load(Options, <proposalID | validator.Address>)
// 				if (validatorOption == nil)
// 					// validator did not vote
// 					slash validator by activeProcedure.GovernancePenalty
//
//
// 		totalNonAbstain = proposal.Votes.YesVotes + proposal.Votes.NoVotes + proposal.Votes.NoWithVetoVotes
// 		if( proposal.Votes.YesVotes/totalNonAbstain > 0.5 AND proposal.Votes.NoWithVetoVotes/totalNonAbstain  < 1/3)
//
// 			//  proposal was accepted at the end of the voting period
// 			//  refund deposits (non-voters already punished)
//
// 			var newDeposits []Deposits
//
// 			for each (amount, depositer) in proposal.Deposits
// 				newDeposits.append[{0, depositer}]
// 				depositer.AtomBalance += amount
//
// 			proposal.Deposits = newDeposits
// 			store(Proposals, proposalID, proposal)
//
// 			checkProposal()

func checkProposal(ctx sdk.Context, k Keeper) sdk.Error {
	proposal, err := k.ProposalQueueHead(ctx)
	if err != nil {
		return err
	}

	// Proposal reached the end of the voting period
	if ctx.BlockHeight() == proposal.SubmitBlock+proposal.BlockLimit {
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
		// return checkProposal() // XXX where's this function defined ? why no params ?
	}
	return nil
}

const minDeposit = 100 // How do you set the min deposit

func handleSubmitProposalMsg(ctx sdk.Context, k Keeper, msg SubmitProposalMsg) sdk.Result {
	_, _, err := k.ck.SubtractCoins(ctx, msg.Submitter, msg.Deposit)
	if err != nil {
		return err.Result() // Code and Log of the error
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
	}

	return sdk.Result{} // return proper result
}

// TODO func proposal IsOpen()

func handleVoteMsg(ctx sdk.Context, k Keeper, msg VoteMsg) sdk.Result {
	proposal, err := k.GetProposal(ctx, msg.ProposalID)
	if err != nil {
		return err.Result()
	}

	if ctx.BlockHeight() > proposal.SubmitBlock+proposal.BlockLimit {
		return ErrVotingPeriodClosed().Result()
	}

	delegatedTo := k.sm.GetDelegations(ctx, msg.Voter, 10)

	if len(delegatedTo) == 0 {
		return stake.ErrNoDelegatorForAddress(DefaultCodespace).Result()
	}

	var key []byte

	proposalIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(proposalIDBytes, uint64(msg.ProposalID))

	key = append(proposalIDBytes, msg.Voter.Bytes()...)
	voterOption, err := k.GetOption(ctx, key)

	if err != nil {
		return err.Result()
	}
	// TODO check if this line is OK
	// nil option return error in ValidateBasic
	if voterOption == "" {
		// voter has not voted yet
		for _, delegation := range delegatedTo {
			bondShares := delegation.GetBondShares().Denom()
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

	k.SetOption(ctx, key, msg.Option)
	k.SetProposal(ctx, msg.ProposalID, proposal)

	return sdk.Result{} // return proper result

}
