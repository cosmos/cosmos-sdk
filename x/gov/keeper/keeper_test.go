package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	cdc               codec.Codec
	ctx               sdk.Context
	govKeeper         *keeper.Keeper
	acctKeeper        *govtestutil.MockAccountKeeper
	bankKeeper        *govtestutil.MockBankKeeper
	stakingKeeper     *govtestutil.MockStakingKeeper
	queryClient       v1.QueryClient
	legacyQueryClient v1beta1.QueryClient
	addrs             []sdk.AccAddress
	msgSrvr           v1.MsgServer
	legacyMsgSrvr     v1beta1.MsgServer
}

func (suite *KeeperTestSuite) SetupTest() {
	govKeeper, acctKeeper, bankKeeper, stakingKeeper, encCfg, ctx := setupGovKeeper(suite.T())

	// Populate the gov account with some coins, as the TestProposal we have
	// is a MsgSend from the gov account.
	coins := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100000)))
	err := bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
	suite.NoError(err)
	err = bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, coins)
	suite.NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, encCfg.InterfaceRegistry)
	v1.RegisterQueryServer(queryHelper, suite.govKeeper)
	legacyQueryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, encCfg.InterfaceRegistry)
	v1beta1.RegisterQueryServer(legacyQueryHelper, keeper.NewLegacyQueryServer(suite.govKeeper))
	queryClient := v1.NewQueryClient(queryHelper)
	legacyQueryClient := v1beta1.NewQueryClient(legacyQueryHelper)

	suite.ctx = ctx
	suite.govKeeper = govKeeper
	suite.acctKeeper = acctKeeper
	suite.bankKeeper = bankKeeper
	suite.stakingKeeper = stakingKeeper
	suite.cdc = encCfg.Codec
	suite.queryClient = queryClient
	suite.legacyQueryClient = legacyQueryClient
	suite.msgSrvr = keeper.NewMsgServerImpl(suite.govKeeper)

	govAcct := govKeeper.GetGovernanceAccount(suite.ctx).GetAddress()
	suite.legacyMsgSrvr = keeper.NewLegacyMsgServerImpl(govAcct.String(), suite.msgSrvr)
	suite.addrs = simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(30000000))
}

func setupGovKeeper(t *testing.T) (
	*keeper.Keeper,
	*govtestutil.MockAccountKeeper,
	*govtestutil.MockBankKeeper,
	*govtestutil.MockStakingKeeper,
	moduletestutil.TestEncodingConfig,
	sdk.Context,
) {
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// gomock initializations
	ctrl := gomock.NewController(t)
	acctKeeper := govtestutil.NewMockAccountKeeper(ctrl)
	bankKeeper := govtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := govtestutil.NewMockStakingKeeper(ctrl)
	govKeeper := keeper.NewKeeper(encCfg.Codec, key, acctKeeper, bankKeeper, stakingKeeper, nil, types.DefaultConfig(), "")

	return govKeeper, acctKeeper, bankKeeper, stakingKeeper, encCfg, ctx
}

func TestIncrementProposalNumber(t *testing.T) {
	govKeeper, _, _, _, _, ctx := setupGovKeeper(t)

	tp := TestProposal
	_, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	_, err = govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	_, err = govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	_, err = govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	_, err = govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)
	proposal6, err := govKeeper.SubmitProposal(ctx, tp, "")
	require.NoError(t, err)

	require.Equal(t, uint64(6), proposal6.Id)
}

func TestProposalQueues(t *testing.T) {
	govKeeper, _, _, _, _, ctx := setupGovKeeper(t)

	// create test proposals
	tp := TestProposal
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "")
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
