package simulation

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const defaultMinProposalID = 100000000000000

var minProposalID = defaultMinProposalID

// ContentSimulator defines a function type alias for generating random proposal
// content.
type ContentSimulator func(r *rand.Rand, ctx sdk.Context, accs []simulation.Account) types.Content

// SimulateSubmitProposal simulates creating a msg Submit Proposal
// voting on the proposal, and subsequently slashing the proposal. It is implemented using
// future operations.
func SimulateSubmitProposal(ak types.AccountKeeper,
	k keeper.Keeper, contentSim ContentSimulator) simulation.Operation {
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

	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		// 1) submit proposal now
		content := contentSim(r, ctx, accs)
		if content == nil {
			// skip
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, _ := simulation.RandomAcc(r, accs)
		deposit, skip, err := randomDeposit(r, ctx, ak, k, simAccount.Address)
		switch {
		case skip:
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		case err != nil:
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		msg := types.NewMsgSubmitProposal(content, deposit, simAccount.Address)

		account := ak.GetAccount(ctx, simAccount.Address)
		coins := account.SpendableCoins(ctx.BlockTime())

		var fees sdk.Coins
		coins, hasNeg := coins.SafeSub(deposit)
		if !hasNeg {
			fees, err = simulation.RandomFees(r, ctx, coins)
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, err
			}
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			[]crypto.PrivKey{simAccount.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		opMsg := simulation.NewOperationMsg(msg, true, "")

		// get the submitted proposal ID
		proposalID, err := k.GetProposalID(ctx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		// 2) Schedule operations for votes
		// 2.1) first pick a number of people to vote.
		curNumVotesState = numVotesTransitionMatrix.NextState(r, curNumVotesState)
		numVotes := int(math.Ceil(float64(len(accs)) * statePercentageArray[curNumVotesState]))

		// 2.2) select who votes and when
		whoVotes := r.Perm(len(accs))

		// didntVote := whoVotes[numVotes:]
		whoVotes = whoVotes[:numVotes]
		votingPeriod := k.GetVotingParams(ctx).VotingPeriod

		fops := make([]simulation.FutureOperation, numVotes+1)
		for i := 0; i < numVotes; i++ {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simulation.FutureOperation{
				BlockTime: whenVote,
				Op:        operationSimulateMsgVote(ak, k, accs[whoVotes[i]], int64(proposalID)),
			}
		}

		// 3) Make an operation to ensure slashes were done correctly. (Really should be a future invariant)
		// TODO: Find a way to check if a validator was slashed other than just checking their balance a block
		// before and after.

		return opMsg, fops, nil
	}
}

// SimulateTextProposalContent returns random text proposal content.
func SimulateTextProposalContent(r *rand.Rand, _ sdk.Context, _ []simulation.Account) types.Content {
	return types.NewTextProposal(
		simulation.RandStringOfLength(r, 140),
		simulation.RandStringOfLength(r, 5000),
	)
}

// SimulateMsgDeposit generates a MsgDeposit with random values.
func SimulateMsgDeposit(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		simAccount, _ := simulation.RandomAcc(r, accs)
		proposalID, ok := randomProposalID(r, k, ctx, types.StatusDepositPeriod)
		if !ok {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		deposit, skip, err := randomDeposit(r, ctx, ak, k, simAccount.Address)
		switch {
		case skip:
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		case err != nil:
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		msg := types.NewMsgDeposit(simAccount.Address, proposalID, deposit)

		account := ak.GetAccount(ctx, simAccount.Address)
		coins := account.SpendableCoins(ctx.BlockTime())

		var fees sdk.Coins
		coins, hasNeg := coins.SafeSub(deposit)
		if !hasNeg {
			fees, err = simulation.RandomFees(r, ctx, coins)
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, err
			}
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			[]crypto.PrivKey{simAccount.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgVote generates a MsgVote with random values.
func SimulateMsgVote(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return operationSimulateMsgVote(ak, k, simulation.Account{}, -1)
}

func operationSimulateMsgVote(ak types.AccountKeeper, k keeper.Keeper,
	simAccount simulation.Account, proposalIDInt int64) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		if simAccount.Equals(simulation.Account{}) {
			simAccount, _ = simulation.RandomAcc(r, accs)
		}

		var proposalID uint64

		switch {
		case proposalIDInt < 0:
			var ok bool
			proposalID, ok = randomProposalID(r, k, ctx, types.StatusVotingPeriod)
			if !ok {
				return simulation.NoOpMsg(types.ModuleName), nil, nil
			}
		default:
			proposalID = uint64(proposalIDInt)
		}

		option := randomVotingOption(r)

		msg := types.NewMsgVote(simAccount.Address, proposalID, option)

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			[]crypto.PrivKey{simAccount.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// Pick a random deposit
func randomDeposit(r *rand.Rand, ctx sdk.Context,
	ak types.AccountKeeper, k keeper.Keeper, addr sdk.AccAddress,
) (deposit sdk.Coins, skip bool, err error) {
	account := ak.GetAccount(ctx, addr)
	coins := account.SpendableCoins(ctx.BlockHeader().Time)
	if coins.Empty() {
		return nil, true, nil // skip
	}

	minDeposit := k.GetDepositParams(ctx).MinDeposit
	denomIndex := r.Intn(len(minDeposit))
	denom := minDeposit[denomIndex].Denom

	depositCoins := coins.AmountOf(denom)
	if depositCoins.IsZero() {
		return nil, true, nil
	}

	maxAmt := depositCoins
	if maxAmt.GT(minDeposit[0].Amount) {
		maxAmt = minDeposit[0].Amount
	}

	amount, err := simulation.RandPositiveInt(r, maxAmt)
	if err != nil {
		return nil, false, err
	}

	return sdk.Coins{sdk.NewCoin(denom, amount)}, false, nil
}

// Pick a random proposal ID from a proposal with a given status.
// It does not provide a default proposal ID.
func randomProposalID(r *rand.Rand, k keeper.Keeper,
	ctx sdk.Context, status types.ProposalStatus) (proposalID uint64, found bool) {
	proposalID, _ = k.GetProposalID(ctx)
	if minProposalID == defaultMinProposalID {
		minProposalID = int(math.Min(float64(minProposalID), float64(proposalID)))
	}

	if minProposalID < int(proposalID) {
		proposalID = uint64(simulation.RandIntBetween(r, minProposalID, int(proposalID)))
	}

	proposal, ok := k.GetProposal(ctx, proposalID)
	if !ok || proposal.Status != status {
		return proposalID, false
	}

	return proposalID, true
}

// Pick a random voting option
func randomVotingOption(r *rand.Rand) types.VoteOption {
	switch r.Intn(4) {
	case 0:
		return types.OptionYes
	case 1:
		return types.OptionAbstain
	case 2:
		return types.OptionNo
	case 3:
		return types.OptionNoWithVeto
	default:
		panic("invalid vote option")
	}
}
