package keeper_test

import (
	"testing"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	authtypes "cosmossdk.io/x/auth/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtestutil "cosmossdk.io/x/staking/testutil"
	stakingtypes "cosmossdk.io/x/staking/types"

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
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()})
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
	require.NoError(keeper.Params.Set(ctx, stakingtypes.DefaultParams()))

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
	resParams, err := keeper.Params.Get(ctx)
	require.NoError(err)
	require.Equal(expParams, resParams)

	expParams.MaxValidators = 555
	expParams.MaxEntries = 111
	require.NoError(keeper.Params.Set(ctx, expParams))
	resParams, err = keeper.Params.Get(ctx)
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

// getRedelegationTimeKey returns a key prefix for indexing an unbonding
// redelegation based on a completion time.
func getRedelegationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	redelegationQueueKey := []byte{0x42}
	return append(redelegationQueueKey, bz...)
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

// getValidatorQueueKey returns the prefix key used for getting a set of unbonding
// validators whose unbonding completion occurs at the given time and height.
func getValidatorQueueKey(timestamp time.Time, height int64) []byte {
	validatorQueueKey := []byte{0x43}

	heightBz := sdk.Uint64ToBigEndian(uint64(height))
	timeBz := sdk.FormatTimeBytes(timestamp)
	timeBzL := len(timeBz)
	prefixL := len(validatorQueueKey)

	bz := make([]byte, prefixL+8+timeBzL+8)

	// copy the prefix
	copy(bz[:prefixL], validatorQueueKey)

	// copy the encoded time bytes length
	copy(bz[prefixL:prefixL+8], sdk.Uint64ToBigEndian(uint64(timeBzL)))

	// copy the encoded time bytes
	copy(bz[prefixL+8:prefixL+8+timeBzL], timeBz)

	// copy the encoded height
	copy(bz[prefixL+8+timeBzL:], heightBz)

	return bz
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
		"198aa9b8c1d9bc02308b7b2a48944f3e4b05c6b8312cb0bcc73518d1260f682d",
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			var intV gogotypes.Int64Value
			intV.Value = i

			err = s.stakingKeeper.LastValidatorPower.Set(s.ctx, valAddrs[i], intV)
			s.Require().NoError(err)
		},
		"198aa9b8c1d9bc02308b7b2a48944f3e4b05c6b8312cb0bcc73518d1260f682d",
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
		"cae99e5c0498356a290f9478b7db73d522840b736878a9d4c00b56d1ddd7fd04",
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
		"cae99e5c0498356a290f9478b7db73d522840b736878a9d4c00b56d1ddd7fd04",
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
		"1b7687449a83f8176a60aeced7bcfc69a2b957b9eefad60c69a9fae9acfdaa81", // this hash obtained when ran this test in main branch
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
		"1b7687449a83f8176a60aeced7bcfc69a2b957b9eefad60c69a9fae9acfdaa81",
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
			bz := s.cdc.MustMarshal(&ubd)
			s.ctx.KVStore(s.key).Set(getUBDKey(delAddrs[i], valAddrs[i]), bz)
			s.ctx.KVStore(s.key).Set(getUBDByValIndexKey(delAddrs[i], valAddrs[i]), []byte{})
		},
		"70454ad98368368aaff32d207a7a115fba49133ecf2a225d8e3eca88c6b2324c",
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
		"70454ad98368368aaff32d207a7a115fba49133ecf2a225d8e3eca88c6b2324c",
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
		"2dd1dd08ea1cc2b0a076c420e3888b218647b9409b435f75e5730b0e4f25e890",
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
		"2dd1dd08ea1cc2b0a076c420e3888b218647b9409b435f75e5730b0e4f25e890",
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
			valBz := s.cdc.MustMarshal(&val)
			// legacy Set method
			s.ctx.KVStore(s.key).Set(getValidatorKey(valAddrs[i]), valBz)
		},
		"aa495d55fb45df89fcf1d4326331bfc1244ef879764abe76f6ce2a41ccd4180d",
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
		"aa495d55fb45df89fcf1d4326331bfc1244ef879764abe76f6ce2a41ccd4180d",
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestValidatorQueueMigrationToColls() {
	s.SetupTest()
	_, valAddrs := createValAddrs(100)
	endTime := time.Unix(0, 0).UTC()
	endHeight := int64(10)
	err := testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			var addrs []string
			addrs = append(addrs, valAddrs[i].String())
			bz, err := s.cdc.Marshal(&stakingtypes.ValAddresses{Addresses: addrs})
			s.Require().NoError(err)

			// legacy Set method
			s.ctx.KVStore(s.key).Set(getValidatorQueueKey(endTime, endHeight), bz)
		},
		"b23a5905ced2b76c46ddd0f7d39e2ed7dcc68cd81993c497ee314b2e1a158595",
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			var addrs []string
			addrs = append(addrs, valAddrs[i].String())

			err := s.stakingKeeper.SetUnbondingValidatorsQueue(s.ctx, endTime, endHeight, addrs)
			s.Require().NoError(err)
		},
		"b23a5905ced2b76c46ddd0f7d39e2ed7dcc68cd81993c497ee314b2e1a158595",
	)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestRedelegationQueueMigrationToColls() {
	s.SetupTest()

	addrs, valAddrs := createValAddrs(101)
	err := testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			date := time.Unix(i, i)
			dvvTriplets := stakingtypes.DVVTriplets{
				Triplets: []stakingtypes.DVVTriplet{
					{
						DelegatorAddress:    addrs[i].String(),
						ValidatorSrcAddress: valAddrs[i].String(),
						ValidatorDstAddress: valAddrs[i+1].String(),
					},
				},
			}
			bz, err := s.cdc.Marshal(&dvvTriplets)
			s.Require().NoError(err)
			s.ctx.KVStore(s.key).Set(getRedelegationTimeKey(date), bz)
		},
		"d6a1c46c7c5793ff7094b67252c82883aecb75c8359428a59aacd3657fa16235",
	)
	s.Require().NoError(err)

	err = testutil.DiffCollectionsMigration(
		s.ctx,
		s.key,
		100,
		func(i int64) {
			date := time.Unix(i, i)
			dvvTriplets := stakingtypes.DVVTriplets{
				Triplets: []stakingtypes.DVVTriplet{
					{
						DelegatorAddress:    addrs[i].String(),
						ValidatorSrcAddress: valAddrs[i].String(),
						ValidatorDstAddress: valAddrs[i+1].String(),
					},
				},
			}
			err := s.stakingKeeper.SetRedelegationQueueTimeSlice(s.ctx, date, dvvTriplets.Triplets)
			s.Require().NoError(err)
		},
		"d6a1c46c7c5793ff7094b67252c82883aecb75c8359428a59aacd3657fa16235",
	)
	s.Require().NoError(err)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
