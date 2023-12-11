package simulation

import (
	"math"
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/gov/keeper"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"

	"github.com/cosmos/cosmos-sdk/client"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

var initialProposalID = uint64(100000000000000)

func GenerateMsgSubmitProposal(r *rand.Rand, ctx sdk.Context, txGen client.TxConfig, accs []simtypes.Account, ak types.AccountKeeper, bk types.BankKeeper, k *keeper.Keeper, proposals []sdk.Msg) sdk.Tx {
	simAccount, _ := simtypes.RandomAcc(r, accs)
	expedited := r.Intn(2) == 0
	deposit, _, err := randomDeposit(r, ctx, ak, bk, k, simAccount.Address, true, expedited)
	if err != nil {
		panic(err)
	}

	msg, err := v1.NewMsgSubmitProposal(
		proposals,
		deposit,
		simAccount.Address.String(),
		simtypes.RandStringOfLength(r, 100),
		simtypes.RandStringOfLength(r, 100),
		simtypes.RandStringOfLength(r, 100),
		expedited,
	)
	if err != nil {
		panic(err)
	}

	account := ak.GetAccount(ctx, simAccount.Address)
	tx, err := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
		simtestutil.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		simAccount.PrivKey,
	)
	if err != nil {
		panic(err)
	}
	return tx
}

type VoteGenerator struct {
	numVotesTransitionMatrix simtypes.TransitionMatrix
	curNumVotesState         int
}

func NewVoteGenerator() *VoteGenerator {
	// The states are:
	// column 1: All validators vote
	// column 2: 90% vote
	// column 3: 75% vote
	// column 4: 40% vote
	// column 5: 15% vote
	// column 6: noone votes
	// All columns sum to 100 for simplicity, values chosen by @valardragon semi-arbitrarily,
	// feel free to change.
	m, _ := simulation.CreateTransitionMatrix([][]int{
		{20, 10, 0, 0, 0, 0},
		{55, 50, 20, 10, 0, 0},
		{25, 25, 30, 25, 30, 15},
		{0, 15, 30, 25, 30, 30},
		{0, 0, 20, 30, 30, 30},
		{0, 0, 0, 10, 10, 25},
	})
	return &VoteGenerator{
		numVotesTransitionMatrix: m,
		curNumVotesState:         1,
	}
}

type Vote struct {
	BlockTime time.Time
	Vote      sdk.Tx
}

func (g *VoteGenerator) GenerateVotes(r *rand.Rand, ctx sdk.Context, txGen client.TxConfig, accs []simtypes.Account, ak types.AccountKeeper, bk types.BankKeeper, k *keeper.Keeper) []Vote {
	statePercentageArray := []float64{1, .9, .75, .4, .15, 0}
	// get the submitted proposal ID
	proposalID, err := k.ProposalID.Peek(ctx)
	if err != nil {
		panic(err)
	}

	// 2) Schedule operations for votes
	// 2.1) first pick a number of people to vote.
	g.curNumVotesState = g.numVotesTransitionMatrix.NextState(r, g.curNumVotesState)
	numVotes := int(math.Ceil(float64(len(accs)) * statePercentageArray[g.curNumVotesState]))

	// 2.2) select who votes and when
	whoVotes := r.Perm(len(accs))

	// didntVote := whoVotes[numVotes:]
	whoVotes = whoVotes[:numVotes]
	params, _ := k.Params.Get(ctx)
	votingPeriod := params.VotingPeriod

	votes := make([]Vote, numVotes)
	for i := range votes {
		whenVote := ctx.HeaderInfo().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
		votes[i] = Vote{
			BlockTime: whenVote,
			Vote:      GenerateMsgVote(r, ctx, txGen, accs[whoVotes[i]], ak, bk, k, int64(proposalID)),
		}
	}

	return votes
}

// GenerateMsgDeposit generates a MsgDeposit with random values.
func GenerateMsgDeposit(r *rand.Rand, ctx sdk.Context, txGen client.TxConfig, accs []simtypes.Account, ak types.AccountKeeper, bk types.BankKeeper, k *keeper.Keeper) sdk.Tx {
	simAccount, _ := simtypes.RandomAcc(r, accs)
	proposalID, _ := randomProposalID(r, k, ctx, v1.StatusDepositPeriod)

	deposit, _, err := randomDeposit(r, ctx, ak, bk, k, simAccount.Address, false, false)
	if err != nil {
		panic(err)
	}

	msg := v1.NewMsgDeposit(simAccount.Address, proposalID, deposit)

	account := ak.GetAccount(ctx, simAccount.Address)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	spendable = spendable.Sub(deposit...)
	fees, _ := simtypes.RandomFees(r, spendable)

	tx, err := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		fees,
		simtestutil.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		simAccount.PrivKey,
	)
	if err != nil {
		panic(err)
	}
	return tx
}

func GenerateMsgVote(r *rand.Rand, ctx sdk.Context, txGen client.TxConfig, acc simtypes.Account, ak types.AccountKeeper, bk types.BankKeeper, k *keeper.Keeper, proposalIDInt int64) sdk.Tx {
	var proposalID uint64

	switch {
	case proposalIDInt < 0:
		var ok bool
		proposalID, ok = randomProposalID(r, k, ctx, v1.StatusVotingPeriod)
		if !ok {
			panic("no proposal id")
		}
	default:
		proposalID = uint64(proposalIDInt)
	}

	option := randomVotingOption(r)

	msg := v1.NewMsgVote(acc.Address, proposalID, option, "")

	account := ak.GetAccount(ctx, acc.Address)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	fees, _ := simtypes.RandomFees(r, spendable)

	tx, err := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		fees,
		simtestutil.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		acc.PrivKey,
	)
	if err != nil {
		panic(err)
	}

	return tx
}

// GenerateMsgVoteWeighted generates a MsgVoteWeighted with random values.
func GenerateMsgVoteWeighted(r *rand.Rand, ctx sdk.Context, txGen client.TxConfig, accs []simtypes.Account, ak types.AccountKeeper, bk types.BankKeeper, k *keeper.Keeper) sdk.Tx {
	simAccount, _ := simtypes.RandomAcc(r, accs)

	proposalID, _ := randomProposalID(r, k, ctx, v1.StatusVotingPeriod)

	options := randomWeightedVotingOptions(r)
	msg := v1.NewMsgVoteWeighted(simAccount.Address, proposalID, options, "")

	account := ak.GetAccount(ctx, simAccount.Address)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	fees, _ := simtypes.RandomFees(r, spendable)

	tx, err := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		fees,
		simtestutil.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		simAccount.PrivKey,
	)
	if err != nil {
		panic(err)
	}

	return tx
}

// GenerateMsgCancelProposal generates a MsgCancelProposal.
func GenerateMsgCancelProposal(r *rand.Rand, ctx sdk.Context, txGen client.TxConfig, accs []simtypes.Account, ak types.AccountKeeper, bk types.BankKeeper, k *keeper.Keeper) sdk.Tx {
	simAccount := accs[0]
	proposal := randomProposal(r, k, ctx)

	account := ak.GetAccount(ctx, simAccount.Address)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	id := uint64(0)
	if proposal != nil {
		id = proposal.Id
	}
	msg := v1.NewMsgCancelProposal(id, account.GetAddress().String())

	tx, err := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		spendable,
		simtestutil.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		simAccount.PrivKey,
	)
	if err != nil {
		panic(err)
	}

	return tx
}

// Pick a random deposit with a random denomination with a
// deposit amount between (0, min(balance, minDepositAmount))
// This is to simulate multiple users depositing to get the
// proposal above the minimum deposit amount
func randomDeposit(
	r *rand.Rand,
	ctx sdk.Context,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k *keeper.Keeper,
	addr sdk.AccAddress,
	useMinAmount bool,
	expedited bool,
) (deposit sdk.Coins, skip bool, err error) {
	account := ak.GetAccount(ctx, addr)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	if spendable.Empty() {
		return nil, true, nil // skip
	}

	params, _ := k.Params.Get(ctx)
	minDeposit := params.MinDeposit
	if expedited {
		minDeposit = params.ExpeditedMinDeposit
	}
	denomIndex := r.Intn(len(minDeposit))
	denom := minDeposit[denomIndex].Denom

	spendableBalance := spendable.AmountOf(denom)
	if spendableBalance.IsZero() {
		return nil, true, nil
	}

	minDepositAmount := minDeposit[denomIndex].Amount

	minDepositRatio, err := sdkmath.LegacyNewDecFromStr(params.GetMinDepositRatio())
	if err != nil {
		return nil, false, err
	}

	threshold := minDepositAmount.ToLegacyDec().Mul(minDepositRatio).TruncateInt()

	minAmount := sdkmath.ZeroInt()
	if useMinAmount {
		minDepositPercent, err := sdkmath.LegacyNewDecFromStr(params.MinInitialDepositRatio)
		if err != nil {
			return nil, false, err
		}

		minAmount = sdkmath.LegacyNewDecFromInt(minDepositAmount).Mul(minDepositPercent).TruncateInt()
	}

	amount, err := simtypes.RandPositiveInt(r, minDepositAmount.Sub(minAmount))
	if err != nil {
		return nil, false, err
	}
	amount = amount.Add(minAmount)

	if amount.GT(spendableBalance) || amount.LT(threshold) {
		return nil, true, nil
	}

	return sdk.Coins{sdk.NewCoin(denom, amount)}, false, nil
}

// randomProposal returns a random proposal stored in state
func randomProposal(r *rand.Rand, k *keeper.Keeper, ctx sdk.Context) *v1.Proposal {
	var proposals []*v1.Proposal
	err := k.Proposals.Walk(ctx, nil, func(key uint64, value v1.Proposal) (stop bool, err error) {
		proposals = append(proposals, &value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	if len(proposals) == 0 {
		return nil
	}
	randomIndex := r.Intn(len(proposals))
	return proposals[randomIndex]
}

// Pick a random proposal ID between the initial proposal ID
// (defined in gov GenesisState) and the latest proposal ID
// that matches a given Status.
// It does not provide a default ID.
func randomProposalID(r *rand.Rand, k *keeper.Keeper, ctx sdk.Context, status v1.ProposalStatus) (proposalID uint64, found bool) {
	proposalID, _ = k.ProposalID.Peek(ctx)

	switch {
	case proposalID > initialProposalID:
		// select a random ID between [initialProposalID, proposalID]
		proposalID = uint64(simtypes.RandIntBetween(r, int(initialProposalID), int(proposalID)))

	default:
		// This is called on the first call to this funcion
		// in order to update the global variable
		initialProposalID = proposalID
	}

	proposal, err := k.Proposals.Get(ctx, proposalID)
	if err != nil || proposal.Status != status {
		return proposalID, false
	}

	return proposalID, true
}

// Pick a random voting option
func randomVotingOption(r *rand.Rand) v1.VoteOption {
	switch r.Intn(4) {
	case 0:
		return v1.OptionYes
	case 1:
		return v1.OptionAbstain
	case 2:
		return v1.OptionNo
	case 3:
		return v1.OptionNoWithVeto
	default:
		panic("invalid vote option")
	}
}

// Pick a random weighted voting options
func randomWeightedVotingOptions(r *rand.Rand) v1.WeightedVoteOptions {
	w1 := r.Intn(100 + 1)
	w2 := r.Intn(100 - w1 + 1)
	w3 := r.Intn(100 - w1 - w2 + 1)
	w4 := 100 - w1 - w2 - w3
	weightedVoteOptions := v1.WeightedVoteOptions{}
	if w1 > 0 {
		weightedVoteOptions = append(weightedVoteOptions, &v1.WeightedVoteOption{
			Option: v1.OptionYes,
			Weight: sdkmath.LegacyNewDecWithPrec(int64(w1), 2).String(),
		})
	}
	if w2 > 0 {
		weightedVoteOptions = append(weightedVoteOptions, &v1.WeightedVoteOption{
			Option: v1.OptionAbstain,
			Weight: sdkmath.LegacyNewDecWithPrec(int64(w2), 2).String(),
		})
	}
	if w3 > 0 {
		weightedVoteOptions = append(weightedVoteOptions, &v1.WeightedVoteOption{
			Option: v1.OptionNo,
			Weight: sdkmath.LegacyNewDecWithPrec(int64(w3), 2).String(),
		})
	}
	if w4 > 0 {
		weightedVoteOptions = append(weightedVoteOptions, &v1.WeightedVoteOption{
			Option: v1.OptionNoWithVeto,
			Weight: sdkmath.LegacyNewDecWithPrec(int64(w4), 2).String(),
		})
	}
	return weightedVoteOptions
}
