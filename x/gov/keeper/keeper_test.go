package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

var address1 = "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r"

type KeeperTestSuite struct {
	suite.Suite

	cdc               codec.Codec
	ctx               sdk.Context
	govKeeper         *keeper.Keeper
	acctKeeper        *govtestutil.MockAccountKeeper
	bankKeeper        *govtestutil.MockBankKeeper
	stakingKeeper     *govtestutil.MockStakingKeeper
	distKeeper        *govtestutil.MockDistributionKeeper
	queryClient       v1.QueryClient
	legacyQueryClient v1beta1.QueryClient
	addrs             []sdk.AccAddress
	msgSrvr           v1.MsgServer
	legacyMsgSrvr     v1beta1.MsgServer
}

func (suite *KeeperTestSuite) SetupSuite() {
	suite.reset()
}

func (suite *KeeperTestSuite) reset() {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, distKeeper, encCfg, ctx := setupGovKeeper(suite.T())

	// Populate the gov account with some coins, as the TestProposal we have
	// is a MsgSend from the gov account.
	coins := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(100000)))
	err := bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
	suite.NoError(err)
	err = bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, coins)
	suite.NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	v1.RegisterQueryServer(queryHelper, govKeeper)
	legacyQueryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	v1beta1.RegisterQueryServer(legacyQueryHelper, keeper.NewLegacyQueryServer(govKeeper))
	queryClient := v1.NewQueryClient(queryHelper)
	legacyQueryClient := v1beta1.NewQueryClient(legacyQueryHelper)

	suite.ctx = ctx
	suite.govKeeper = govKeeper
	suite.acctKeeper = acctKeeper
	suite.bankKeeper = bankKeeper
	suite.stakingKeeper = stakingKeeper
	suite.distKeeper = distKeeper
	suite.cdc = encCfg.Codec
	suite.queryClient = queryClient
	suite.legacyQueryClient = legacyQueryClient
	suite.msgSrvr = keeper.NewMsgServerImpl(suite.govKeeper)

	suite.legacyMsgSrvr = keeper.NewLegacyMsgServerImpl(govAcct.String(), suite.msgSrvr)
	suite.addrs = simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 3, sdkmath.NewInt(30000000))

	for _, addr := range suite.addrs {
		suite.acctKeeper.EXPECT().BytesToString(addr).Return(addr.String(), nil).AnyTimes()
		suite.acctKeeper.EXPECT().StringToBytes(addr.String()).Return(addr, nil).AnyTimes()
	}
}

func TestIncrementProposalNumber(t *testing.T) {
	govKeeper, authKeeper, _, _, _, _, ctx := setupGovKeeper(t)

	ac := address.NewBech32Codec("cosmos")
	addrBz, err := ac.StringToBytes(address1)
	require.NoError(t, err)
	authKeeper.EXPECT().StringToBytes(address1).Return(addrBz, nil).AnyTimes()
	authKeeper.EXPECT().BytesToString(addrBz).Return(address1, nil).AnyTimes()

	tp := TestProposal
	_, err = govKeeper.SubmitProposal(ctx, tp, "", "test", "summary", addrBz, false)
	require.NoError(t, err)
	_, err = govKeeper.SubmitProposal(ctx, tp, "", "test", "summary", addrBz, false)
	require.NoError(t, err)
	_, err = govKeeper.SubmitProposal(ctx, tp, "", "test", "summary", addrBz, true)
	require.NoError(t, err)
	_, err = govKeeper.SubmitProposal(ctx, tp, "", "test", "summary", addrBz, true)
	require.NoError(t, err)
	_, err = govKeeper.SubmitProposal(ctx, tp, "", "test", "summary", addrBz, false)
	require.NoError(t, err)
	proposal6, err := govKeeper.SubmitProposal(ctx, tp, "", "test", "summary", addrBz, false)
	require.NoError(t, err)

	require.Equal(t, uint64(6), proposal6.Id)
}

func TestProposalQueues(t *testing.T) {
	govKeeper, authKeeper, _, _, _, _, ctx := setupGovKeeper(t)

	ac := address.NewBech32Codec("cosmos")
	addrBz, err := ac.StringToBytes(address1)
	require.NoError(t, err)
	authKeeper.EXPECT().StringToBytes(address1).Return(addrBz, nil).AnyTimes()
	authKeeper.EXPECT().BytesToString(addrBz).Return(address1, nil).AnyTimes()

	// create test proposals
	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "", "test", "summary", addrBz, false)
	require.NoError(t, err)

	inactiveIterator := govKeeper.InactiveProposalQueueIterator(ctx, *proposal.DepositEndTime)
	require.True(t, inactiveIterator.Valid())

	proposalID := types.GetProposalIDFromBytes(inactiveIterator.Value())
	require.Equal(t, proposalID, proposal.Id)
	inactiveIterator.Close()

	govKeeper.ActivateVotingPeriod(ctx, proposal)

	proposal, ok := govKeeper.GetProposal(ctx, proposal.Id)
	require.True(t, ok)

	activeIterator := govKeeper.ActiveProposalQueueIterator(ctx, *proposal.VotingEndTime)
	require.True(t, activeIterator.Valid())

	proposalID, _ = types.SplitActiveProposalQueueKey(activeIterator.Key())
	require.Equal(t, proposalID, proposal.Id)

	activeIterator.Close()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
