package gov_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	codecstd "github.com/cosmos/cosmos-sdk/codec/std"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestTickExpiredDepositPeriod(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	header := abci.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govHandler := gov.NewHandler(app.GovKeeper)

	inactiveQueue := app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg, err := codecstd.NewMsgSubmitProposal(
		gov.ContentFromProposalType("test", "test", gov.ProposalTypeText),
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
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	header := abci.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govHandler := gov.NewHandler(app.GovKeeper)

	inactiveQueue := app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg, err := codecstd.NewMsgSubmitProposal(
		gov.ContentFromProposalType("test", "test", gov.ProposalTypeText),
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

	newProposalMsg2, err := codecstd.NewMsgSubmitProposal(
		gov.ContentFromProposalType("test2", "test2", gov.ProposalTypeText),
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
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	header := abci.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govHandler := gov.NewHandler(app.GovKeeper)

	inactiveQueue := app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	newProposalMsg, err := codecstd.NewMsgSubmitProposal(
		gov.ContentFromProposalType("test2", "test2", gov.ProposalTypeText),
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0],
	)
	require.NoError(t, err)

	res, err := govHandler(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	proposalID := gov.GetProposalIDFromBytes(res.Data)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newDepositMsg := gov.NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)})

	res, err = govHandler(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	activeQueue = app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}

func TestTickPassedVotingPeriod(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	SortAddresses(addrs)

	header := abci.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govHandler := gov.NewHandler(app.GovKeeper)

	inactiveQueue := app.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(5))}
	newProposalMsg, err := codecstd.NewMsgSubmitProposal(TestProposal, proposalCoins, addrs[0])
	require.NoError(t, err)

	res, err := govHandler(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	proposalID := gov.GetProposalIDFromBytes(res.Data)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	newDepositMsg := gov.NewMsgDeposit(addrs[1], proposalID, proposalCoins)

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

	activeProposalID := gov.GetProposalIDFromBytes(activeQueue.Value())
	proposal, ok := app.GovKeeper.GetProposal(ctx, activeProposalID)
	require.True(t, ok)
	require.Equal(t, gov.StatusVotingPeriod, proposal.Status)

	activeQueue.Close()

	gov.EndBlocker(ctx, app.GovKeeper)

	activeQueue = app.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}

func TestProposalPassedEndblocker(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 10, valTokens)

	SortAddresses(addrs)

	handler := gov.NewHandler(app.GovKeeper)
	stakingHandler := staking.NewHandler(app.StakingKeeper)

	header := abci.Header{Height: app.LastBlockHeight() + 1}
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
	newDepositMsg := gov.NewMsgDeposit(addrs[0], proposal.ProposalID, proposalCoins)

	res, err := handler(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	macc = app.GovKeeper.GetGovernanceAccount(ctx)
	require.NotNil(t, macc)
	moduleAccCoins := app.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

	deposits := initialModuleAccCoins.Add(proposal.TotalDeposit...).Add(proposalCoins...)
	require.True(t, moduleAccCoins.IsEqual(deposits))

	err = app.GovKeeper.AddVote(ctx, proposal.ProposalID, addrs[0], gov.OptionYes)
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
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 1, valTokens)

	SortAddresses(addrs)

	handler := gov.NewHandler(app.GovKeeper)
	stakingHandler := staking.NewHandler(app.StakingKeeper)

	header := abci.Header{Height: app.LastBlockHeight() + 1}
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
	newDepositMsg := gov.NewMsgDeposit(addrs[0], proposal.ProposalID, proposalCoins)

	res, err := handler(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	err = app.GovKeeper.AddVote(ctx, proposal.ProposalID, addrs[0], gov.OptionYes)
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
