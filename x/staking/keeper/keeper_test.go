package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
	s.key = key
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

// getUBDKey creates the key for an unbonding delegation by delegator and validator addr
// VALUE: staking/UnbondingDelegation
func getUBDKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	unbondingDelegationKey := []byte{0x32}
	return append(append(unbondingDelegationKey, addresstypes.MustLengthPrefix(delAddr)...), addresstypes.MustLengthPrefix(valAddr)...)
}

// getUBDByValIndexKey creates the index-key for an unbonding delegation, stored by validator-index
// VALUE: none (key rearrangement used)
func getUBDByValIndexKey(delAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	unbondingDelegationByValIndexKey := []byte{0x33}
	return append(append(unbondingDelegationByValIndexKey, addresstypes.MustLengthPrefix(valAddr)...), addresstypes.MustLengthPrefix(delAddr)...)
}

// getUnbondingDelegationTimeKey creates the prefix for all unbonding delegations from a delegator
func getUnbondingDelegationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	unbondingQueueKey := []byte{0x41}
	return append(unbondingQueueKey, bz...)
}

// getValidatorKey creates the key for the validator with address
// VALUE: staking/Validator
func getValidatorKey(operatorAddr sdk.ValAddress) []byte {
	validatorsKey := []byte{0x21}
	return append(validatorsKey, addresstypes.MustLengthPrefix(operatorAddr)...)
}

// getLastValidatorPowerKey creates the bonded validator index key for an operator address
func getLastValidatorPowerKey(operator sdk.ValAddress) []byte {
	lastValidatorPowerKey := []byte{0x11}
	return append(lastValidatorPowerKey, addresstypes.MustLengthPrefix(operator)...)
}

func (s *KeeperTestSuite) TestLastTotalPowerMigrationToColls() {
	s.SetupTest()

	_, valAddrs := createValAddrs(100)

	err := testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			bz, err := s.cdc.Marshal(&gogotypes.Int64Value{Value: i})
			s.Require().NoError(err)

			s.ctx.KVStore(s.key).Set(getLastValidatorPowerKey(valAddrs[i]), bz)
		},
		"f28811f2b0a0ab9db60cdcae93680faff9dbadd4a3a8a2d088bb19b0428ad3a9",
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			bz, err := s.cdc.Marshal(&gogotypes.Int64Value{Value: i})
			s.Require().NoError(err)

			err = s.stakingKeeper.LastValidatorPower.Set(s.ctx, valAddrs[i], bz)
			s.Require().NoError(err)
		},
		"f28811f2b0a0ab9db60cdcae93680faff9dbadd4a3a8a2d088bb19b0428ad3a9",
	)
	s.Require().NoError(err)
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

func (s *KeeperTestSuite) TestUnbondingDelegationsMigrationToColls() {
	s.SetupTest()

	delAddrs, valAddrs := createValAddrs(100)
	err := testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			ubd := stakingtypes.UnbondingDelegation{
				DelegatorAddress: delAddrs[i].String(),
				ValidatorAddress: valAddrs[i].String(),
				Entries: []stakingtypes.UnbondingDelegationEntry{
					{
						CreationHeight: i,
						CompletionTime: time.Unix(i, 0).UTC(),
						Balance:        math.NewInt(i),
						UnbondingId:    uint64(i),
					},
				},
			}
			bz := stakingtypes.MustMarshalUBD(s.cdc, ubd)
			s.ctx.KVStore(s.key).Set(getUBDKey(delAddrs[i], valAddrs[i]), bz)
			s.ctx.KVStore(s.key).Set(getUBDByValIndexKey(delAddrs[i], valAddrs[i]), []byte{})
		},
		"d03ca412f3f6849b5148a2ca49ac2555f65f90b7fab6a289575ed337f15c0f4b",
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			ubd := stakingtypes.UnbondingDelegation{
				DelegatorAddress: delAddrs[i].String(),
				ValidatorAddress: valAddrs[i].String(),
				Entries: []stakingtypes.UnbondingDelegationEntry{
					{
						CreationHeight: i,
						CompletionTime: time.Unix(i, 0).UTC(),
						Balance:        math.NewInt(i),
						UnbondingId:    uint64(i),
					},
				},
			}
			err := s.stakingKeeper.SetUnbondingDelegation(s.ctx, ubd)
			s.Require().NoError(err)
		},
		"d03ca412f3f6849b5148a2ca49ac2555f65f90b7fab6a289575ed337f15c0f4b",
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestUBDQueueMigrationToColls() {
	s.SetupTest()

	err := testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			date := time.Date(2023, 8, 21, 14, 33, 1, 0, &time.Location{})
			// legacy Set method
			s.ctx.KVStore(s.key).Set(getUnbondingDelegationTimeKey(date), []byte{})
		},
		"7b8965aacc97646d6766a5a53bae397fe149d1c98fed027bea8774a18621ce6a",
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			date := time.Date(2023, 8, 21, 14, 33, 1, 0, &time.Location{})
			err := s.stakingKeeper.SetUBDQueueTimeSlice(s.ctx, date, nil)
			s.Require().NoError(err)
		},
		"7b8965aacc97646d6766a5a53bae397fe149d1c98fed027bea8774a18621ce6a",
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestValidatorsMigrationToColls() {
	s.SetupTest()
	pkAny, err := codectypes.NewAnyWithValue(PKs[0])
	s.Require().NoError(err)

	_, valAddrs := createValAddrs(100)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			val := stakingtypes.Validator{
				OperatorAddress:   valAddrs[i].String(),
				ConsensusPubkey:   pkAny,
				Jailed:            false,
				Status:            stakingtypes.Bonded,
				Tokens:            sdk.DefaultPowerReduction,
				DelegatorShares:   math.LegacyOneDec(),
				Description:       stakingtypes.Description{},
				UnbondingHeight:   int64(0),
				UnbondingTime:     time.Unix(0, 0).UTC(),
				Commission:        stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
				MinSelfDelegation: math.ZeroInt(),
			}
			valBz := stakingtypes.MustMarshalValidator(s.cdc, &val)
			// legacy Set method
			s.ctx.KVStore(s.key).Set(getValidatorKey(valAddrs[i]), valBz)
		},
		"6a8737af6309d53494a601e900832ec27763adefd7fe8ff104477d8130d7405f",
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			val := stakingtypes.Validator{
				OperatorAddress:   valAddrs[i].String(),
				ConsensusPubkey:   pkAny,
				Jailed:            false,
				Status:            stakingtypes.Bonded,
				Tokens:            sdk.DefaultPowerReduction,
				DelegatorShares:   math.LegacyOneDec(),
				Description:       stakingtypes.Description{},
				UnbondingHeight:   int64(0),
				UnbondingTime:     time.Unix(0, 0).UTC(),
				Commission:        stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
				MinSelfDelegation: math.ZeroInt(),
			}

			err := s.stakingKeeper.SetValidator(s.ctx, val)
			s.Require().NoError(err)
		},
		"6a8737af6309d53494a601e900832ec27763adefd7fe8ff104477d8130d7405f",
	)
	s.Require().NoError(err)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
