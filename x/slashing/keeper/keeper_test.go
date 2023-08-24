package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	st "cosmossdk.io/api/cosmos/staking/v1beta1"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	addresstypes "github.com/cosmos/cosmos-sdk/types/address"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtestutil "github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

var consAddr = sdk.ConsAddress(sdk.AccAddress([]byte("addr1_______________")))

type KeeperTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	stakingKeeper  *slashingtestutil.MockStakingKeeper
	slashingKeeper slashingkeeper.Keeper
	queryClient    slashingtypes.QueryClient
	msgServer      slashingtypes.MsgServer
	key            *storetypes.KVStoreKey
}

func (s *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(slashingtypes.StoreKey)
	s.key = key
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
	err := s.slashingKeeper.Params.Set(ctx, slashingtestutil.TestParams())
	s.Require().NoError(err)
	slashingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	slashingtypes.RegisterQueryServer(queryHelper, slashingkeeper.NewQuerier(s.slashingKeeper))

	s.queryClient = slashingtypes.NewQueryClient(queryHelper)
	s.msgServer = slashingkeeper.NewMsgServerImpl(s.slashingKeeper)
}

func (s *KeeperTestSuite) TestPubkey() {
	ctx, keeper := s.ctx, s.slashingKeeper
	require := s.Require()

	_, pubKey, addr := testdata.KeyTestPubAddr()
	require.NoError(keeper.AddrPubkeyRelation.Set(ctx, pubKey.Address(), pubKey))

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
		st.Infraction_INFRACTION_UNSPECIFIED,
	).Return(sdkmath.NewInt(0), nil)

	err = s.slashingKeeper.Slash(
		s.ctx,
		consAddr,
		slashFractionDoubleSign,
		sdk.TokensToConsensusPower(sdkmath.NewInt(1), sdk.DefaultPowerReduction),
		s.ctx.BlockHeight(),
	)
	s.Require().NoError(err)
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
		st.Infraction_INFRACTION_DOUBLE_SIGN,
	).Return(sdkmath.NewInt(0), nil)

	err = s.slashingKeeper.SlashWithInfractionReason(
		s.ctx,
		consAddr,
		slashFractionDoubleSign,
		sdk.TokensToConsensusPower(sdkmath.NewInt(1), sdk.DefaultPowerReduction),
		s.ctx.BlockHeight(),
		st.Infraction_INFRACTION_DOUBLE_SIGN,
	)
	s.Require().NoError(err)
	s.stakingKeeper.EXPECT().Jail(s.ctx, consAddr).Return(nil)
	s.Require().NoError(s.slashingKeeper.Jail(s.ctx, consAddr))
}

// validatorMissedBlockBitmapPrefixKey returns the key prefix for a validator's
// missed block bitmap.
func validatorMissedBlockBitmapPrefixKey(v sdk.ConsAddress) []byte {
	validatorMissedBlockBitmapKeyPrefix := []byte{0x02} // Prefix for missed block bitmap

	return append(validatorMissedBlockBitmapKeyPrefix, addresstypes.MustLengthPrefix(v.Bytes())...)
}

// func (s *KeeperTestSuite) TestValidatorMissedBlockBMMigrationToColls() {
// 	s.SetupTest()

// 	err := testutil.DiffCollectionsMigration(
// 		s.ctx,
// 		s.key,
// 		100,
// 		func(i int64) {
// 			s.ctx.KVStore(s.key).Set(getLastValidatorPowerKey(valAddrs[i]), bz)
// 		},
// 		"6cd9b908445fbe0b280b82cac51758cdb125882674a91d348b690dac4b7055cb",
// 	)
// 	s.Require().NoError(err)

// 	err = testutil.DiffCollectionsMigration(
// 		s.ctx,
// 		s.key,
// 		100,
// 		func(i int64) {
// 			err := s.stakingKeeper.LastValidatorPower.Set(s.ctx, valAddrs[i], bz)
// 			s.Require().NoError(err)
// 		},
// 		"6cd9b908445fbe0b280b82cac51758cdb125882674a91d348b690dac4b7055cb",
// 	)
// 	s.Require().NoError(err)
// }

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
