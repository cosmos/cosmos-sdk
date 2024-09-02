package gov_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/gov/keeper"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestUnregisteredProposal_InactiveProposalFails(t *testing.T) {
	suite := createTestSuite(t)
	ctx := suite.app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)
	addr0Str, err := suite.AccountKeeper.AddressCodec().BytesToString(addrs[0])
	require.NoError(t, err)

	// manually set proposal in store
	startTime, endTime := time.Now().Add(-4*time.Hour), ctx.BlockHeader().Time
	proposal, err := v1.NewProposal([]sdk.Msg{
		&v1.Proposal{}, // invalid proposal message
	}, 1, startTime, startTime, "", "Unsupported proposal", "Unsupported proposal", addr0Str, v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	require.NoError(t, err)

	err = suite.GovKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	require.NoError(t, err)

	// manually set proposal in inactive proposal queue
	err = suite.GovKeeper.InactiveProposalsQueue.Set(ctx, collections.Join(endTime, proposal.Id), proposal.Id)
	require.NoError(t, err)

	err = suite.GovKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	_, err = suite.GovKeeper.Proposals.Get(ctx, proposal.Id)
	require.Error(t, err, collections.ErrNotFound)
}

func TestUnregisteredProposal_ActiveProposalFails(t *testing.T) {
	suite := createTestSuite(t)
	ctx := suite.app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)
	addr0Str, err := suite.AccountKeeper.AddressCodec().BytesToString(addrs[0])
	require.NoError(t, err)
	// manually set proposal in store
	startTime, endTime := time.Now().Add(-4*time.Hour), ctx.BlockHeader().Time
	proposal, err := v1.NewProposal([]sdk.Msg{
		&v1.Proposal{}, // invalid proposal message
	}, 1, startTime, startTime, "", "Unsupported proposal", "Unsupported proposal", addr0Str, v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	require.NoError(t, err)
	proposal.Status = v1.StatusVotingPeriod
	proposal.VotingEndTime = &endTime

	err = suite.GovKeeper.Proposals.Set(ctx, proposal.Id, proposal)
	require.NoError(t, err)

	// manually set proposal in active proposal queue
	err = suite.GovKeeper.ActiveProposalsQueue.Set(ctx, collections.Join(endTime, proposal.Id), proposal.Id)
	require.NoError(t, err)

	err = suite.GovKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	p, err := suite.GovKeeper.Proposals.Get(ctx, proposal.Id)
	require.NoError(t, err)
	require.Equal(t, v1.StatusFailed, p.Status)
}

func TestTickExpiredDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.app
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)

	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
		v1.ProposalType_PROPOSAL_TYPE_STANDARD,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	newHeader := ctx.HeaderInfo()
	newHeader.Time = ctx.HeaderInfo().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithHeaderInfo(newHeader)

	params, _ := suite.GovKeeper.Params.Get(ctx)
	newHeader = ctx.HeaderInfo()
	newHeader.Time = ctx.HeaderInfo().Time.Add(*params.MaxDepositPeriod)
	ctx = ctx.WithHeaderInfo(newHeader)

	err = suite.GovKeeper.EndBlocker(ctx)
	require.NoError(t, err)
}

func TestTickMultipleExpiredDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.app
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)
	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
		v1.ProposalType_PROPOSAL_TYPE_STANDARD,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	newHeader := ctx.HeaderInfo()
	newHeader.Time = ctx.HeaderInfo().Time.Add(time.Duration(2) * time.Second)
	ctx = ctx.WithHeaderInfo(newHeader)

	newProposalMsg2, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
		v1.ProposalType_PROPOSAL_TYPE_STANDARD,
	)
	require.NoError(t, err)

	res, err = govMsgSvr.SubmitProposal(ctx, newProposalMsg2)
	require.NoError(t, err)
	require.NotNil(t, res)

	newHeader = ctx.HeaderInfo()
	params, _ := suite.GovKeeper.Params.Get(ctx)
	newHeader.Time = ctx.HeaderInfo().Time.Add(*params.MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithHeaderInfo(newHeader)

	require.NoError(t, suite.GovKeeper.EndBlocker(ctx))

	newHeader = ctx.HeaderInfo()
	newHeader.Time = ctx.HeaderInfo().Time.Add(time.Duration(5) * time.Second)
	ctx = ctx.WithHeaderInfo(newHeader)
	require.NoError(t, suite.GovKeeper.EndBlocker(ctx))
}

func TestTickPassedDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.app
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)
	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
		v1.ProposalType_PROPOSAL_TYPE_STANDARD,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	proposalID := res.ProposalId

	newHeader := ctx.HeaderInfo()
	newHeader.Time = ctx.HeaderInfo().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithHeaderInfo(newHeader)

	addr1Str, err := suite.AccountKeeper.AddressCodec().BytesToString(addrs[1])
	require.NoError(t, err)
	newDepositMsg := v1.NewMsgDeposit(addr1Str, proposalID, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)})

	res1, err := govMsgSvr.Deposit(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res1)
}

func TestProposalDepositRefundFailEndBlocker(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.app
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens)
	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

	depositMultiplier := getDepositMultiplier(v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, suite.StakingKeeper.TokensFromConsensusPower(ctx, 15*depositMultiplier))}

	// create a proposal that empties the gov module account
	// which will cause the proposal deposit refund to fail
	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{},
		proposalCoins,
		addrs[0].String(),
		"metadata",
		"proposal",
		"description of proposal",
		v1.ProposalType_PROPOSAL_TYPE_STANDARD,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	proposal, err := suite.GovKeeper.Proposals.Get(ctx, res.ProposalId)
	require.NoError(t, err)
	require.Equal(t, v1.StatusVotingPeriod, proposal.Status)

	proposalID := res.ProposalId
	err = suite.GovKeeper.AddVote(ctx, proposalID, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "metadata")
	require.NoError(t, err)

	// empty the gov module account before the proposal ends
	err = suite.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addrs[1], proposalCoins)
	require.NoError(t, err)

	// fast forward to the end of the voting period
	newHeader := ctx.HeaderInfo()
	newHeader.Time = proposal.VotingEndTime.Add(time.Duration(100) * time.Second)
	ctx = ctx.WithHeaderInfo(newHeader)

	err = suite.GovKeeper.EndBlocker(ctx)
	require.NoError(t, err) // no error, means does not halt the chain

	events := ctx.EventManager().Events()
	attr, ok := events.GetAttributes(types.AttributeKeyProposalDepositError)
	require.True(t, ok)
	require.Contains(t, attr[0].Value, "failed to refund or burn deposits")
}

func TestTickPassedVotingPeriod(t *testing.T) {
	testcases := []struct {
		name         string
		proposalType v1.ProposalType
	}{
		{
			name: "regular - deleted",
		},
		{
			name:         "expedited - converted to regular",
			proposalType: v1.ProposalType_PROPOSAL_TYPE_EXPEDITED,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			suite := createTestSuite(t)
			app := suite.app
			ctx := app.BaseApp.NewContext(false)
			depositMultiplier := getDepositMultiplier(tc.proposalType)
			addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens.Mul(math.NewInt(depositMultiplier)))

			SortAddresses(addrs)
			govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)

			proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, suite.StakingKeeper.TokensFromConsensusPower(ctx, 5*depositMultiplier))}
			newProposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{mkTestLegacyContent(t)}, proposalCoins, addrs[0].String(), "", "Proposal", "description of proposal", tc.proposalType)
			require.NoError(t, err)

			res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
			require.NoError(t, err)
			require.NotNil(t, res)

			proposalID := res.ProposalId

			newHeader := ctx.HeaderInfo()
			newHeader.Time = ctx.HeaderInfo().Time.Add(time.Duration(1) * time.Second)
			ctx = ctx.WithHeaderInfo(newHeader)

			addr1Str, err := suite.AccountKeeper.AddressCodec().BytesToString(addrs[1])
			require.NoError(t, err)
			newDepositMsg := v1.NewMsgDeposit(addr1Str, proposalID, proposalCoins)

			res1, err := govMsgSvr.Deposit(ctx, newDepositMsg)
			require.NoError(t, err)
			require.NotNil(t, res1)

			params, _ := suite.GovKeeper.Params.Get(ctx)
			votingPeriod := params.VotingPeriod
			if tc.proposalType == v1.ProposalType_PROPOSAL_TYPE_EXPEDITED {
				votingPeriod = params.ExpeditedVotingPeriod
			}

			newHeader = ctx.HeaderInfo()
			newHeader.Time = ctx.HeaderInfo().Time.Add(*params.MaxDepositPeriod).Add(*votingPeriod)
			ctx = ctx.WithHeaderInfo(newHeader)

			proposal, err := suite.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.NoError(t, err)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)

			err = suite.GovKeeper.EndBlocker(ctx)
			require.NoError(t, err)

			if tc.proposalType != v1.ProposalType_PROPOSAL_TYPE_EXPEDITED {
				return
			}

			// If expedited, it should be converted to a regular proposal instead.
			proposal, err = suite.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.Nil(t, err)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)
			require.False(t, proposal.ProposalType == v1.ProposalType_PROPOSAL_TYPE_EXPEDITED)
			require.Equal(t, proposal.VotingStartTime.Add(*params.VotingPeriod), *proposal.VotingEndTime)
		})
	}
}

func TestProposalPassedEndblocker(t *testing.T) {
	testcases := []struct {
		name         string
		proposalType v1.ProposalType
	}{
		{
			name:         "regular",
			proposalType: v1.ProposalType_PROPOSAL_TYPE_STANDARD,
		},
		{
			name:         "expedited",
			proposalType: v1.ProposalType_PROPOSAL_TYPE_EXPEDITED,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			suite := createTestSuite(t)
			app := suite.app
			ctx := app.BaseApp.NewContext(false)
			depositMultiplier := getDepositMultiplier(tc.proposalType)
			addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 10, valTokens.Mul(math.NewInt(depositMultiplier)))

			SortAddresses(addrs)

			govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)
			stakingMsgSvr := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)
			valAddr := sdk.ValAddress(addrs[0])
			proposer := addrs[0]
			acc := suite.AccountKeeper.NewAccountWithAddress(ctx, addrs[0])
			suite.AccountKeeper.SetAccount(ctx, acc)

			createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
			_, err := suite.StakingKeeper.EndBlocker(ctx)
			require.NoError(t, err)
			macc := suite.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			initialModuleAccCoins := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

			proposal, err := suite.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, "", "title", "summary", proposer, tc.proposalType)
			require.NoError(t, err)

			proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, suite.StakingKeeper.TokensFromConsensusPower(ctx, 10*depositMultiplier))}
			addr0Str, err := suite.AccountKeeper.AddressCodec().BytesToString(addrs[0])
			require.NoError(t, err)
			newDepositMsg := v1.NewMsgDeposit(addr0Str, proposal.Id, proposalCoins)

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

			newHeader := ctx.HeaderInfo()
			params, _ := suite.GovKeeper.Params.Get(ctx)
			newHeader.Time = ctx.HeaderInfo().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)
			ctx = ctx.WithHeaderInfo(newHeader)

			err = suite.GovKeeper.EndBlocker(ctx)
			require.NoError(t, err)
			macc = suite.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			require.True(t, suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress()).Equal(initialModuleAccCoins))
		})
	}
}

func TestEndBlockerProposalHandlerFailed(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.app
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 1, valTokens)

	SortAddresses(addrs)

	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)

	valAddr := sdk.ValAddress(addrs[0])
	proposer := addrs[0]

	ac := addresscodec.NewBech32Codec("cosmos")
	addrStr, err := ac.BytesToString(authtypes.NewModuleAddress(types.ModuleName))
	require.NoError(t, err)
	toAddrStr, err := ac.BytesToString(addrs[0])
	require.NoError(t, err)

	acc := suite.AccountKeeper.NewAccountWithAddress(ctx, addrs[0])
	suite.AccountKeeper.SetAccount(ctx, acc)

	createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
	_, err = suite.StakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)
	msg := banktypes.NewMsgSend(addrStr, toAddrStr, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000))))
	proposal, err := suite.GovKeeper.SubmitProposal(ctx, []sdk.Msg{msg}, "", "title", "summary", proposer, v1.ProposalType_PROPOSAL_TYPE_STANDARD)
	require.NoError(t, err)

	proposalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, suite.StakingKeeper.TokensFromConsensusPower(ctx, 10)))
	addr0Str, err := suite.AccountKeeper.AddressCodec().BytesToString(addrs[0])
	require.NoError(t, err)
	newDepositMsg := v1.NewMsgDeposit(addr0Str, proposal.Id, proposalCoins)

	govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)
	res, err := govMsgSvr.Deposit(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	err = suite.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "")
	require.NoError(t, err)

	params, _ := suite.GovKeeper.Params.Get(ctx)
	newHeader := ctx.HeaderInfo()
	newHeader.Time = ctx.HeaderInfo().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)
	ctx = ctx.WithHeaderInfo(newHeader)

	// validate that the proposal fails/has been rejected
	err = suite.GovKeeper.EndBlocker(ctx)
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
			app := suite.app
			ctx := app.BaseApp.NewContext(false)
			depositMultiplier := getDepositMultiplier(v1.ProposalType_PROPOSAL_TYPE_EXPEDITED)
			addrs := simtestutil.AddTestAddrs(suite.BankKeeper, suite.StakingKeeper, ctx, 3, valTokens.Mul(math.NewInt(depositMultiplier)))
			params, err := suite.GovKeeper.Params.Get(ctx)
			require.NoError(t, err)

			SortAddresses(addrs)

			govMsgSvr := keeper.NewMsgServerImpl(suite.GovKeeper)
			stakingMsgSvr := stakingkeeper.NewMsgServerImpl(suite.StakingKeeper)
			valAddr := sdk.ValAddress(addrs[0])
			proposer := addrs[0]

			acc := suite.AccountKeeper.NewAccountWithAddress(ctx, addrs[0])
			suite.AccountKeeper.SetAccount(ctx, acc)
			// Create a validator so that able to vote on proposal.
			createValidators(t, stakingMsgSvr, ctx, []sdk.ValAddress{valAddr}, []int64{10})
			_, err = suite.StakingKeeper.EndBlocker(ctx)
			require.NoError(t, err)

			macc := suite.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			initialModuleAccCoins := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

			submitterInitialBalance := suite.BankKeeper.GetAllBalances(ctx, addrs[0])
			depositorInitialBalance := suite.BankKeeper.GetAllBalances(ctx, addrs[1])

			proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, suite.StakingKeeper.TokensFromConsensusPower(ctx, 5*depositMultiplier))}
			newProposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{}, proposalCoins, proposer.String(), "metadata", "title", "summary", v1.ProposalType_PROPOSAL_TYPE_EXPEDITED)
			require.NoError(t, err)

			res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
			require.NoError(t, err)
			require.NotNil(t, res)

			proposalID := res.ProposalId

			newHeader := ctx.HeaderInfo()
			newHeader.Time = ctx.HeaderInfo().Time.Add(time.Duration(1) * time.Second)
			ctx = ctx.WithHeaderInfo(newHeader)

			addr1Str, err := suite.AccountKeeper.AddressCodec().BytesToString(addrs[1])
			require.NoError(t, err)
			newDepositMsg := v1.NewMsgDeposit(addr1Str, proposalID, proposalCoins)

			res1, err := govMsgSvr.Deposit(ctx, newDepositMsg)
			require.NoError(t, err)
			require.NotNil(t, res1)

			newHeader = ctx.HeaderInfo()
			newHeader.Time = ctx.HeaderInfo().Time.Add(*params.MaxDepositPeriod).Add(*params.ExpeditedVotingPeriod)
			ctx = ctx.WithHeaderInfo(newHeader)

			proposal, err := suite.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.Nil(t, err)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)

			if tc.expeditedPasses {
				// Validator votes YES, letting the expedited proposal pass.
				err = suite.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "metadata")
				require.NoError(t, err)
			}

			// Here the expedited proposal is converted to regular after expiry.
			err = suite.GovKeeper.EndBlocker(ctx)
			require.NoError(t, err)
			if tc.expeditedPasses {
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
			proposal, err = suite.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.Nil(t, err)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)
			require.False(t, proposal.ProposalType == v1.ProposalType_PROPOSAL_TYPE_EXPEDITED)
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
			newHeader.Time = ctx.HeaderInfo().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)
			ctx = ctx.WithHeaderInfo(newHeader)

			if tc.regularEventuallyPassing {
				// Validator votes YES, letting the converted regular proposal pass.
				err = suite.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "metadata")
				require.NoError(t, err)
			}

			// Here we validate the converted regular proposal
			err = suite.GovKeeper.EndBlocker(ctx)
			require.NoError(t, err)
			macc = suite.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			eventualModuleAccCoins := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

			submitterEventualBalance := suite.BankKeeper.GetAllBalances(ctx, addrs[0])
			depositorEventualBalance := suite.BankKeeper.GetAllBalances(ctx, addrs[1])

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
func getDepositMultiplier(proposalType v1.ProposalType) int64 {
	switch proposalType {
	case v1.ProposalType_PROPOSAL_TYPE_EXPEDITED:
		return v1.DefaultMinExpeditedDepositTokensRatio
	default:
		return 1
	}
}
