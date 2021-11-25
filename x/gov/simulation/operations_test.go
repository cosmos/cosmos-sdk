package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	types.RegisterLegacyAminoCodec(amino)
	amino.RegisterInterface((*sdk.Msg)(nil), nil)
}

// TestWeightedOperations tests the weights of the operations.
func TestWeightedOperations(t *testing.T) {
	app, ctx := createTestApp(t, false)
	ctx.WithChainID("test-chain")

	cdc := app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, app.AccountKeeper,
		app.BankKeeper, app.GovKeeper,
	)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := getTestingAccounts(t, r, app, ctx, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simappparams.DefaultWeightMsgSignalProposal, types.ModuleName, types.TypeMsgSubmitProposal},
		{simappparams.DefaultWeightMsgDeposit, types.ModuleName, types.TypeMsgDeposit},
		{simappparams.DefaultWeightMsgVote, types.ModuleName, types.TypeMsgVote},
		{simappparams.DefaultWeightMsgVoteWeighted, types.ModuleName, types.TypeMsgVoteWeighted},
	}

	for i, w := range weightedOps {
		operationMsg, _, _ := w.Op()(r, app.BaseApp, ctx, accs, ctx.ChainID())
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		require.Equal(t, expected[i].weight, w.Weight(), "weight should be the same")
		require.Equal(t, expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		require.Equal(t, expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgSubmitProposal tests the normal scenario of a valid message of type TypeMsgSubmitProposal.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func TestSimulateMsgSubmitProposal(t *testing.T) {
	app, ctx := createTestApp(t, false)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgSubmitProposal(app.AccountKeeper, app.BankKeeper, app.GovKeeper)
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgSubmitProposal
	err = ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	require.NoError(t, err)

	govAcc := app.GovKeeper.GetGovernanceAccount(ctx).GetAddress()
	expectedProposalMsgs := []sdk.Msg{types.NewMsgVote(govAcc, 1, types.OptionYes)}

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Proposer)
	require.Equal(t, "560969stake", msg.InitialDeposit.String())
	require.Equal(t, "gov", msg.Route())
	proposalMsgs, err := msg.GetMessages()
	require.NoError(t, err)
	require.Equal(t, expectedProposalMsgs, proposalMsgs)
	require.Equal(t, types.TypeMsgSubmitProposal, msg.Type())
}

// TestSimulateMsgDeposit tests the normal scenario of a valid message of type TypeMsgDeposit.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func TestSimulateMsgDeposit(t *testing.T) {
	app, ctx := createTestApp(t, false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	submitTime := ctx.BlockHeader().Time
	depositPeriod := app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod

	proposalMsgs := []sdk.Msg{types.NewMsgVote(accounts[0].Address, 1, types.OptionYes)}
	proposal, err := types.NewProposal(proposalMsgs, 1, submitTime, submitTime.Add(depositPeriod))
	require.NoError(t, err)

	app.GovKeeper.SetProposal(ctx, proposal)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgDeposit(app.AccountKeeper, app.BankKeeper, app.GovKeeper)
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgDeposit
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, uint64(1), msg.ProposalId)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Depositor)
	require.Equal(t, "560969stake", msg.Amount.String())
	require.Equal(t, "gov", msg.Route())
	require.Equal(t, types.TypeMsgDeposit, msg.Type())
}

// TestSimulateMsgVote tests the normal scenario of a valid message of type TypeMsgVote.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func TestSimulateMsgVote(t *testing.T) {
	app, ctx := createTestApp(t, false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	submitTime := ctx.BlockHeader().Time
	depositPeriod := app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod

	proposalMsgs := []sdk.Msg{types.NewMsgVote(accounts[0].Address, 1, types.OptionYes)}
	proposal, err := types.NewProposal(proposalMsgs, 1, submitTime, submitTime.Add(depositPeriod))
	require.NoError(t, err)

	app.GovKeeper.ActivateVotingPeriod(ctx, proposal)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgVote(app.AccountKeeper, app.BankKeeper, app.GovKeeper)
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgVote
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, uint64(1), msg.ProposalId)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Voter)
	require.Equal(t, types.OptionYes, msg.Option)
	require.Equal(t, "gov", msg.Route())
	require.Equal(t, types.TypeMsgVote, msg.Type())
}

// TestSimulateMsgVoteWeighted tests the normal scenario of a valid message of type TypeMsgVoteWeighted.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func TestSimulateMsgVoteWeighted(t *testing.T) {
	app, ctx := createTestApp(t, false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	submitTime := ctx.BlockHeader().Time
	depositPeriod := app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod

	proposalMsgs := []sdk.Msg{types.NewMsgVote(accounts[0].Address, 1, types.OptionYes)}
	proposal, err := types.NewProposal(proposalMsgs, 1, submitTime, submitTime.Add(depositPeriod))
	require.NoError(t, err)

	app.GovKeeper.ActivateVotingPeriod(ctx, proposal)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgVoteWeighted(app.AccountKeeper, app.BankKeeper, app.GovKeeper)
	operationMsg, _, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgVoteWeighted
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, uint64(1), msg.ProposalId)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Voter)
	require.True(t, len(msg.Options) >= 1)
	require.Equal(t, "gov", msg.Route())
	require.Equal(t, types.TypeMsgVoteWeighted, msg.Type())
}

// returns context and an app with updated mint keeper
func createTestApp(t *testing.T, isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(t, isCheckTx)

	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	app.MintKeeper.SetParams(ctx, minttypes.DefaultParams())
	app.MintKeeper.SetMinter(ctx, minttypes.DefaultInitialMinter())

	return app, ctx
}

func getTestingAccounts(t *testing.T, r *rand.Rand, app *simapp.SimApp, ctx sdk.Context, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := app.StakingKeeper.TokensFromConsensusPower(ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
		app.AccountKeeper.SetAccount(ctx, acc)
		require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, account.Address, initCoins))
	}

	return accounts
}
