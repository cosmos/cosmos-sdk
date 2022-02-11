package gov_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestTickExpiredDepositPeriod(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govMsgSvr := keeper.NewMsgServerImpl(app.GovKeeper)

	inactiveQueue := app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg, err := v1beta2.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0].String(),
		nil,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(sdk.WrapSDKContext(ctx), newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	gov.EndBlocker(ctx, app.GovKeeper)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
}

func TestTickMultipleExpiredDepositPeriod(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govMsgSvr := keeper.NewMsgServerImpl(app.GovKeeper)

	inactiveQueue := app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg, err := v1beta2.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0].String(),
		nil,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(sdk.WrapSDKContext(ctx), newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(2) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg2, err := v1beta2.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0].String(),
		nil,
	)
	require.NoError(t, err)

	res, err = govMsgSvr.SubmitProposal(sdk.WrapSDKContext(ctx), newProposalMsg2)
	require.NoError(t, err)
	require.NotNil(t, res)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	gov.EndBlocker(ctx, app.GovKeeper)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(5) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	gov.EndBlocker(ctx, app.GovKeeper)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
}

func TestTickPassedDepositPeriod(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govMsgSvr := keeper.NewMsgServerImpl(app.GovKeeper)

	inactiveQueue := app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	newProposalMsg, err := v1beta2.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0].String(),
		nil,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(sdk.WrapSDKContext(ctx), newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	proposalID := res.ProposalId

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newDepositMsg := v1beta2.NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)})

	res1, err := govMsgSvr.Deposit(sdk.WrapSDKContext(ctx), newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res1)

	activeQueue = app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}

func TestTickPassedVotingPeriod(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	SortAddresses(addrs)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govMsgSvr := keeper.NewMsgServerImpl(app.GovKeeper)

	inactiveQueue := app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, app.StakingKeeper.TokensFromConsensusPower(ctx, 5))}
	newProposalMsg, err := v1beta2.NewMsgSubmitProposal([]sdk.Msg{mkTestLegacyContent(t)}, proposalCoins, addrs[0].String(), nil)
	require.NoError(t, err)

	wrapCtx := sdk.WrapSDKContext(ctx)

	res, err := govMsgSvr.SubmitProposal(wrapCtx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	proposalID := res.ProposalId

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newDepositMsg := v1beta2.NewMsgDeposit(addrs[1], proposalID, proposalCoins)

	res1, err := govMsgSvr.Deposit(wrapCtx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res1)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod).Add(*app.GovKeeper.GetVotingParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	activeQueue = app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, activeQueue.Valid())

	activeProposalID := types.GetProposalIDFromBytes(activeQueue.Value())
	proposal, ok := app.GovKeeper.GetProposal(ctx, activeProposalID)
	require.True(t, ok)
	require.Equal(t, v1beta2.StatusVotingPeriod, proposal.Status)

	activeQueue.Close()

	gov.EndBlocker(ctx, app.GovKeeper)

	activeQueue = app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}

func TestProposalPassedEndblocker(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	SortAddresses(addrs)

	govMsgSvr := keeper.NewMsgServerImpl(app.GovKeeper)
	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	valAddr := sdk.ValAddress(addrs[0])

	createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
	staking.EndBlocker(ctx, app.StakingKeeper)

	macc := app.GovKeeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	initialModuleAccCoins := app.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

	proposal, err := app.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, nil)
	require.NoError(t, err)

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, app.StakingKeeper.TokensFromConsensusPower(ctx, 10))}
	newDepositMsg := v1beta2.NewMsgDeposit(addrs[0], proposal.ProposalId, proposalCoins)

	res, err := govMsgSvr.Deposit(sdk.WrapSDKContext(ctx), newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	macc = app.GovKeeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	moduleAccCoins := app.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

	deposits := initialModuleAccCoins.Add(proposal.TotalDeposit...).Add(proposalCoins...)
	require.True(t, moduleAccCoins.IsEqual(deposits))

	err = app.GovKeeper.AddVote(ctx, proposal.ProposalId, addrs[0], v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes))
	require.NoError(t, err)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod).Add(*app.GovKeeper.GetVotingParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	gov.EndBlocker(ctx, app.GovKeeper)

	macc = app.GovKeeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	require.True(t, app.BankKeeper.GetAllBalances(ctx, macc.GetAddress()).IsEqual(initialModuleAccCoins))
}

func TestEndBlockerProposalHandlerFailed(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 1, valTokens)

	SortAddresses(addrs)

	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(app.StakingKeeper)
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	valAddr := sdk.ValAddress(addrs[0])

	createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
	staking.EndBlocker(ctx, app.StakingKeeper)

	// Create a proposal where the handler will pass for the test proposal
	// because the value of contextKeyBadProposal is true.
	ctx = ctx.WithValue(contextKeyBadProposal, true)
	proposal, err := app.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, nil)
	require.NoError(t, err)

	proposalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, app.StakingKeeper.TokensFromConsensusPower(ctx, 10)))
	newDepositMsg := v1beta2.NewMsgDeposit(addrs[0], proposal.ProposalId, proposalCoins)

	govMsgSvr := keeper.NewMsgServerImpl(app.GovKeeper)
	res, err := govMsgSvr.Deposit(sdk.WrapSDKContext(ctx), newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	err = app.GovKeeper.AddVote(ctx, proposal.ProposalId, addrs[0], v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes))
	require.NoError(t, err)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod).Add(*app.GovKeeper.GetVotingParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	// Set the contextKeyBadProposal value to false so that the handler will fail
	// during the processing of the proposal in the EndBlocker.
	ctx = ctx.WithValue(contextKeyBadProposal, false)

	// validate that the proposal fails/has been rejected
	gov.EndBlocker(ctx, app.GovKeeper)
}

func createValidators(t *testing.T, stakingMsgSvr stakingtypes.MsgServer, ctx sdk.Context, addrs []sdk.ValAddress, powerAmt []int64) {
	require.True(t, len(addrs) <= len(pubkeys), "Not enough pubkeys specified at top of file.")

	for i := 0; i < len(addrs); i++ {
		valTokens := sdk.TokensFromConsensusPower(powerAmt[i], sdk.DefaultPowerReduction)
		valCreateMsg, err := stakingtypes.NewMsgCreateValidator(
			addrs[i], pubkeys[i], sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
			TestDescription, TestCommissionRates, sdk.OneInt(),
		)
		require.NoError(t, err)
		res, err := stakingMsgSvr.CreateValidator(sdk.WrapSDKContext(ctx), valCreateMsg)
		require.NoError(t, err)
		require.NotNil(t, res)
	}
}
