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

func TestTickExpiredDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)

	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Hash:   app.LastCommitID().Hash,
	})
	require.NoError(t, err)
	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

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

	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Hash:   app.LastCommitID().Hash,
	})
	require.NoError(t, err)
	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

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

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(2) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

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

	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Hash:   app.LastCommitID().Hash,
	})
	require.NoError(t, err)
	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

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

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)

	newDepositMsg := v1.NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)})

	res1, err := govMsgSvr.Deposit(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res1)

	checkActiveProposalsQueue(t, ctx, suite.GovKeeper)
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
			ctx := app.BaseApp.NewContext(false)
			depositMultiplier := getDepositMultiplier(tc.expedited)
			addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens.Mul(math.NewInt(depositMultiplier)))

			SortAddresses(addrs)

			_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
				Height: app.LastBlockHeight() + 1,
				Hash:   app.LastCommitID().Hash,
			})
			require.NoError(t, err)
			govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

			checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)
			checkActiveProposalsQueue(t, ctx, suite.GovKeeper)

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

			params, _ := suite.GovKeeper.Params.Get(ctx)
			votingPeriod := params.VotingPeriod
			if tc.expedited {
				votingPeriod = params.ExpeditedVotingPeriod
			}

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

			if !tc.expedited {
				checkActiveProposalsQueue(t, ctx, suite.GovKeeper)
				return
			}

			// If expedited, it should be converted to a regular proposal instead.
			checkActiveProposalsQueue(t, ctx, suite.GovKeeper)

			proposal, err = suite.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.Nil(t, err)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)
			require.False(t, proposal.Expedited)
			require.Equal(t, proposal.VotingStartTime.Add(*params.VotingPeriod), *proposal.VotingEndTime)
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
			ctx := app.BaseApp.NewContext(false)
			depositMultiplier := getDepositMultiplier(tc.expedited)
			addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens.Mul(math.NewInt(depositMultiplier)))

			SortAddresses(addrs)

			govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)
			stakingMsgSvr := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)

			_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
				Height: app.LastBlockHeight() + 1,
				Hash:   app.LastCommitID().Hash,
			})
			require.NoError(t, err)

			valAddr := sdk.ValAddress(addrs[0])
			proposer := addrs[0]

			createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
			_, err = suite.StakingKeeper.EndBlocker(ctx)
			require.NoError(t, err)
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
			params, _ := suite.GovKeeper.Params.Get(ctx)
			newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)
			ctx = ctx.WithBlockHeader(newHeader)

			err = gov.EndBlocker(ctx, suite.GovKeeper)
			require.NoError(t, err)
			macc = suite.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			require.True(t, suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress()).Equal(initialModuleAccCoins))
		})
	}
}

func TestEndBlockerProposalHandlerFailed(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 1, valTokens)

	SortAddresses(addrs)

	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)

	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
		Hash:   app.LastCommitID().Hash,
	})
	require.NoError(t, err)

	valAddr := sdk.ValAddress(addrs[0])
	proposer := addrs[0]

	createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
	_, err = suite.StakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)
	msg := banktypes.NewMsgSend(authtypes.NewModuleAddress(types.ModuleName), addrs[0], sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000))))
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

	params, _ := suite.GovKeeper.Params.Get(ctx)
	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	// validate that the proposal fails/has been rejected
	err = gov.EndBlocker(ctx, suite.GovKeeper)
	require.NoError(t, err)
	// check proposal events
	events := ctx.EventManager().Events()
	attr, eventOk := events.GetAttributes(types.AttributeKeyProposalLog)
	require.True(t, eventOk)
	require.Contains(t, attr[0].Value, "failed on execution")

	proposal, err = suite.GovKeeper.Proposals.Get(ctx, proposal.Id)
	require.Nil(t, err)
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
			ctx := app.BaseApp.NewContext(false)
			depositMultiplier := getDepositMultiplier(true)
			addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 3, valTokens.Mul(math.NewInt(depositMultiplier)))
			params, err := suite.GovKeeper.Params.Get(ctx)
			require.NoError(t, err)

			SortAddresses(addrs)

			govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)
			stakingMsgSvr := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)

			_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
				Height: app.LastBlockHeight() + 1,
				Hash:   app.LastCommitID().Hash,
			})
			require.NoError(t, err)

			valAddr := sdk.ValAddress(addrs[0])
			proposer := addrs[0]

			// Create a validator so that able to vote on proposal.
			createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
			_, err = suite.StakingKeeper.EndBlocker(ctx)
			require.NoError(t, err)
			checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)
			checkActiveProposalsQueue(t, ctx, suite.GovKeeper)

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

			checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)
			checkActiveProposalsQueue(t, ctx, suite.GovKeeper)

			proposal, err := suite.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.Nil(t, err)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)

			if tc.expeditedPasses {
				// Validator votes YES, letting the expedited proposal pass.
				err = suite.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "metadata")
				require.NoError(t, err)
			}

			// Here the expedited proposal is converted to regular after expiry.
			err = gov.EndBlocker(ctx, suite.GovKeeper)
			require.NoError(t, err)
			if tc.expeditedPasses {
				checkActiveProposalsQueue(t, ctx, suite.GovKeeper)

				proposal, err = suite.GovKeeper.Proposals.Get(ctx, res.ProposalId)
				require.Nil(t, err)

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
			checkActiveProposalsQueue(t, ctx, suite.GovKeeper)
			proposal, err = suite.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.Nil(t, err)
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

			checkInactiveProposalsQueue(t, ctx, suite.GovKeeper)
			checkActiveProposalsQueue(t, ctx, suite.GovKeeper)

			if tc.regularEventuallyPassing {
				// Validator votes YES, letting the converted regular proposal pass.
				err = suite.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "metadata")
				require.NoError(t, err)
			}

			// Here we validate the converted regular proposal
			err = gov.EndBlocker(ctx, suite.GovKeeper)
			require.NoError(t, err)
			macc = suite.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			eventualModuleAccCoins := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

			submitterEventualBalance := suite.BankKeeper.GetAllBalances(ctx, addrs[0])
			depositorEventualBalance := suite.BankKeeper.GetAllBalances(ctx, addrs[1])

			checkActiveProposalsQueue(t, ctx, suite.GovKeeper)

			proposal, err = suite.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.Nil(t, err)

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
	t.Helper()
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

// With expedited proposal's minimum deposit set higher than the default deposit, we must
// initialize and deposit an amount depositMultiplier times larger
// than the regular min deposit amount.
func getDepositMultiplier(expedited bool) int64 {
	if expedited {
		return v1.DefaultMinExpeditedDepositTokensRatio
	}

	return 1
}

func checkActiveProposalsQueue(t *testing.T, ctx sdk.Context, k *keeper.Keeper) {
	t.Helper()
	err := k.ActiveProposalsQueue.Walk(ctx, collections.NewPrefixUntilPairRange[time.Time, uint64](ctx.BlockTime()), func(key collections.Pair[time.Time, uint64], value uint64) (stop bool, err error) {
		return false, err
	})

	require.NoError(t, err)
}

func checkInactiveProposalsQueue(t *testing.T, ctx sdk.Context, k *keeper.Keeper) {
	t.Helper()
	err := k.InactiveProposalsQueue.Walk(ctx, collections.NewPrefixUntilPairRange[time.Time, uint64](ctx.BlockTime()), func(key collections.Pair[time.Time, uint64], value uint64) (stop bool, err error) {
		return false, err
	})

	require.NoError(t, err)
}
