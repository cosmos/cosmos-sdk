package gov_test

import (
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestUnregisteredProposal_InactiveProposalFails(t *testing.T) {
	suite := createTestSuite(t)
	ctx := suite.App.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)

	// manually set proposal in store
	startTime, endTime := time.Now().Add(-4*time.Hour), ctx.BlockHeader().Time
	proposal, err := v1.NewProposal([]sdk.Msg{
		&v1.Proposal{}, // invalid proposal message
	}, 1, startTime, startTime, "", "Unsupported proposal", "Unsupported proposal", addrs[0])
	require.NoError(t, err)

	err = suite.GovKeeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// manually set proposal in inactive proposal queue
	err = suite.GovKeeper.InactiveProposalsQueue.Set(ctx, collections.Join(endTime, proposal.Id), proposal.Id)
	require.NoError(t, err)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	err = gov.EndBlocker(ctx, suite.GovKeeper)
	require.NoError(t, err)

	_, err = suite.GovKeeper.Proposals.Get(ctx, proposal.Id)
	require.Error(t, err, collections.ErrNotFound)
}

func TestUnregisteredProposal_ActiveProposalFails(t *testing.T) {
	suite := createTestSuite(t)
	ctx := suite.App.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)

	// manually set proposal in store
	startTime, endTime := time.Now().Add(-4*time.Hour), ctx.BlockHeader().Time
	proposal, err := v1.NewProposal([]sdk.Msg{
		&v1.Proposal{}, // invalid proposal message
	}, 1, startTime, startTime, "", "Unsupported proposal", "Unsupported proposal", addrs[0])
	require.NoError(t, err)
	proposal.Status = v1.StatusVotingPeriod
	proposal.VotingEndTime = &endTime

	err = suite.GovKeeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// manually set proposal in active proposal queue
	err = suite.GovKeeper.ActiveProposalsQueue.Set(ctx, collections.Join(endTime, proposal.Id), proposal.Id)
	require.NoError(t, err)

	checkActiveProposalsQueue(t, ctx, suite.GovKeeper)

	err = gov.EndBlocker(ctx, suite.GovKeeper)
	require.NoError(t, err)

	p, err := suite.GovKeeper.Proposals.Get(ctx, proposal.Id)
	require.NoError(t, err)
	require.Equal(t, v1.StatusFailed, p.Status)
}

func TestTickExpiredDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)

	app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Hash:   app.LastCommitID().Hash,
	})

	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	params, _ := suite.GovKeeper.Params.Get(ctx)
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	err = gov.EndBlocker(ctx, suite.GovKeeper)
	require.NoError(t, err)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)
}

func TestTickMultipleExpiredDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)

	app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Hash:   app.LastCommitID().Hash,
	})

	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(2) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	newProposalMsg2, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
	)
	require.NoError(t, err)

	res, err = govMsgSvr.SubmitProposal(ctx, newProposalMsg2)
	require.NoError(t, err)
	require.NotNil(t, res)

	newHeader = ctx.BlockHeader()
	params, _ := suite.GovKeeper.Params.Get(ctx)
	newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)
	require.NoError(t, gov.EndBlocker(ctx, suite.GovKeeper))
	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(5) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)
	require.NoError(t, gov.EndBlocker(ctx, suite.GovKeeper))
	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)
}

func TestTickPassedDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)

	app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Hash:   app.LastCommitID().Hash,
	})

	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	proposalID := res.ProposalId

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	newDepositMsg := v1.NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)})

	res1, err := govMsgSvr.Deposit(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res1)

	checkActiveProposalsQueue(t, ctx, suite.GovKeeper)
}

func TestTickPassedVotingPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)

	SortAddresses(addrs)

	app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Hash:   app.LastCommitID().Hash,
	})

	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)
	checkActiveProposalsQueue(t, ctx, suite.GovKeeper)

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, suite.StakingKeeper.TokensFromConsensusPower(ctx, 5))}
	newProposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{mkTestLegacyContent(t)}, proposalCoins, addrs[0].String(), "", "Proposal", "description of proposal")
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	proposalID := res.ProposalId

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newDepositMsg := v1.NewMsgDeposit(addrs[1], proposalID, proposalCoins)

	res1, err := govMsgSvr.Deposit(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res1)

	params, _ := suite.GovKeeper.Params.Get(ctx)
	votingPeriod := params.VotingPeriod

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*votingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)
	checkActiveProposalsQueue(t, ctx, suite.GovKeeper)

	proposal, err := suite.GovKeeper.Proposals.Get(ctx, res.ProposalId)
	require.NoError(t, err)
	require.Equal(t, v1.StatusVotingPeriod, proposal.Status)

	err = gov.EndBlocker(ctx, suite.GovKeeper)
	require.NoError(t, err)
	checkActiveProposalsQueue(t, ctx, suite.GovKeeper)
}

func TestProposalPassedEndblocker(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)

	SortAddresses(addrs)

	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)
	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)

	app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Hash:   app.LastCommitID().Hash,
	})

	valAddr := sdk.ValAddress(addrs[0])
	proposer := addrs[0]

	createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
	suite.StakingKeeper.EndBlocker(ctx)

	macc := suite.GovKeeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	initialModuleAccCoins := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

	proposal, err := suite.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, "", "title", "summary", proposer)
	require.NoError(t, err)

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, suite.StakingKeeper.TokensFromConsensusPower(ctx, 10))}
	newDepositMsg := v1.NewMsgDeposit(addrs[0], proposal.Id, proposalCoins)

	res, err := govMsgSvr.Deposit(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	macc = suite.GovKeeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	moduleAccCoins := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

	deposits := initialModuleAccCoins.Add(proposal.TotalDeposit...).Add(proposalCoins...)
	require.True(t, moduleAccCoins.Equal(deposits))

	err = suite.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "")
	require.NoError(t, err)

	newHeader := ctx.BlockHeader()
	params, _ := suite.GovKeeper.Params.Get(ctx)
	newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	gov.EndBlocker(ctx, suite.GovKeeper)

	macc = suite.GovKeeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	require.True(t, suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress()).Equal(initialModuleAccCoins))
}

func TestEndBlockerProposalHandlerFailed(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 1, valTokens)

	SortAddresses(addrs)

	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)

	app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Hash:   app.LastCommitID().Hash,
	})

	valAddr := sdk.ValAddress(addrs[0])
	proposer := addrs[0]

	createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
	suite.StakingKeeper.EndBlocker(ctx)

	msg := banktypes.NewMsgSend(authtypes.NewModuleAddress(types.ModuleName), addrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000))))
	proposal, err := suite.GovKeeper.SubmitProposal(ctx, []sdk.Msg{msg}, "", "title", "summary", proposer)
	require.NoError(t, err)

	proposalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, suite.StakingKeeper.TokensFromConsensusPower(ctx, 10)))
	newDepositMsg := v1.NewMsgDeposit(addrs[0], proposal.Id, proposalCoins)

	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)
	res, err := govMsgSvr.Deposit(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	err = suite.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "")
	require.NoError(t, err)

	params, _ := suite.GovKeeper.Params.Get(ctx)
	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	// validate that the proposal fails/has been rejected
	gov.EndBlocker(ctx, suite.GovKeeper)

	// check proposal events
	events := ctx.EventManager().Events()
	attr, eventOk := events.GetAttributes(types.AttributeKeyProposalLog)
	require.True(t, eventOk)
	require.Contains(t, attr[0].Value, "failed on execution")

	proposal, err = suite.GovKeeper.Proposals.Get(ctx, proposal.Id)
	require.Nil(t, err)
	require.Equal(t, v1.StatusFailed, proposal.Status)
}

func createValidators(t *testing.T, stakingMsgSvr stakingtypes.MsgServer, ctx sdk.Context, addrs []sdk.ValAddress, powerAmt []int64) {
	require.True(t, len(addrs) <= len(pubkeys), "Not enough pubkeys specified at top of file.")

	for i := 0; i < len(addrs); i++ {
		valTokens := sdk.TokensFromConsensusPower(powerAmt[i], sdk.DefaultPowerReduction)
		valCreateMsg, err := stakingtypes.NewMsgCreateValidator(
			addrs[i].String(), pubkeys[i], sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
			TestDescription, TestCommissionRates, math.OneInt(),
		)
		require.NoError(t, err)
		res, err := stakingMsgSvr.CreateValidator(ctx, valCreateMsg)
		require.NoError(t, err)
		require.NotNil(t, res)
	}
}

func checkActiveProposalsQueue(t *testing.T, ctx sdk.Context, k *keeper.Keeper) {
	err := k.ActiveProposalsQueue.Walk(ctx, collections.NewPrefixUntilPairRange[time.Time, uint64](ctx.BlockTime()), func(key collections.Pair[time.Time, uint64], value uint64) (stop bool, err error) {
		return false, err
	})

	require.NoError(t, err)
}

func checkInactiveProposalsQueue(t *testing.T, ctx sdk.Context, k *keeper.Keeper) {
	err := k.InactiveProposalsQueue.Walk(ctx, collections.NewPrefixUntilPairRange[time.Time, uint64](ctx.BlockTime()), func(key collections.Pair[time.Time, uint64], value uint64) (stop bool, err error) {
		return false, err
	})

	require.NoError(t, err)
}
