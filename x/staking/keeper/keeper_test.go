package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
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
}

func (s *KeeperTestSuite) SetupTest() {
	require := s.Require()
	key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	s.key = key
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.key = key
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

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

// getREDByValDstIndexKey creates the index-key for a redelegation, stored by destination-validator-index
// VALUE: none (key rearrangement used)
func getREDByValDstIndexKey(delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) []byte {
	REDSToValsDstKey := getREDsToValDstIndexKey(valDstAddr)
	offset := len(REDSToValsDstKey)

	// key is of the form REDSToValsDstKey || delAddrLen (1 byte) || delAddr || valSrcAddrLen (1 byte) || valSrcAddr
	key := make([]byte, offset+2+len(delAddr)+len(valSrcAddr))
	copy(key[0:offset], REDSToValsDstKey)
	key[offset] = byte(len(delAddr))
	copy(key[offset+1:offset+1+len(delAddr)], delAddr.Bytes())
	key[offset+1+len(delAddr)] = byte(len(valSrcAddr))
	copy(key[offset+2+len(delAddr):], valSrcAddr.Bytes())

	return key
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

// GetREDsToValDstIndexKey returns a key prefix for indexing a redelegation to a
// destination (target) validator.
func getREDsToValDstIndexKey(valDstAddr sdk.ValAddress) []byte {
	redelegationByValDstIndexKey := []byte{0x36}
	return append(redelegationByValDstIndexKey, addresstypes.MustLengthPrefix(valDstAddr)...)
}

// GetREDsFromValSrcIndexKey returns a key prefix for indexing a redelegation to
// a source validator.
func getREDsFromValSrcIndexKey(valSrcAddr sdk.ValAddress) []byte {
	redelegationByValSrcIndexKey := []byte{0x35}
	return append(redelegationByValSrcIndexKey, addresstypes.MustLengthPrefix(valSrcAddr)...)
}

func (s *KeeperTestSuite) TestSrcRedelegationsMigrationToColls() {
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

func (s *KeeperTestSuite) TestDstRedelegationsMigrationToColls() {
	s.SetupTest()

	addrs, valAddrs := createValAddrs(101)

	err := testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			// legacy method to set in the state
			s.ctx.KVStore(s.key).Set(getREDByValDstIndexKey(addrs[i], valAddrs[i], valAddrs[i+1]), []byte{})
		},
		"4beb77994beff3c8ad9cecca9ee3a74fb551356250f0b8bd3936c4e4f506443b", // this hash obtained when ran this test in main branch
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			// using collections
			err := s.stakingKeeper.RedelegationsByValDst.Set(s.ctx, collections.Join3(valAddrs[i+1].Bytes(), addrs[i].Bytes(), valAddrs[i].Bytes()), []byte{})
			s.Require().NoError(err)
		},
		"4beb77994beff3c8ad9cecca9ee3a74fb551356250f0b8bd3936c4e4f506443b",
	)

	s.Require().NoError(err)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestDiffCollsMigration() {
	s.SetupTest()

	err := testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			date := time.Date(2023, 8, 21, 14, 33, 1, 0, &time.Location{})
			err := s.stakingKeeper.SetRedelegationQueueTimeSlice(s.ctx, date, nil)
			s.Require().NoError(err)
		},
		"035e246f9d0bf0aa3dfeb88acf1665684168256d8afb742ae065872d6334f6d6",
	)
	s.Require().NoError(err)
}
