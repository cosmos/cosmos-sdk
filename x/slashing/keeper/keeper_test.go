package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v2"
	cmttime "github.com/cometbft/cometbft/v2/types/time"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtestutil "github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var consAddr = sdk.ConsAddress(sdk.AccAddress([]byte("addr1_______________")))

type KeeperTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	stakingKeeper  *slashingtestutil.MockStakingKeeper
	slashingKeeper slashingkeeper.Keeper
	queryClient    slashingtypes.QueryClient
	msgServer      slashingtypes.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(slashingtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// gomock initializations
	ctrl := gomock.NewController(s.T())
	s.stakingKeeper = slashingtestutil.NewMockStakingKeeper(ctrl)
	s.stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec("cosmosvaloper")).AnyTimes()
	s.stakingKeeper.EXPECT().ConsensusAddressCodec().Return(address.NewBech32Codec("cosmosvalcons")).AnyTimes()

	s.ctx = ctx
	s.slashingKeeper = slashingkeeper.NewKeeper(
		encCfg.Codec,
		encCfg.Amino,
		storeService,
		s.stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	// set test params
	s.Require().NoError(s.slashingKeeper.SetParams(ctx, slashingtestutil.TestParams()))

	slashingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	slashingtypes.RegisterQueryServer(queryHelper, s.slashingKeeper)

	s.queryClient = slashingtypes.NewQueryClient(queryHelper)
	s.msgServer = slashingkeeper.NewMsgServerImpl(s.slashingKeeper)
}

func (s *KeeperTestSuite) TestPubkey() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	_, pubKey, addr := testdata.KeyTestPubAddr()
	require.NoError(keeper.AddPubkey(ctx, pubKey))

	expectedPubKey, err := keeper.GetPubkey(ctx, addr.Bytes())
	require.NoError(err)
	require.Equal(pubKey, expectedPubKey)
}

func (s *KeeperTestSuite) TestJailAndSlash() {
	slashFractionDoubleSign, err := s.slashingKeeper.SlashFractionDoubleSign(s.ctx)
	s.Require().NoError(err)

	s.stakingKeeper.EXPECT().SlashWithInfractionReason(s.ctx,
		consAddr,
		s.ctx.BlockHeight(),
		sdk.TokensToConsensusPower(sdkmath.NewInt(1), sdk.DefaultPowerReduction),
		slashFractionDoubleSign,
		stakingtypes.Infraction_INFRACTION_UNSPECIFIED,
	).Return(sdkmath.NewInt(0), nil)

	s.Require().NoError(s.slashingKeeper.Slash(
		s.ctx,
		consAddr,
		slashFractionDoubleSign,
		sdk.TokensToConsensusPower(sdkmath.NewInt(1), sdk.DefaultPowerReduction),
		s.ctx.BlockHeight(),
	))

	s.stakingKeeper.EXPECT().Jail(s.ctx, consAddr).Return(nil)
	s.Require().NoError(s.slashingKeeper.Jail(s.ctx, consAddr))
}

func (s *KeeperTestSuite) TestJailAndSlashWithInfractionReason() {
	slashFractionDoubleSign, err := s.slashingKeeper.SlashFractionDoubleSign(s.ctx)
	s.Require().NoError(err)

	s.stakingKeeper.EXPECT().SlashWithInfractionReason(s.ctx,
		consAddr,
		s.ctx.BlockHeight(),
		sdk.TokensToConsensusPower(sdkmath.NewInt(1), sdk.DefaultPowerReduction),
		slashFractionDoubleSign,
		stakingtypes.Infraction_INFRACTION_DOUBLE_SIGN,
	).Return(sdkmath.NewInt(0), nil)

	s.Require().NoError(s.slashingKeeper.SlashWithInfractionReason(
		s.ctx,
		consAddr,
		slashFractionDoubleSign,
		sdk.TokensToConsensusPower(sdkmath.NewInt(1), sdk.DefaultPowerReduction),
		s.ctx.BlockHeight(),
		stakingtypes.Infraction_INFRACTION_DOUBLE_SIGN,
	))

	s.stakingKeeper.EXPECT().Jail(s.ctx, consAddr).Return(nil)
	s.Require().NoError(s.slashingKeeper.Jail(s.ctx, consAddr))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
