package gov_test

import (
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestTickExpiredDepositPeriod(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govHandler := gov.NewHandler(app.GovKeeper)

	inactiveQueue := app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg, err := types.NewMsgSubmitProposal(
		types.ContentFromProposalType("test", "test", types.ProposalTypeText),
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0],
	)
	require.NoError(t, err)

	res, err := govHandler(ctx, newProposalMsg)
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
	newHeader.Time = ctx.BlockHeader().Time.Add(app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod)
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
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govHandler := gov.NewHandler(app.GovKeeper)

	inactiveQueue := app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg, err := types.NewMsgSubmitProposal(
		types.ContentFromProposalType("test", "test", types.ProposalTypeText),
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0],
	)
	require.NoError(t, err)

	res, err := govHandler(ctx, newProposalMsg)
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

	newProposalMsg2, err := types.NewMsgSubmitProposal(
		types.ContentFromProposalType("test2", "test2", types.ProposalTypeText),
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0],
	)
	require.NoError(t, err)

	res, err = govHandler(ctx, newProposalMsg2)
	require.NoError(t, err)
	require.NotNil(t, res)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
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
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govHandler := gov.NewHandler(app.GovKeeper)

	inactiveQueue := app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	newProposalMsg, err := types.NewMsgSubmitProposal(
		types.ContentFromProposalType("test2", "test2", types.ProposalTypeText),
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0],
	)
	require.NoError(t, err)

	res, err := govHandler(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	var proposalData types.MsgSubmitProposalResponse
	err = proto.Unmarshal(res.Data, &proposalData)
	require.NoError(t, err)

	proposalID := proposalData.ProposalId

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newDepositMsg := types.NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)})

	res, err = govHandler(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	activeQueue = app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}

func TestTickPassedVotingPeriod(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	SortAddresses(addrs)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govHandler := gov.NewHandler(app.GovKeeper)

	inactiveQueue := app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(5))}
	newProposalMsg, err := types.NewMsgSubmitProposal(TestProposal, proposalCoins, addrs[0])
	require.NoError(t, err)

	res, err := govHandler(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	var proposalData types.MsgSubmitProposalResponse
	err = proto.Unmarshal(res.Data, &proposalData)
	require.NoError(t, err)

	proposalID := proposalData.ProposalId

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newDepositMsg := types.NewMsgDeposit(addrs[1], proposalID, proposalCoins)

	res, err = govHandler(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod).Add(app.GovKeeper.GetVotingParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	activeQueue = app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, activeQueue.Valid())

	activeProposalID := types.GetProposalIDFromBytes(activeQueue.Value())
	proposal, ok := app.GovKeeper.GetProposal(ctx, activeProposalID)
	require.True(t, ok)
	require.Equal(t, types.StatusVotingPeriod, proposal.Status)

	activeQueue.Close()

	gov.EndBlocker(ctx, app.GovKeeper)

	activeQueue = app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}

func TestProposalPassedEndblocker(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	SortAddresses(addrs)

	handler := gov.NewHandler(app.GovKeeper)
	stakingHandler := staking.NewHandler(app.StakingKeeper)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	valAddr := sdk.ValAddress(addrs[0])

	createValidators(t, stakingHandler, ctx, []sdk.ValAddress{valAddr}, []int64{10})
	staking.EndBlocker(ctx, app.StakingKeeper)

	macc := app.GovKeeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	initialModuleAccCoins := app.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

	proposal, err := app.GovKeeper.SubmitProposal(ctx, TestProposal)
	require.NoError(t, err)

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10))}
	newDepositMsg := types.NewMsgDeposit(addrs[0], proposal.ProposalId, proposalCoins)

	handleAndCheck(t, handler, ctx, newDepositMsg)

	macc = app.GovKeeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	moduleAccCoins := app.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

	deposits := initialModuleAccCoins.Add(proposal.TotalDeposit...).Add(proposalCoins...)
	require.True(t, moduleAccCoins.IsEqual(deposits))

	err = app.GovKeeper.AddVote(ctx, proposal.ProposalId, addrs[0], types.OptionYes)
	require.NoError(t, err)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod).Add(app.GovKeeper.GetVotingParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	gov.EndBlocker(ctx, app.GovKeeper)

	macc = app.GovKeeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	require.True(t, app.BankKeeper.GetAllBalances(ctx, macc.GetAddress()).IsEqual(initialModuleAccCoins))
}

func TestEndBlockerProposalHandlerFailed(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 1, valTokens)

	SortAddresses(addrs)

	stakingHandler := staking.NewHandler(app.StakingKeeper)
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	valAddr := sdk.ValAddress(addrs[0])

	createValidators(t, stakingHandler, ctx, []sdk.ValAddress{valAddr}, []int64{10})
	staking.EndBlocker(ctx, app.StakingKeeper)

	// Create a proposal where the handler will pass for the test proposal
	// because the value of contextKeyBadProposal is true.
	ctx = ctx.WithValue(contextKeyBadProposal, true)
	proposal, err := app.GovKeeper.SubmitProposal(ctx, TestProposal)
	require.NoError(t, err)

	proposalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10)))
	newDepositMsg := types.NewMsgDeposit(addrs[0], proposal.ProposalId, proposalCoins)

	handleAndCheck(t, gov.NewHandler(app.GovKeeper), ctx, newDepositMsg)

	err = app.GovKeeper.AddVote(ctx, proposal.ProposalId, addrs[0], types.OptionYes)
	require.NoError(t, err)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(app.GovKeeper.GetDepositParams(ctx).MaxDepositPeriod).Add(app.GovKeeper.GetVotingParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	// Set the contextKeyBadProposal value to false so that the handler will fail
	// during the processing of the proposal in the EndBlocker.
	ctx = ctx.WithValue(contextKeyBadProposal, false)

	// validate that the proposal fails/has been rejected
	gov.EndBlocker(ctx, app.GovKeeper)
}
