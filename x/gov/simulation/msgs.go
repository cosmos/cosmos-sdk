package simulation

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

const (
	denom = "steak"
)

// SimulateSubmittingVotingAndSlashingForProposal simulates creating a msg Submit Proposal
// voting on the proposal, and subsequently slashing the proposal. It is implemented using
// future operations.
// TODO: Vote more intelligently, so we can actually do some checks regarding votes passing or failing
// TODO: Actually check that validator slashings happened
func SimulateSubmittingVotingAndSlashingForProposal(k gov.Keeper, sk stake.Keeper) simulation.Operation {
	handler := gov.NewHandler(k)
	// The states are:
	// column 1: All validators vote
	// column 2: 90% vote
	// column 3: 75% vote
	// column 4: 40% vote
	// column 5: 15% vote
	// column 6: noone votes
	// All columns sum to 100 for simplicity, values chosen by @valardragon semi-arbitrarily,
	// feel free to change.
	numVotesTransitionMatrix, _ := simulation.CreateTransitionMatrix([][]int{
		{20, 10, 0, 0, 0, 0},
		{55, 50, 20, 10, 0, 0},
		{25, 25, 30, 25, 30, 15},
		{0, 15, 30, 25, 30, 30},
		{0, 0, 20, 30, 30, 30},
		{0, 0, 0, 10, 10, 25},
	})
	statePercentageArray := []float64{1, .9, .75, .4, .15, 0}
	curNumVotesState := 1
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, event func(string)) (action string, fOps []simulation.FutureOperation, err error) {
		// 1) submit proposal now
		sender := simulation.RandomKey(r, keys)
		msg, err := simulationCreateMsgSubmitProposal(r, sender)
		if err != nil {
			return "", nil, err
		}
		action, ok := simulateHandleMsgSubmitProposal(msg, sk, handler, ctx, event)
		// don't schedule votes if proposal failed
		if !ok {
			return action, nil, nil
		}
		proposalID := k.GetLastProposalID(ctx)
		// 2) Schedule operations for votes
		// 2.1) first pick a number of people to vote.
		curNumVotesState = numVotesTransitionMatrix.NextState(r, curNumVotesState)
		numVotes := int(math.Ceil(float64(len(keys)) * statePercentageArray[curNumVotesState]))
		// 2.2) select who votes and when
		whoVotes := r.Perm(len(keys))
		// didntVote := whoVotes[numVotes:]
		whoVotes = whoVotes[:numVotes]
		votingPeriod := k.GetVotingProcedure(ctx).VotingPeriod
		fops := make([]simulation.FutureOperation, numVotes+1)
		for i := 0; i < numVotes; i++ {
			whenVote := ctx.BlockHeight() + r.Int63n(votingPeriod)
			fops[i] = simulation.FutureOperation{BlockHeight: int(whenVote), Op: operationSimulateMsgVote(k, sk, keys[whoVotes[i]], proposalID)}
		}
		// 3) Make an operation to ensure slashes were done correctly. (Really should be a future invariant)
		// TODO: Find a way to check if a validator was slashed other than just checking their balance a block
		// before and after.

		return action, fops, nil
	}
}

// SimulateMsgSubmitProposal simulates a msg Submit Proposal
// Note: Currently doesn't ensure that the proposal txt is in JSON form
func SimulateMsgSubmitProposal(k gov.Keeper, sk stake.Keeper) simulation.Operation {
	handler := gov.NewHandler(k)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, event func(string)) (action string, fOps []simulation.FutureOperation, err error) {
		sender := simulation.RandomKey(r, keys)
		msg, err := simulationCreateMsgSubmitProposal(r, sender)
		if err != nil {
			return "", nil, err
		}
		action, _ = simulateHandleMsgSubmitProposal(msg, sk, handler, ctx, event)
		return action, nil, nil
	}
}

func simulateHandleMsgSubmitProposal(msg gov.MsgSubmitProposal, sk stake.Keeper, handler sdk.Handler, ctx sdk.Context, event func(string)) (action string, ok bool) {
	ctx, write := ctx.CacheContext()
	result := handler(ctx, msg)
	ok = result.IsOK()
	if ok {
		// Update pool to keep invariants
		pool := sk.GetPool(ctx)
		pool.LooseTokens = pool.LooseTokens.Sub(sdk.NewDecFromInt(msg.InitialDeposit.AmountOf(denom)))
		sk.SetPool(ctx, pool)
		write()
	}
	event(fmt.Sprintf("gov/MsgSubmitProposal/%v", ok))
	action = fmt.Sprintf("TestMsgSubmitProposal: ok %v, msg %s", ok, msg.GetSignBytes())
	return
}

func simulationCreateMsgSubmitProposal(r *rand.Rand, sender crypto.PrivKey) (msg gov.MsgSubmitProposal, err error) {
	addr := sdk.AccAddress(sender.PubKey().Address())
	deposit := randomDeposit(r)
	msg = gov.NewMsgSubmitProposal(
		simulation.RandStringOfLength(r, 5),
		simulation.RandStringOfLength(r, 5),
		gov.ProposalTypeText,
		addr,
		deposit,
	)
	if msg.ValidateBasic() != nil {
		err = fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
	}
	return
}

// SimulateMsgDeposit
func SimulateMsgDeposit(k gov.Keeper, sk stake.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, event func(string)) (action string, fOp []simulation.FutureOperation, err error) {
		key := simulation.RandomKey(r, keys)
		addr := sdk.AccAddress(key.PubKey().Address())
		proposalID, ok := randomProposalID(r, k, ctx)
		if !ok {
			return "no-operation", nil, nil
		}
		deposit := randomDeposit(r)
		msg := gov.NewMsgDeposit(addr, proposalID, deposit)
		if msg.ValidateBasic() != nil {
			return "", nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}
		ctx, write := ctx.CacheContext()
		result := gov.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			// Update pool to keep invariants
			pool := sk.GetPool(ctx)
			pool.LooseTokens = pool.LooseTokens.Sub(sdk.NewDecFromInt(deposit.AmountOf(denom)))
			sk.SetPool(ctx, pool)
			write()
		}
		event(fmt.Sprintf("gov/MsgDeposit/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgDeposit: ok %v, msg %s", result.IsOK(), msg.GetSignBytes())
		return action, nil, nil
	}
}

// SimulateMsgVote
// nolint: unparam
func SimulateMsgVote(k gov.Keeper, sk stake.Keeper) simulation.Operation {
	return operationSimulateMsgVote(k, sk, nil, -1)
}

// nolint: unparam
func operationSimulateMsgVote(k gov.Keeper, sk stake.Keeper, key crypto.PrivKey, proposalID int64) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, event func(string)) (action string, fOp []simulation.FutureOperation, err error) {
		if key == nil {
			key = simulation.RandomKey(r, keys)
		}

		var ok bool

		if proposalID < 0 {
			proposalID, ok = randomProposalID(r, k, ctx)
			if !ok {
				return "no-operation", nil, nil
			}
		}
		addr := sdk.AccAddress(key.PubKey().Address())
		option := randomVotingOption(r)

		msg := gov.NewMsgVote(addr, proposalID, option)
		if msg.ValidateBasic() != nil {
			return "", nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}

		ctx, write := ctx.CacheContext()
		result := gov.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}

		event(fmt.Sprintf("gov/MsgVote/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgVote: ok %v, msg %s", result.IsOK(), msg.GetSignBytes())
		return action, nil, nil
	}
}

// Pick a random deposit
func randomDeposit(r *rand.Rand) sdk.Coins {
	// TODO Choose based on account balance and min deposit
	amount := int64(r.Intn(20)) + 1
	return sdk.Coins{sdk.NewInt64Coin(denom, amount)}
}

// Pick a random proposal ID
func randomProposalID(r *rand.Rand, k gov.Keeper, ctx sdk.Context) (proposalID int64, ok bool) {
	lastProposalID := k.GetLastProposalID(ctx)
	if lastProposalID < 1 {
		return 0, false
	}
	proposalID = int64(r.Intn(int(lastProposalID)))
	return proposalID, true
}

// Pick a random voting option
func randomVotingOption(r *rand.Rand) gov.VoteOption {
	switch r.Intn(4) {
	case 0:
		return gov.OptionYes
	case 1:
		return gov.OptionAbstain
	case 2:
		return gov.OptionNo
	case 3:
		return gov.OptionNoWithVeto
	}
	panic("should not happen")
}
