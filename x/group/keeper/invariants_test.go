package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
	"github.com/cosmos/cosmos-sdk/x/group/keeper"
)

type invariantTestSuite struct {
	suite.Suite

	ctx sdk.Context
	cdc *codec.ProtoCodec
	key *storetypes.KVStoreKey
}

func TestInvariantTestSuite(t *testing.T) {
	suite.Run(t, new(invariantTestSuite))
}

func (s *invariantTestSuite) SetupSuite() {
	interfaceRegistry := types.NewInterfaceRegistry()
	group.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	key := storetypes.NewKVStoreKey(group.ModuleName)
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	_ = cms.LoadLatestVersion()
	sdkCtx := sdk.NewContext(cms, cmtproto.Header{}, false, log.NewNopLogger())

	s.ctx = sdkCtx
	s.cdc = cdc
	s.key = key
}

func (s *invariantTestSuite) TestGroupTotalWeightInvariant() {
	sdkCtx, _ := s.ctx.CacheContext()
	curCtx, cdc, key := sdkCtx, s.cdc, s.key

	// Group Table
	groupTable, err := orm.NewAutoUInt64Table([2]byte{keeper.GroupTablePrefix}, keeper.GroupTableSeqPrefix, &group.GroupInfo{}, cdc)
	s.Require().NoError(err)

	// Group Member Table
	groupMemberTable, err := orm.NewPrimaryKeyTable([2]byte{keeper.GroupMemberTablePrefix}, &group.GroupMember{}, cdc)
	s.Require().NoError(err)

	groupMemberByGroupIndex, err := orm.NewIndex(groupMemberTable, keeper.GroupMemberByGroupIndexPrefix, func(val interface{}) ([]interface{}, error) {
		group := val.(*group.GroupMember).GroupId
		return []interface{}{group}, nil
	}, group.GroupMember{}.GroupId)
	s.Require().NoError(err)

	_, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()

	specs := map[string]struct {
		groupsInfo   *group.GroupInfo
		groupMembers []*group.GroupMember
		expBroken    bool
	}{
		"invariant not broken": {
			groupsInfo: &group.GroupInfo{
				Id:          1,
				Admin:       addr1.String(),
				Version:     1,
				TotalWeight: "3",
			},
			groupMembers: []*group.GroupMember{
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr1.String(),
						Weight:  "1",
					},
				},
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr2.String(),
						Weight:  "2",
					},
				},
			},
			expBroken: false,
		},

		"group's TotalWeight must be equal to sum of its members weight ": {
			groupsInfo: &group.GroupInfo{
				Id:          1,
				Admin:       addr1.String(),
				Version:     1,
				TotalWeight: "3",
			},
			groupMembers: []*group.GroupMember{
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr1.String(),
						Weight:  "2",
					},
				},
				{
					GroupId: 1,
					Member: &group.Member{
						Address: addr2.String(),
						Weight:  "2",
					},
				},
			},
			expBroken: true,
		},
	}

	for _, spec := range specs {
		cacheCurCtx, _ := curCtx.CacheContext()
		groupsInfo := spec.groupsInfo
		groupMembers := spec.groupMembers

		_, err := groupTable.Create(cacheCurCtx.KVStore(key), groupsInfo)
		s.Require().NoError(err)

		for i := 0; i < len(groupMembers); i++ {
			err := groupMemberTable.Create(cacheCurCtx.KVStore(key), groupMembers[i])
			s.Require().NoError(err)
		}

		_, broken := keeper.GroupTotalWeightInvariantHelper(cacheCurCtx, key, *groupTable, groupMemberByGroupIndex)
		s.Require().Equal(spec.expBroken, broken)

	}
}
