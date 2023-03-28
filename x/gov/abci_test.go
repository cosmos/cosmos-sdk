package gov_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestTickExpiredDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)

	header := cmtproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	inactiveQueue := suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*suite.GovKeeper.GetParams(ctx).MaxDepositPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	gov.EndBlocker(ctx, suite.GovKeeper)

	inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
}

func TestTickMultipleExpiredDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)

	header := cmtproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	inactiveQueue := suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(2) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newProposalMsg2, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	require.NoError(t, err)

	res, err = govMsgSvr.SubmitProposal(ctx, newProposalMsg2)
	require.NoError(t, err)
	require.NotNil(t, res)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*suite.GovKeeper.GetParams(ctx).MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	gov.EndBlocker(ctx, suite.GovKeeper)

	inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(5) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.True(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	gov.EndBlocker(ctx, suite.GovKeeper)

	inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
}

func TestTickPassedDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)

	header := cmtproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	inactiveQueue := suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()
	activeQueue := suite.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()

	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	proposalID := res.ProposalId

	inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, inactiveQueue.Valid())
	inactiveQueue.Close()

	newDepositMsg := v1.NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)})

	res1, err := govMsgSvr.Deposit(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res1)

	activeQueue = suite.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	require.False(t, activeQueue.Valid())
	activeQueue.Close()
}

func TestTickPassedVotingPeriod(t *testing.T) {
	testcases := []struct {
		name      string
		expedited bool
	}{
		{
			name: "regular - deleted",
		},
		{
			name:      "expedited - converted to regular",
			expedited: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			suite := createTestSuite(t)
			app := suite.App
			ctx := app.BaseApp.NewContext(false, cmtproto.Header{})
			depositMultiplier := getDepositMultiplier(tc.expedited)
			addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens.Mul(math.NewInt(depositMultiplier)))

			SortAddresses(addrs)

			header := cmtproto.Header{Height: app.LastBlockHeight() + 1}
			app.BeginBlock(abci.RequestBeginBlock{Header: header})

			govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

			inactiveQueue := suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
			require.False(t, inactiveQueue.Valid())
			inactiveQueue.Close()
			activeQueue := suite.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
			require.False(t, activeQueue.Valid())
			activeQueue.Close()

			proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, suite.StakingKeeper.TokensFromConsensusPower(ctx, 5*depositMultiplier))}
			newProposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{mkTestLegacyContent(t)}, proposalCoins, addrs[0].String(), "", "Proposal", "description of proposal", tc.expedited)
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

			params := suite.GovKeeper.GetParams(ctx)
			votingPeriod := params.VotingPeriod
			if tc.expedited {
				votingPeriod = params.ExpeditedVotingPeriod
			}

			newHeader = ctx.BlockHeader()
			newHeader.Time = ctx.BlockHeader().Time.Add(*suite.GovKeeper.GetParams(ctx).MaxDepositPeriod).Add(*votingPeriod)
			ctx = ctx.WithBlockHeader(newHeader)

			inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
			require.False(t, inactiveQueue.Valid())
			inactiveQueue.Close()

			activeQueue = suite.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
			require.True(t, activeQueue.Valid())

			activeProposalID := types.GetProposalIDFromBytes(activeQueue.Value())
			proposal, ok := suite.GovKeeper.GetProposal(ctx, activeProposalID)
			require.True(t, ok)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)

			activeQueue.Close()

			gov.EndBlocker(ctx, suite.GovKeeper)

			activeQueue = suite.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
			if !tc.expedited {
				require.False(t, activeQueue.Valid())
				activeQueue.Close()
				return
			}

			// If expedited, it should be converted to a regular proposal instead.
			require.True(t, activeQueue.Valid())

			activeProposalID = types.GetProposalIDFromBytes(activeQueue.Value())
			proposal, ok = suite.GovKeeper.GetProposal(ctx, activeProposalID)
			require.True(t, ok)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)
			require.False(t, proposal.Expedited)
			require.Equal(t, proposal.VotingStartTime.Add(*params.VotingPeriod), *proposal.VotingEndTime)

			activeQueue.Close()
		})
	}
}

func TestProposalPassedEndblocker(t *testing.T) {
	testcases := []struct {
		name      string
		expedited bool
	}{
		{
			name: "regular",
		},
		{
			name:      "expedited",
			expedited: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			suite := createTestSuite(t)
			app := suite.App
			ctx := app.BaseApp.NewContext(false, cmtproto.Header{})
			depositMultiplier := getDepositMultiplier(tc.expedited)
			addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens.Mul(math.NewInt(depositMultiplier)))

			SortAddresses(addrs)

			govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)
			stakingMsgSvr := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)

			header := cmtproto.Header{Height: app.LastBlockHeight() + 1}
			app.BeginBlock(abci.RequestBeginBlock{Header: header})

			valAddr := sdk.ValAddress(addrs[0])
			proposer := addrs[0]

			createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
			staking.EndBlocker(ctx, suite.StakingKeeper)

			macc := suite.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			initialModuleAccCoins := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

			proposal, err := suite.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, "", "title", "summary", proposer, tc.expedited)
			require.NoError(t, err)

			proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, suite.StakingKeeper.TokensFromConsensusPower(ctx, 10*depositMultiplier))}
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
			newHeader.Time = ctx.BlockHeader().Time.Add(*suite.GovKeeper.GetParams(ctx).MaxDepositPeriod).Add(*suite.GovKeeper.GetParams(ctx).VotingPeriod)
			ctx = ctx.WithBlockHeader(newHeader)

			gov.EndBlocker(ctx, suite.GovKeeper)

			macc = suite.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			require.True(t, suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress()).Equal(initialModuleAccCoins))
		})
	}
}

func TestEndBlockerProposalHandlerFailed(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 1, valTokens)

	SortAddresses(addrs)

	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)
	header := cmtproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	valAddr := sdk.ValAddress(addrs[0])
	proposer := addrs[0]

	createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
	staking.EndBlocker(ctx, suite.StakingKeeper)

	msg := banktypes.NewMsgSend(authtypes.NewModuleAddress(types.ModuleName), addrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000))))
	proposal, err := suite.GovKeeper.SubmitProposal(ctx, []sdk.Msg{msg}, "", "title", "summary", proposer, false)
	require.NoError(t, err)

	proposalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, suite.StakingKeeper.TokensFromConsensusPower(ctx, 10)))
	newDepositMsg := v1.NewMsgDeposit(addrs[0], proposal.Id, proposalCoins)

	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)
	res, err := govMsgSvr.Deposit(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	err = suite.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "")
	require.NoError(t, err)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*suite.GovKeeper.GetParams(ctx).MaxDepositPeriod).Add(*suite.GovKeeper.GetParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	// validate that the proposal fails/has been rejected
	gov.EndBlocker(ctx, suite.GovKeeper)

	// check proposal events
	events := ctx.EventManager().Events()
	attr, eventOk := events.GetAttributes(types.AttributeKeyProposalLog)
	require.True(t, eventOk)
	require.Contains(t, attr[0].Value, "failed on execution")

	proposal, ok := suite.GovKeeper.GetProposal(ctx, proposal.Id)
	require.True(t, ok)
	require.Equal(t, v1.StatusFailed, proposal.Status)
}

func TestExpeditedProposal_PassAndConversionToRegular(t *testing.T) {
	testcases := []struct {
		name string
		// indicates whether the expedited proposal passes.
		expeditedPasses bool
		// indicates whether the converted regular proposal is expected to eventually pass
		regularEventuallyPassing bool
	}{
		{
			name:            "expedited passes and not converted to regular",
			expeditedPasses: true,
		},
		{
			name:                     "expedited fails, converted to regular - regular eventually passes",
			expeditedPasses:          false,
			regularEventuallyPassing: true,
		},
		{
			name:                     "expedited fails, converted to regular - regular eventually fails",
			expeditedPasses:          false,
			regularEventuallyPassing: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			suite := createTestSuite(t)
			app := suite.App
			ctx := app.BaseApp.NewContext(false, cmtproto.Header{})
			depositMultiplier := getDepositMultiplier(true)
			addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 3, valTokens.Mul(math.NewInt(depositMultiplier)))
			params := suite.GovKeeper.GetParams(ctx)

			SortAddresses(addrs)

			govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)
			stakingMsgSvr := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)

			header := cmtproto.Header{Height: app.LastBlockHeight() + 1}
			app.BeginBlock(abci.RequestBeginBlock{Header: header})

			valAddr := sdk.ValAddress(addrs[0])
			proposer := addrs[0]

			// Create a validator so that able to vote on proposal.
			createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
			staking.EndBlocker(ctx, suite.StakingKeeper)

			inactiveQueue := suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
			require.False(t, inactiveQueue.Valid())
			inactiveQueue.Close()
			activeQueue := suite.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
			require.False(t, activeQueue.Valid())
			activeQueue.Close()

			macc := suite.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			initialModuleAccCoins := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

			submitterInitialBalance := suite.BankKeeper.GetAllBalances(ctx, addrs[0])
			depositorInitialBalance := suite.BankKeeper.GetAllBalances(ctx, addrs[1])

			proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, suite.StakingKeeper.TokensFromConsensusPower(ctx, 5*depositMultiplier))}
			newProposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{}, proposalCoins, proposer.String(), "metadata", "title", "summary", true)
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

			newHeader = ctx.BlockHeader()
			newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.ExpeditedVotingPeriod)
			ctx = ctx.WithBlockHeader(newHeader)

			inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
			require.False(t, inactiveQueue.Valid())
			inactiveQueue.Close()

			activeQueue = suite.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
			require.True(t, activeQueue.Valid())

			activeProposalID := types.GetProposalIDFromBytes(activeQueue.Value())
			proposal, ok := suite.GovKeeper.GetProposal(ctx, activeProposalID)
			require.True(t, ok)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)

			activeQueue.Close()

			if tc.expeditedPasses {
				// Validator votes YES, letting the expedited proposal pass.
				err = suite.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "metadata")
				require.NoError(t, err)
			}

			// Here the expedited proposal is converted to regular after expiry.
			gov.EndBlocker(ctx, suite.GovKeeper)

			activeQueue = suite.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)

			if tc.expeditedPasses {
				require.False(t, activeQueue.Valid())

				proposal, ok = suite.GovKeeper.GetProposal(ctx, activeProposalID)
				require.True(t, ok)

				require.Equal(t, v1.StatusPassed, proposal.Status)

				submitterEventualBalance := suite.BankKeeper.GetAllBalances(ctx, addrs[0])
				depositorEventualBalance := suite.BankKeeper.GetAllBalances(ctx, addrs[1])

				eventualModuleAccCoins := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

				// Module account has refunded the deposit
				require.Equal(t, initialModuleAccCoins, eventualModuleAccCoins)

				require.Equal(t, submitterInitialBalance, submitterEventualBalance)
				require.Equal(t, depositorInitialBalance, depositorEventualBalance)
				return
			}

			// Expedited proposal should be converted to a regular proposal instead.
			require.True(t, activeQueue.Valid())

			activeProposalID = types.GetProposalIDFromBytes(activeQueue.Value())
			activeQueue.Close()

			proposal, ok = suite.GovKeeper.GetProposal(ctx, activeProposalID)
			require.True(t, ok)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)
			require.False(t, proposal.Expedited)
			require.Equal(t, proposal.VotingStartTime.Add(*params.VotingPeriod), *proposal.VotingEndTime)

			// We also want to make sure that the deposit is not refunded yet and is still present in the module account
			macc = suite.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			intermediateModuleAccCoins := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())
			require.NotEqual(t, initialModuleAccCoins, intermediateModuleAccCoins)

			// Submit proposal deposit + 1 extra top up deposit
			expectedIntermediateMofuleAccCoings := initialModuleAccCoins.Add(proposalCoins...).Add(proposalCoins...)
			require.Equal(t, expectedIntermediateMofuleAccCoings, intermediateModuleAccCoins)

			// block header time at the voting period
			newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)
			ctx = ctx.WithBlockHeader(newHeader)

			inactiveQueue = suite.GovKeeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
			require.False(t, inactiveQueue.Valid())
			inactiveQueue.Close()

			activeQueue = suite.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
			require.True(t, activeQueue.Valid())

			if tc.regularEventuallyPassing {
				// Validator votes YES, letting the converted regular proposal pass.
				err = suite.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "metadata")
				require.NoError(t, err)
			}

			// Here we validate the converted regular proposal
			gov.EndBlocker(ctx, suite.GovKeeper)

			macc = suite.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			eventualModuleAccCoins := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

			submitterEventualBalance := suite.BankKeeper.GetAllBalances(ctx, addrs[0])
			depositorEventualBalance := suite.BankKeeper.GetAllBalances(ctx, addrs[1])

			activeQueue = suite.GovKeeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
			require.False(t, activeQueue.Valid())

			proposal, ok = suite.GovKeeper.GetProposal(ctx, activeProposalID)
			require.True(t, ok)

			if tc.regularEventuallyPassing {
				// Module account has refunded the deposit
				require.Equal(t, initialModuleAccCoins, eventualModuleAccCoins)
				require.Equal(t, submitterInitialBalance, submitterEventualBalance)
				require.Equal(t, depositorInitialBalance, depositorEventualBalance)

				require.Equal(t, v1.StatusPassed, proposal.Status)
				return
			}

			// Not enough votes - module account has returned the deposit
			require.Equal(t, initialModuleAccCoins, eventualModuleAccCoins)
			require.Equal(t, submitterInitialBalance, submitterEventualBalance)
			require.Equal(t, depositorInitialBalance, depositorEventualBalance)

			require.Equal(t, v1.StatusRejected, proposal.Status)
		})
	}
}

func createValidators(t *testing.T, stakingMsgSvr stakingtypes.MsgServer, ctx sdk.Context, addrs []sdk.ValAddress, powerAmt []int64) {
	require.True(t, len(addrs) <= len(pubkeys), "Not enough pubkeys specified at top of file.")

	for i := 0; i < len(addrs); i++ {
		valTokens := sdk.TokensFromConsensusPower(powerAmt[i], sdk.DefaultPowerReduction)
		valCreateMsg, err := stakingtypes.NewMsgCreateValidator(
			addrs[i], pubkeys[i], sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
			TestDescription, TestCommissionRates, math.OneInt(),
		)
		require.NoError(t, err)
		res, err := stakingMsgSvr.CreateValidator(ctx, valCreateMsg)
		require.NoError(t, err)
		require.NotNil(t, res)
	}
}

// With expedited proposal's minimum deposit set higher than the default deposit, we must
// initialize and deposit an amount depositMultiplier times larger
// than the regular min deposit amount.
func getDepositMultiplier(expedited bool) int64 {
	if expedited {
		return v1.DefaultMinExpeditedDepositTokensRatio
	}

	return 1
}
