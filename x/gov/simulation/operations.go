package simulation

import (
	"math"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

var initialProposalID = uint64(100000000000000)

// Simulation operation weights constants
const (
	OpWeightMsgDeposit      = "op_weight_msg_deposit"
	OpWeightMsgVote         = "op_weight_msg_vote"
	OpWeightMsgVoteWeighted = "op_weight_msg_weighted_vote"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, ak types.AccountKeeper,
	bk types.BankKeeper, k keeper.Keeper, wContents []simtypes.WeightedProposalContent,
) simulation.WeightedOperations {

	var (
		weightMsgDeposit      int
		weightMsgVote         int
		weightMsgVoteWeighted int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgDeposit, &weightMsgDeposit, nil,
		func(_ *rand.Rand) {
			weightMsgDeposit = simappparams.DefaultWeightMsgDeposit
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgVote, &weightMsgVote, nil,
		func(_ *rand.Rand) {
			weightMsgVote = simappparams.DefaultWeightMsgVote
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgVoteWeighted, &weightMsgVoteWeighted, nil,
		func(_ *rand.Rand) {
			weightMsgVoteWeighted = simappparams.DefaultWeightMsgVoteWeighted
		},
	)

	// generate the weighted operations for the proposal contents
	var wProposalOps simulation.WeightedOperations

	for _, wContent := range wContents {
		wContent := wContent // pin variable
		var weight int
		appParams.GetOrGenerate(cdc, wContent.AppParamsKey(), &weight, nil,
			func(_ *rand.Rand) { weight = wContent.DefaultWeight() })

		wProposalOps = append(
			wProposalOps,
			simulation.NewWeightedOperation(
				weight,
				SimulateMsgSubmitProposal(ak, bk, k, wContent.ContentSimulatorFn()),
			),
		)
	}

	wGovOps := simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgDeposit,
			SimulateMsgDeposit(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgVote,
			SimulateMsgVote(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgVoteWeighted,
			SimulateMsgVoteWeighted(ak, bk, k),
		),
	}

	return append(wProposalOps, wGovOps...)
}

// SimulateMsgSubmitProposal simulates creating a msg Submit Proposal
// voting on the proposal, and subsequently slashing the proposal. It is implemented using
// future operations.
func SimulateMsgSubmitProposal(
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, contentSim simtypes.ContentSimulatorFn,
) simtypes.Operation {
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
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// 1) submit proposal now
		content := contentSim(r, ctx, accs)
		if content == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSubmitProposal, "content is nil"), nil, nil
		}

		simAccount, _ := simtypes.RandomAcc(r, accs)
		deposit, skip, err := randomDeposit(r, ctx, ak, bk, k, simAccount.Address)
		switch {
		case skip:
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSubmitProposal, "skip deposit"), nil, nil
		case err != nil:
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSubmitProposal, "unable to generate deposit"), nil, err
		}

		msg, err := types.NewMsgSubmitProposal(content, deposit, simAccount.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate a submit proposal msg"), nil, err
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		var fees sdk.Coins
		coins, hasNeg := spendable.SafeSub(deposit)
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, coins)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate fees"), nil, err
			}
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		opMsg := simtypes.NewOperationMsg(msg, true, "", nil)

		// get the submitted proposal ID
		proposalID, err := k.GetProposalID(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate proposalID"), nil, err
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

		fops := make([]simtypes.FutureOperation, numVotes+1)
		for i := 0; i < numVotes; i++ {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simtypes.FutureOperation{
				BlockTime: whenVote,
				Op:        operationSimulateMsgVote(ak, bk, k, accs[whoVotes[i]], int64(proposalID)),
			}
		}

		return opMsg, fops, nil
	}
}

// SimulateMsgDeposit generates a MsgDeposit with random values.
func SimulateMsgDeposit(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		proposalID, ok := randomProposalID(r, k, ctx, types.StatusDepositPeriod)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeposit, "unable to generate proposalID"), nil, nil
		}

		deposit, skip, err := randomDeposit(r, ctx, ak, bk, k, simAccount.Address)
		switch {
		case skip:
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeposit, "skip deposit"), nil, nil
		case err != nil:
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeposit, "unable to generate deposit"), nil, err
		}

		msg := types.NewMsgDeposit(simAccount.Address, proposalID, deposit)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		var fees sdk.Coins
		coins, hasNeg := spendable.SafeSub(deposit)
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, coins)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate fees"), nil, err
			}
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}
		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

// SimulateMsgVote generates a MsgVote with random values.
func SimulateMsgVote(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return operationSimulateMsgVote(ak, bk, k, simtypes.Account{}, -1)
}

func operationSimulateMsgVote(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
	simAccount simtypes.Account, proposalIDInt int64) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if simAccount.Equals(simtypes.Account{}) {
			simAccount, _ = simtypes.RandomAcc(r, accs)
		}

		var proposalID uint64

		switch {
		case proposalIDInt < 0:
			var ok bool
			proposalID, ok = randomProposalID(r, k, ctx, types.StatusVotingPeriod)
			if !ok {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgVote, "unable to generate proposalID"), nil, nil
			}
		default:
			proposalID = uint64(proposalIDInt)
		}

		option := randomVotingOption(r)
		msg := types.NewMsgVote(simAccount.Address, proposalID, option)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		fees, err := simtypes.RandomFees(r, ctx, spendable)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate fees"), nil, err
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

// SimulateMsgVoteWeighted generates a MsgVoteWeighted with random values.
func SimulateMsgVoteWeighted(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return operationSimulateMsgVoteWeighted(ak, bk, k, simtypes.Account{}, -1)
}

func operationSimulateMsgVoteWeighted(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
	simAccount simtypes.Account, proposalIDInt int64) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if simAccount.Equals(simtypes.Account{}) {
			simAccount, _ = simtypes.RandomAcc(r, accs)
		}

		var proposalID uint64

		switch {
		case proposalIDInt < 0:
			var ok bool
			proposalID, ok = randomProposalID(r, k, ctx, types.StatusVotingPeriod)
			if !ok {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgVoteWeighted, "unable to generate proposalID"), nil, nil
			}
		default:
			proposalID = uint64(proposalIDInt)
		}

		options := randomWeightedVotingOptions(r)
		msg := types.NewMsgVoteWeighted(simAccount.Address, proposalID, options)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		fees, err := simtypes.RandomFees(r, ctx, spendable)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate fees"), nil, err
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

// Pick a random deposit with a random denomination with a
// deposit amount between (0, min(balance, minDepositAmount))
// This is to simulate multiple users depositing to get the
// proposal above the minimum deposit amount
func randomDeposit(r *rand.Rand, ctx sdk.Context,
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, addr sdk.AccAddress,
) (deposit sdk.Coins, skip bool, err error) {
	account := ak.GetAccount(ctx, addr)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	if spendable.Empty() {
		return nil, true, nil // skip
	}

	minDeposit := k.GetDepositParams(ctx).MinDeposit
	denomIndex := r.Intn(len(minDeposit))
	denom := minDeposit[denomIndex].Denom

	depositCoins := spendable.AmountOf(denom)
	if depositCoins.IsZero() {
		return nil, true, nil
	}

	maxAmt := depositCoins
	if maxAmt.GT(minDeposit[denomIndex].Amount) {
		maxAmt = minDeposit[denomIndex].Amount
	}

	amount, err := simtypes.RandPositiveInt(r, maxAmt)
	if err != nil {
		return nil, false, err
	}

	return sdk.Coins{sdk.NewCoin(denom, amount)}, false, nil
}

// Pick a random proposal ID between the initial proposal ID
// (defined in gov GenesisState) and the latest proposal ID
// that matches a given Status.
// It does not provide a default ID.
func randomProposalID(r *rand.Rand, k keeper.Keeper,
	ctx sdk.Context, status types.ProposalStatus) (proposalID uint64, found bool) {
	proposalID, _ = k.GetProposalID(ctx)

	switch {
	case proposalID > initialProposalID:
		// select a random ID between [initialProposalID, proposalID]
		proposalID = uint64(simtypes.RandIntBetween(r, int(initialProposalID), int(proposalID)))

	default:
		// This is called on the first call to this funcion
		// in order to update the global variable
		initialProposalID = proposalID
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

// Pick a random weighted voting options
func randomWeightedVotingOptions(r *rand.Rand) types.WeightedVoteOptions {
	w1 := r.Intn(100 + 1)
	w2 := r.Intn(100 - w1 + 1)
	w3 := r.Intn(100 - w1 - w2 + 1)
	w4 := 100 - w1 - w2 - w3
	weightedVoteOptions := types.WeightedVoteOptions{}
	if w1 > 0 {
		weightedVoteOptions = append(weightedVoteOptions, types.WeightedVoteOption{
			Option: types.OptionYes,
			Weight: sdk.NewDecWithPrec(int64(w1), 2),
		})
	}
	if w2 > 0 {
		weightedVoteOptions = append(weightedVoteOptions, types.WeightedVoteOption{
			Option: types.OptionAbstain,
			Weight: sdk.NewDecWithPrec(int64(w2), 2),
		})
	}
	if w3 > 0 {
		weightedVoteOptions = append(weightedVoteOptions, types.WeightedVoteOption{
			Option: types.OptionNo,
			Weight: sdk.NewDecWithPrec(int64(w3), 2),
		})
	}
	if w4 > 0 {
		weightedVoteOptions = append(weightedVoteOptions, types.WeightedVoteOption{
			Option: types.OptionNoWithVeto,
			Weight: sdk.NewDecWithPrec(int64(w4), 2),
		})
	}
	return weightedVoteOptions
}
