package v6_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	storetypes "cosmossdk.io/core/store"
	storetypesv1 "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	v6 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v6"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	bondedAcc    = authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName)
	notBondedAcc = authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName)
)

type MigrationsTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	stakingKeeper *stakingkeeper.Keeper
	bankKeeper    *stakingtestutil.MockBankKeeper
	accountKeeper *stakingtestutil.MockAccountKeeper
	storeService  storetypes.KVStoreService
	cdc           codec.BinaryCodec
}

func (s *MigrationsTestSuite) SetupTest() {
	require := s.Require()
	key := storetypesv1.NewKVStoreKey(stakingtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypesv1.NewTransientStoreKey("transient_test"))
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
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		address.NewBech32Codec("cosmosvaloper"),
		address.NewBech32Codec("cosmosvalcons"),
	)
	require.NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

	s.ctx = ctx
	s.stakingKeeper = keeper
	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper
	s.storeService = storeService
	s.cdc = encCfg.Codec
}

func TestMigrationsTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationsTestSuite))
}

// setValidatorQueueEntryPreMigration sets a validator queue entry in the old format (pre-migration)
// for testing migration functions.
func (s *MigrationsTestSuite) setValidatorQueueEntryPreMigration(endTime time.Time, endHeight int64, addrs []string) error {
	store := s.storeService.OpenKVStore(s.ctx)
	bz, err := s.cdc.Marshal(&stakingtypes.ValAddresses{Addresses: addrs})
	if err != nil {
		return err
	}
	return store.Set(stakingtypes.GetValidatorQueueKey(endTime, endHeight), bz)
}

// setUBDQueueEntryPreMigration sets a UBD queue entry in the old format (pre-migration)
// for testing migration functions.
func (s *MigrationsTestSuite) setUBDQueueEntryPreMigration(timestamp time.Time) error {
	store := s.storeService.OpenKVStore(s.ctx)
	timeBz := sdk.FormatTimeBytes(timestamp)
	return store.Set(stakingtypes.GetUnbondingDelegationTimeKey(timestamp), timeBz)
}

// setRedelegationQueueEntryPreMigration sets a redelegation queue entry in the old format (pre-migration)
// for testing migration functions.
func (s *MigrationsTestSuite) setRedelegationQueueEntryPreMigration(timestamp time.Time) error {
	store := s.storeService.OpenKVStore(s.ctx)
	timeBz := sdk.FormatTimeBytes(timestamp)
	return store.Set(stakingtypes.GetRedelegationTimeKey(timestamp), timeBz)
}

// migrateStore is a helper function that calls v6.MigrateStore with the correct parameters
func (s *MigrationsTestSuite) migrateStore() error {
	store := s.storeService.OpenKVStore(s.ctx)
	return v6.MigrateStore(s.ctx, store, s.stakingKeeper)
}

// TestMigrateStore_NoEntries tests migration with 0 entries
func (s *MigrationsTestSuite) TestMigrateStore_NoEntries() {
	err := s.migrateStore()
	s.Require().NoError(err)

	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Empty(slots)

	ubdSlots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Empty(ubdSlots)

	redSlots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Empty(redSlots)
}

// TestMigrateStore_AllQueues tests migration with multiple entries in all queues
func (s *MigrationsTestSuite) TestMigrateStore_AllQueues() {
	valTimes := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}
	valHeights := []int64{100, 200, 300}
	ubdTimes := []time.Time{
		time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC),
	}
	redTimes := []time.Time{
		time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC),
	}

	// Set up old format queue entries directly in store (pre-migration format)
	for i, t := range valTimes {
		err := s.setValidatorQueueEntryPreMigration(t, valHeights[i], []string{"cosmosvaloper1abc123"})
		s.Require().NoError(err)
	}
	for _, t := range ubdTimes {
		err := s.setUBDQueueEntryPreMigration(t)
		s.Require().NoError(err)
	}
	for _, t := range redTimes {
		err := s.setRedelegationQueueEntryPreMigration(t)
		s.Require().NoError(err)
	}

	// Run migration
	err := s.migrateStore()
	s.Require().NoError(err)

	// Verify all pending slots were populated
	valSlots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(valSlots, 3)
	s.Require().Equal(valTimes[0], valSlots[0].Time)
	s.Require().Equal(valHeights[0], valSlots[0].Height)
	s.Require().Equal(valTimes[1], valSlots[1].Time)
	s.Require().Equal(valHeights[1], valSlots[1].Height)
	s.Require().Equal(valTimes[2], valSlots[2].Time)
	s.Require().Equal(valHeights[2], valSlots[2].Height)

	ubdSlots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(ubdSlots, 3)
	s.Require().Equal(ubdTimes[0], ubdSlots[0])
	s.Require().Equal(ubdTimes[1], ubdSlots[1])
	s.Require().Equal(ubdTimes[2], ubdSlots[2])

	redSlots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(redSlots, 3)
	s.Require().Equal(redTimes[0], redSlots[0])
	s.Require().Equal(redTimes[1], redSlots[1])
	s.Require().Equal(redTimes[2], redSlots[2])
}

// TestMigrateStore_ValidatorQueue_NoEntries tests migration with 0 validator queue entries
func (s *MigrationsTestSuite) TestMigrateStore_ValidatorQueue_NoEntries() {
	err := s.migrateStore()
	s.Require().NoError(err)

	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Empty(slots)
}

// TestMigrateStore_ValidatorQueue_SingleEntry tests migration with 1 validator queue entry
func (s *MigrationsTestSuite) TestMigrateStore_ValidatorQueue_SingleEntry() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testHeight := int64(100)

	// Set up old format queue entry directly in store (pre-migration format)
	err := s.setValidatorQueueEntryPreMigration(testTime, testHeight, []string{"cosmosvaloper1abc123"})
	s.Require().NoError(err)

	// Run migration
	err = s.migrateStore()
	s.Require().NoError(err)

	// Verify pending slots were populated
	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1)
	s.Require().Equal(testTime, slots[0].Time)
	s.Require().Equal(testHeight, slots[0].Height)
}

// TestMigrateStore_ValidatorQueue_MultipleEntries tests migration with multiple validator queue entries
func (s *MigrationsTestSuite) TestMigrateStore_ValidatorQueue_MultipleEntries() {
	testTimes := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}
	testHeights := []int64{100, 200, 300}

	// Set up old format queue entries directly in store (pre-migration format)
	for i, t := range testTimes {
		err := s.setValidatorQueueEntryPreMigration(t, testHeights[i], []string{"cosmosvaloper1abc123"})
		s.Require().NoError(err)
	}

	// Run migration
	err := s.migrateStore()
	s.Require().NoError(err)

	// Verify pending slots were populated
	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 3)

	// Slots should be sorted by time, then height
	s.Require().Equal(testTimes[0], slots[0].Time)
	s.Require().Equal(testHeights[0], slots[0].Height)
	s.Require().Equal(testTimes[1], slots[1].Time)
	s.Require().Equal(testHeights[1], slots[1].Height)
	s.Require().Equal(testTimes[2], slots[2].Time)
	s.Require().Equal(testHeights[2], slots[2].Height)
}

// TestMigrateStore_UBDQueue_NoEntries tests migration with 0 UBD queue entries
func (s *MigrationsTestSuite) TestMigrateStore_UBDQueue_NoEntries() {
	err := s.migrateStore()
	s.Require().NoError(err)

	slots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Empty(slots)
}

// TestMigrateStore_UBDQueue_SingleEntry tests migration with 1 UBD queue entry
func (s *MigrationsTestSuite) TestMigrateStore_UBDQueue_SingleEntry() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Set up old format queue entry directly in store (pre-migration format)
	err := s.setUBDQueueEntryPreMigration(testTime)
	s.Require().NoError(err)

	// Run migration
	err = s.migrateStore()
	s.Require().NoError(err)

	// Verify pending slots were populated
	slots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1)
	s.Require().Equal(testTime, slots[0])
}

// TestMigrateStore_UBDQueue_MultipleEntries tests migration with multiple UBD queue entries
func (s *MigrationsTestSuite) TestMigrateStore_UBDQueue_MultipleEntries() {
	testTimes := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}

	// Set up old format queue entries directly in store (pre-migration format)
	for _, t := range testTimes {
		err := s.setUBDQueueEntryPreMigration(t)
		s.Require().NoError(err)
	}

	// Run migration
	err := s.migrateStore()
	s.Require().NoError(err)

	// Verify pending slots were populated
	slots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 3)

	// Slots should be sorted by time
	s.Require().Equal(testTimes[0], slots[0])
	s.Require().Equal(testTimes[1], slots[1])
	s.Require().Equal(testTimes[2], slots[2])
}

// TestMigrateStore_RedelegationQueue_NoEntries tests migration with 0 redelegation queue entries
func (s *MigrationsTestSuite) TestMigrateStore_RedelegationQueue_NoEntries() {
	err := s.migrateStore()
	s.Require().NoError(err)

	slots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Empty(slots)
}

// TestMigrateStore_RedelegationQueue_SingleEntry tests migration with 1 redelegation queue entry
func (s *MigrationsTestSuite) TestMigrateStore_RedelegationQueue_SingleEntry() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Set up old format queue entry directly in store (pre-migration format)
	err := s.setRedelegationQueueEntryPreMigration(testTime)
	s.Require().NoError(err)

	// Run migration
	err = s.migrateStore()
	s.Require().NoError(err)

	// Verify pending slots were populated
	slots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1)
	s.Require().Equal(testTime, slots[0])
}

// TestMigrateStore_RedelegationQueue_MultipleEntries tests migration with multiple redelegation queue entries
func (s *MigrationsTestSuite) TestMigrateStore_RedelegationQueue_MultipleEntries() {
	testTimes := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}

	// Set up old format queue entries directly in store (pre-migration format)
	for _, t := range testTimes {
		err := s.setRedelegationQueueEntryPreMigration(t)
		s.Require().NoError(err)
	}

	// Run migration
	err := s.migrateStore()
	s.Require().NoError(err)

	// Verify pending slots were populated
	slots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 3)

	// Slots should be sorted by time
	s.Require().Equal(testTimes[0], slots[0])
	s.Require().Equal(testTimes[1], slots[1])
	s.Require().Equal(testTimes[2], slots[2])
}
