package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	addresstypes "github.com/cosmos/cosmos-sdk/types/address"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	bondedAcc    = authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName)
	notBondedAcc = authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName)
	PKs          = simtestutil.CreateTestPubKeys(500)
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	stakingKeeper *stakingkeeper.Keeper
	bankKeeper    *stakingtestutil.MockBankKeeper
	accountKeeper *stakingtestutil.MockAccountKeeper
	queryClient   stakingtypes.QueryClient
	msgServer     stakingtypes.MsgServer
	key           *storetypes.KVStoreKey
	cdc           codec.Codec
}

func (s *KeeperTestSuite) SetupTest() {
	require := s.Require()
	key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.key = key
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()
	s.cdc = encCfg.Codec

	ctrl := gomock.NewController(s.T())
	accountKeeper := stakingtestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress())
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress())
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	bankKeeper := stakingtestutil.NewMockBankKeeper(ctrl)

	keeper := stakingkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(stakingtypes.GovModuleName).String(),
		address.NewBech32Codec("cosmosvaloper"),
		address.NewBech32Codec("cosmosvalcons"),
	)
	require.NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

	s.ctx = ctx
	s.stakingKeeper = keeper
	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper

	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	stakingtypes.RegisterQueryServer(queryHelper, stakingkeeper.Querier{Keeper: keeper})
	s.queryClient = stakingtypes.NewQueryClient(queryHelper)
	s.msgServer = stakingkeeper.NewMsgServerImpl(keeper)
}

func (s *KeeperTestSuite) TestParams() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	expParams := stakingtypes.DefaultParams()
	// check that the empty keeper loads the default
	resParams, err := keeper.GetParams(ctx)
	require.NoError(err)
	require.Equal(expParams, resParams)

	expParams.MaxValidators = 555
	expParams.MaxEntries = 111
	require.NoError(keeper.SetParams(ctx, expParams))
	resParams, err = keeper.GetParams(ctx)
	require.NoError(err)
	require.True(expParams.Equal(resParams))
}

func (s *KeeperTestSuite) TestLastTotalPower() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	expTotalPower := math.NewInt(10 ^ 9)
	require.NoError(keeper.LastTotalPower.Set(ctx, expTotalPower))
	resTotalPower, err := keeper.LastTotalPower.Get(ctx)
	require.NoError(err)
	require.True(expTotalPower.Equal(resTotalPower))
}

// GetREDByValSrcIndexKey creates the index-key for a redelegation, stored by source-validator-index
// VALUE: none (key rearrangement used)
func getREDByValSrcIndexKey(delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) []byte {
	REDSFromValsSrcKey := getREDsFromValSrcIndexKey(valSrcAddr)
	offset := len(REDSFromValsSrcKey)

	// key is of the form REDSFromValsSrcKey || delAddrLen (1 byte) || delAddr || valDstAddrLen (1 byte) || valDstAddr
	key := make([]byte, offset+2+len(delAddr)+len(valDstAddr))
	copy(key[0:offset], REDSFromValsSrcKey)
	key[offset] = byte(len(delAddr))
	copy(key[offset+1:offset+1+len(delAddr)], delAddr.Bytes())
	key[offset+1+len(delAddr)] = byte(len(valDstAddr))
	copy(key[offset+2+len(delAddr):], valDstAddr.Bytes())

	return key
}

// GetREDsFromValSrcIndexKey returns a key prefix for indexing a redelegation to
// a source validator.
func getREDsFromValSrcIndexKey(valSrcAddr sdk.ValAddress) []byte {
	redelegationByValSrcIndexKey := []byte{0x35}
	return append(redelegationByValSrcIndexKey, addresstypes.MustLengthPrefix(valSrcAddr)...)
}

func getLastValidatorPowerKey(operator sdk.ValAddress) []byte {
	lastValidatorPowerKey := []byte{0x11}
	return append(lastValidatorPowerKey, addresstypes.MustLengthPrefix(operator)...)
}

func (s *KeeperTestSuite) TestLastTotalPowerMigrationToColls() {
	s.SetupTest()

	_, valAddrs := createValAddrs(101)

	bz, err := s.cdc.Marshal(&gogotypes.Int64Value{Value: 1})
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			s.ctx.KVStore(s.key).Set(getLastValidatorPowerKey(valAddrs[i]), bz)
		},
		"6cd9b908445fbe0b280b82cac51758cdb125882674a91d348b690dac4b7055cb",
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			err := s.stakingKeeper.LastValidatorPower.Set(s.ctx, valAddrs[i], bz)
			s.Require().NoError(err)
		},
		"6cd9b908445fbe0b280b82cac51758cdb125882674a91d348b690dac4b7055cb",
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestRedelegationsMigrationToColls() {
	s.SetupTest()

	addrs, valAddrs := createValAddrs(101)

	err := testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			// legacy method to set in the state
			s.ctx.KVStore(s.key).Set(getREDByValSrcIndexKey(addrs[i], valAddrs[i], valAddrs[i+1]), []byte{})
		},
		"cb7b7086b1e03add24f85f894531fb36b3b9746f2e661e1640ec528a4f23a3d9",
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			// using collections
			err := s.stakingKeeper.RedelegationsByValSrc.Set(s.ctx, collections.Join3(valAddrs[i].Bytes(), addrs[i].Bytes(), valAddrs[i+1].Bytes()), []byte{})
			s.Require().NoError(err)
		},
		"cb7b7086b1e03add24f85f894531fb36b3b9746f2e661e1640ec528a4f23a3d9",
	)

	s.Require().NoError(err)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
