package keeper_test

import (
	"bytes"
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmtime "github.com/tendermint/tendermint/libs/time"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/keeper"
)

type TestSuite struct {
	suite.Suite

	app             *simapp.SimApp
	sdkCtx          sdk.Context
	ctx             context.Context
	addrs           []sdk.AccAddress
	groupID         uint64
	groupPolicyAddr sdk.AccAddress
	keeper          keeper.Keeper
	blockTime       time.Time
}

func (s *TestSuite) SetupTest() {
	app := simapp.Setup(s.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	s.blockTime = tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: s.blockTime})

	s.app = app
	s.sdkCtx = ctx
	s.ctx = sdk.WrapSDKContext(ctx)
	s.keeper = s.app.GroupKeeper
	s.addrs = simapp.AddTestAddrsIncremental(app, ctx, 6, sdk.NewInt(30000000))

	// Initial group, group policy and balance setup
	members := []group.Member{
		{Address: s.addrs[4].String(), Weight: "1"}, {Address: s.addrs[1].String(), Weight: "2"},
	}
	groupRes, err := s.keeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    s.addrs[0].String(),
		Members:  members,
		Metadata: nil,
	})
	s.Require().NoError(err)
	s.groupID = groupRes.GroupId

	policy := group.NewThresholdDecisionPolicy(
		"2",
		time.Second,
	)
	policyReq := &group.MsgCreateGroupPolicy{
		Admin:    s.addrs[0].String(),
		GroupId:  s.groupID,
		Metadata: nil,
	}
	err = policyReq.SetDecisionPolicy(policy)
	s.Require().NoError(err)
	policyRes, err := s.keeper.CreateGroupPolicy(s.ctx, policyReq)
	s.Require().NoError(err)
	addr, err := sdk.AccAddressFromBech32(policyRes.Address)
	s.Require().NoError(err)
	s.groupPolicyAddr = addr
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.sdkCtx, s.groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10000)}))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestCreateGroup() {
	addrs := s.addrs
	addr1 := addrs[0]
	addr3 := addrs[2]
	addr5 := addrs[4]
	addr6 := addrs[5]

	members := []group.Member{{
		Address:  addr5.String(),
		Weight:   "1",
		Metadata: nil,
		AddedAt:  s.blockTime,
	}, {
		Address:  addr6.String(),
		Weight:   "2",
		Metadata: nil,
		AddedAt:  s.blockTime,
	}}

	expGroups := []*group.GroupInfo{
		{
			GroupId:     s.groupID,
			Version:     1,
			Admin:       addr1.String(),
			TotalWeight: "3",
			Metadata:    nil,
			CreatedAt:   s.blockTime,
		},
		{
			GroupId:     2,
			Version:     1,
			Admin:       addr1.String(),
			TotalWeight: "3",
			Metadata:    nil,
			CreatedAt:   s.blockTime,
		},
	}

	specs := map[string]struct {
		req       *group.MsgCreateGroup
		expErr    bool
		expGroups []*group.GroupInfo
	}{
		"all good": {
			req: &group.MsgCreateGroup{
				Admin:    addr1.String(),
				Members:  members,
				Metadata: nil,
			},
			expGroups: expGroups,
		},
		"group metadata too long": {
			req: &group.MsgCreateGroup{
				Admin:    addr1.String(),
				Members:  members,
				Metadata: bytes.Repeat([]byte{1}, 256),
			},
			expErr: true,
		},
		"member metadata too long": {
			req: &group.MsgCreateGroup{
				Admin: addr1.String(),
				Members: []group.Member{{
					Address:  addr3.String(),
					Weight:   "1",
					Metadata: bytes.Repeat([]byte{1}, 256),
				}},
				Metadata: nil,
			},
			expErr: true,
		},
		"zero member weight": {
			req: &group.MsgCreateGroup{
				Admin: addr1.String(),
				Members: []group.Member{{
					Address:  addr3.String(),
					Weight:   "0",
					Metadata: nil,
				}},
				Metadata: nil,
			},
			expErr: true,
		},
	}

	var seq uint32 = 1
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			res, err := s.keeper.CreateGroup(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				_, err := s.keeper.GroupInfo(s.ctx, &group.QueryGroupInfoRequest{GroupId: uint64(seq + 1)})
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			id := res.GroupId

			seq++
			s.Assert().Equal(uint64(seq), id)

			// then all data persisted
			loadedGroupRes, err := s.keeper.GroupInfo(s.ctx, &group.QueryGroupInfoRequest{GroupId: id})
			s.Require().NoError(err)
			s.Assert().Equal(spec.req.Admin, loadedGroupRes.Info.Admin)
			s.Assert().Equal(spec.req.Metadata, loadedGroupRes.Info.Metadata)
			s.Assert().Equal(id, loadedGroupRes.Info.GroupId)
			s.Assert().Equal(uint64(1), loadedGroupRes.Info.Version)

			// and members are stored as well
			membersRes, err := s.keeper.GroupMembers(s.ctx, &group.QueryGroupMembersRequest{GroupId: id})
			s.Require().NoError(err)
			loadedMembers := membersRes.Members
			s.Require().Equal(len(members), len(loadedMembers))
			// we reorder members by address to be able to compare them
			sort.Slice(members, func(i, j int) bool {
				addri, err := sdk.AccAddressFromBech32(members[i].Address)
				s.Require().NoError(err)
				addrj, err := sdk.AccAddressFromBech32(members[j].Address)
				s.Require().NoError(err)
				return bytes.Compare(addri, addrj) < 0
			})
			for i := range loadedMembers {
				s.Assert().Equal(members[i].Metadata, loadedMembers[i].Member.Metadata)
				s.Assert().Equal(members[i].Address, loadedMembers[i].Member.Address)
				s.Assert().Equal(members[i].Weight, loadedMembers[i].Member.Weight)
				s.Assert().Equal(members[i].AddedAt, loadedMembers[i].Member.AddedAt)
				s.Assert().Equal(id, loadedMembers[i].GroupId)
			}

			// query groups by admin
			groupsRes, err := s.keeper.GroupsByAdmin(s.ctx, &group.QueryGroupsByAdminRequest{Admin: addr1.String()})
			s.Require().NoError(err)
			loadedGroups := groupsRes.Groups
			s.Require().Equal(len(spec.expGroups), len(loadedGroups))
			for i := range loadedGroups {
				s.Assert().Equal(spec.expGroups[i].Metadata, loadedGroups[i].Metadata)
				s.Assert().Equal(spec.expGroups[i].Admin, loadedGroups[i].Admin)
				s.Assert().Equal(spec.expGroups[i].TotalWeight, loadedGroups[i].TotalWeight)
				s.Assert().Equal(spec.expGroups[i].GroupId, loadedGroups[i].GroupId)
				s.Assert().Equal(spec.expGroups[i].Version, loadedGroups[i].Version)
				s.Assert().Equal(spec.expGroups[i].CreatedAt, loadedGroups[i].CreatedAt)
			}
		})
	}

}

func (s *TestSuite) TestUpdateGroupAdmin() {
	addrs := s.addrs
	addr1 := addrs[0]
	addr2 := addrs[1]
	addr3 := addrs[2]
	addr4 := addrs[3]

	members := []group.Member{{
		Address:  addr1.String(),
		Weight:   "1",
		Metadata: nil,
		AddedAt:  s.blockTime,
	}}
	oldAdmin := addr2.String()
	newAdmin := addr3.String()
	groupRes, err := s.keeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    oldAdmin,
		Members:  members,
		Metadata: nil,
	})
	s.Require().NoError(err)
	groupID := groupRes.GroupId
	specs := map[string]struct {
		req       *group.MsgUpdateGroupAdmin
		expStored *group.GroupInfo
		expErr    bool
	}{
		"with correct admin": {
			req: &group.MsgUpdateGroupAdmin{
				GroupId:  groupID,
				Admin:    oldAdmin,
				NewAdmin: newAdmin,
			},
			expStored: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       newAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     2,
				CreatedAt:   s.blockTime,
			},
		},
		"with wrong admin": {
			req: &group.MsgUpdateGroupAdmin{
				GroupId:  groupID,
				Admin:    addr4.String(),
				NewAdmin: newAdmin,
			},
			expErr: true,
			expStored: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       oldAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
				CreatedAt:   s.blockTime,
			},
		},
		"with unknown groupID": {
			req: &group.MsgUpdateGroupAdmin{
				GroupId:  999,
				Admin:    oldAdmin,
				NewAdmin: newAdmin,
			},
			expErr: true,
			expStored: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       oldAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
				CreatedAt:   s.blockTime,
			},
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			_, err := s.keeper.UpdateGroupAdmin(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// then
			res, err := s.keeper.GroupInfo(s.ctx, &group.QueryGroupInfoRequest{GroupId: groupID})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expStored, res.Info)
		})
	}
}

func (s *TestSuite) TestUpdateGroupMetadata() {
	addrs := s.addrs
	addr1 := addrs[0]
	addr3 := addrs[2]

	oldAdmin := addr1.String()
	groupID := s.groupID

	specs := map[string]struct {
		req       *group.MsgUpdateGroupMetadata
		expErr    bool
		expStored *group.GroupInfo
	}{
		"with correct admin": {
			req: &group.MsgUpdateGroupMetadata{
				GroupId:  groupID,
				Admin:    oldAdmin,
				Metadata: []byte{1, 2, 3},
			},
			expStored: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       oldAdmin,
				Metadata:    []byte{1, 2, 3},
				TotalWeight: "3",
				Version:     2,
				CreatedAt:   s.blockTime,
			},
		},
		"with wrong admin": {
			req: &group.MsgUpdateGroupMetadata{
				GroupId:  groupID,
				Admin:    addr3.String(),
				Metadata: []byte{1, 2, 3},
			},
			expErr: true,
			expStored: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       oldAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
				CreatedAt:   s.blockTime,
			},
		},
		"with unknown groupid": {
			req: &group.MsgUpdateGroupMetadata{
				GroupId:  999,
				Admin:    oldAdmin,
				Metadata: []byte{1, 2, 3},
			},
			expErr: true,
			expStored: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       oldAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
				CreatedAt:   s.blockTime,
			},
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			sdkCtx, _ := s.sdkCtx.CacheContext()
			ctx := sdk.WrapSDKContext(sdkCtx)
			_, err := s.keeper.UpdateGroupMetadata(ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// then
			res, err := s.keeper.GroupInfo(ctx, &group.QueryGroupInfoRequest{GroupId: groupID})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expStored, res.Info)
		})
	}
}

func (s *TestSuite) TestUpdateGroupMembers() {
	addrs := s.addrs
	addr3 := addrs[2]
	addr4 := addrs[3]
	addr5 := addrs[4]
	addr6 := addrs[5]

	member1 := addr5.String()
	member2 := addr6.String()
	members := []group.Member{{
		Address:  member1,
		Weight:   "1",
		Metadata: nil,
	}}

	myAdmin := addr4.String()
	groupRes, err := s.keeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    myAdmin,
		Members:  members,
		Metadata: nil,
	})
	s.Require().NoError(err)
	groupID := groupRes.GroupId

	specs := map[string]struct {
		req        *group.MsgUpdateGroupMembers
		expErr     bool
		expGroup   *group.GroupInfo
		expMembers []*group.GroupMember
	}{
		"add new member": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address:  member2,
					Weight:   "2",
					Metadata: nil,
				}},
			},
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "3",
				Version:     2,
				CreatedAt:   s.blockTime,
			},
			expMembers: []*group.GroupMember{
				{
					Member: &group.Member{
						Address:  member2,
						Weight:   "2",
						Metadata: nil,
					},
					GroupId: groupID,
				},
				{
					Member: &group.Member{
						Address:  member1,
						Weight:   "1",
						Metadata: nil,
					},
					GroupId: groupID,
				},
			},
		},
		"update member": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address:  member1,
					Weight:   "2",
					Metadata: []byte{1, 2, 3},
				}},
			},
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "2",
				Version:     2,
				CreatedAt:   s.blockTime,
			},
			expMembers: []*group.GroupMember{
				{
					GroupId: groupID,
					Member: &group.Member{
						Address:  member1,
						Weight:   "2",
						Metadata: []byte{1, 2, 3},
					},
				},
			},
		},
		"update member with same data": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address: member1,
					Weight:  "1",
				}},
			},
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     2,
				CreatedAt:   s.blockTime,
			},
			expMembers: []*group.GroupMember{
				{
					GroupId: groupID,
					Member: &group.Member{
						Address: member1,
						Weight:  "1",
					},
				},
			},
		},
		"replace member": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{
					{
						Address:  member1,
						Weight:   "0",
						Metadata: nil,
					},
					{
						Address:  member2,
						Weight:   "1",
						Metadata: nil,
					},
				},
			},
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     2,
				CreatedAt:   s.blockTime,
			},
			expMembers: []*group.GroupMember{{
				GroupId: groupID,
				Member: &group.Member{
					Address:  member2,
					Weight:   "1",
					Metadata: nil,
				},
			}},
		},
		"remove existing member": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address:  member1,
					Weight:   "0",
					Metadata: nil,
				}},
			},
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "0",
				Version:     2,
				CreatedAt:   s.blockTime,
			},
			expMembers: []*group.GroupMember{},
		},
		"remove unknown member": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address:  addr4.String(),
					Weight:   "0",
					Metadata: nil,
				}},
			},
			expErr: true,
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
				CreatedAt:   s.blockTime,
			},
			expMembers: []*group.GroupMember{{
				GroupId: groupID,
				Member: &group.Member{
					Address:  member1,
					Weight:   "1",
					Metadata: nil,
				},
			}},
		},
		"with wrong admin": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   addr3.String(),
				MemberUpdates: []group.Member{{
					Address:  member1,
					Weight:   "2",
					Metadata: nil,
				}},
			},
			expErr: true,
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
				CreatedAt:   s.blockTime,
			},
			expMembers: []*group.GroupMember{{
				GroupId: groupID,
				Member: &group.Member{
					Address: member1,
					Weight:  "1",
				},
			}},
		},
		"with unknown groupID": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: 999,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address:  member1,
					Weight:   "2",
					Metadata: nil,
				}},
			},
			expErr: true,
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
				CreatedAt:   s.blockTime,
			},
			expMembers: []*group.GroupMember{{
				GroupId: groupID,
				Member: &group.Member{
					Address: member1,
					Weight:  "1",
				},
			}},
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			sdkCtx, _ := s.sdkCtx.CacheContext()
			ctx := sdk.WrapSDKContext(sdkCtx)
			_, err := s.keeper.UpdateGroupMembers(ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// then
			res, err := s.keeper.GroupInfo(ctx, &group.QueryGroupInfoRequest{GroupId: groupID})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expGroup, res.Info)

			// and members persisted
			membersRes, err := s.keeper.GroupMembers(ctx, &group.QueryGroupMembersRequest{GroupId: groupID})
			s.Require().NoError(err)
			loadedMembers := membersRes.Members
			s.Require().Equal(len(spec.expMembers), len(loadedMembers))
			// we reorder group members by address to be able to compare them
			sort.Slice(spec.expMembers, func(i, j int) bool {
				addri, err := sdk.AccAddressFromBech32(spec.expMembers[i].Member.Address)
				s.Require().NoError(err)
				addrj, err := sdk.AccAddressFromBech32(spec.expMembers[j].Member.Address)
				s.Require().NoError(err)
				return bytes.Compare(addri, addrj) < 0
			})
			for i := range loadedMembers {
				s.Assert().Equal(spec.expMembers[i].Member.Metadata, loadedMembers[i].Member.Metadata)
				s.Assert().Equal(spec.expMembers[i].Member.Address, loadedMembers[i].Member.Address)
				s.Assert().Equal(spec.expMembers[i].Member.Weight, loadedMembers[i].Member.Weight)
				s.Assert().Equal(spec.expMembers[i].GroupId, loadedMembers[i].GroupId)
			}
		})
	}
}

func (s *TestSuite) TestCreateGroupPolicy() {
	addrs := s.addrs
	addr1 := addrs[0]
	addr4 := addrs[3]

	groupRes, err := s.keeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    addr1.String(),
		Members:  nil,
		Metadata: nil,
	})
	s.Require().NoError(err)
	myGroupID := groupRes.GroupId

	specs := map[string]struct {
		req    *group.MsgCreateGroupPolicy
		policy group.DecisionPolicy
		expErr bool
	}{
		"all good": {
			req: &group.MsgCreateGroupPolicy{
				Admin:    addr1.String(),
				Metadata: nil,
				GroupId:  myGroupID,
			},
			policy: group.NewThresholdDecisionPolicy(
				"1",
				time.Second,
			),
		},
		"decision policy threshold > total group weight": {
			req: &group.MsgCreateGroupPolicy{
				Admin:    addr1.String(),
				Metadata: nil,
				GroupId:  myGroupID,
			},
			policy: group.NewThresholdDecisionPolicy(
				"10",
				time.Second,
			),
		},
		"group id does not exists": {
			req: &group.MsgCreateGroupPolicy{
				Admin:    addr1.String(),
				Metadata: nil,
				GroupId:  9999,
			},
			policy: group.NewThresholdDecisionPolicy(
				"1",
				time.Second,
			),
			expErr: true,
		},
		"admin not group admin": {
			req: &group.MsgCreateGroupPolicy{
				Admin:    addr4.String(),
				Metadata: nil,
				GroupId:  myGroupID,
			},
			policy: group.NewThresholdDecisionPolicy(
				"1",
				time.Second,
			),
			expErr: true,
		},
		"metadata too long": {
			req: &group.MsgCreateGroupPolicy{
				Admin:    addr1.String(),
				Metadata: []byte(strings.Repeat("a", 256)),
				GroupId:  myGroupID,
			},
			policy: group.NewThresholdDecisionPolicy(
				"1",
				time.Second,
			),
			expErr: true,
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			err := spec.req.SetDecisionPolicy(spec.policy)
			s.Require().NoError(err)

			res, err := s.keeper.CreateGroupPolicy(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			addr := res.Address

			// then all data persisted
			groupPolicyRes, err := s.keeper.GroupPolicyInfo(s.ctx, &group.QueryGroupPolicyInfoRequest{Address: addr})
			s.Require().NoError(err)

			groupPolicy := groupPolicyRes.Info
			s.Assert().Equal(addr, groupPolicy.Address)
			s.Assert().Equal(myGroupID, groupPolicy.GroupId)
			s.Assert().Equal(spec.req.Admin, groupPolicy.Admin)
			s.Assert().Equal(spec.req.Metadata, groupPolicy.Metadata)
			s.Assert().Equal(uint64(1), groupPolicy.Version)
			s.Assert().Equal(spec.policy.(*group.ThresholdDecisionPolicy), groupPolicy.GetDecisionPolicy())
		})
	}
}

func (s *TestSuite) TestUpdateGroupPolicyAdmin() {
	addrs := s.addrs
	addr1 := addrs[0]
	addr2 := addrs[1]
	addr5 := addrs[4]

	admin, newAdmin := addr1, addr2
	groupPolicyAddr, myGroupID, policy := createGroupAndGroupPolicy(admin, s)

	specs := map[string]struct {
		req            *group.MsgUpdateGroupPolicyAdmin
		expGroupPolicy *group.GroupPolicyInfo
		expErr         bool
	}{
		"with wrong admin": {
			req: &group.MsgUpdateGroupPolicyAdmin{
				Admin:    addr5.String(),
				Address:  groupPolicyAddr,
				NewAdmin: newAdmin.String(),
			},
			expGroupPolicy: &group.GroupPolicyInfo{
				Admin:          admin.String(),
				Address:        groupPolicyAddr,
				GroupId:        myGroupID,
				Metadata:       nil,
				Version:        2,
				DecisionPolicy: nil,
				CreatedAt:      s.blockTime,
			},
			expErr: true,
		},
		"with wrong group policy": {
			req: &group.MsgUpdateGroupPolicyAdmin{
				Admin:    admin.String(),
				Address:  addr5.String(),
				NewAdmin: newAdmin.String(),
			},
			expGroupPolicy: &group.GroupPolicyInfo{
				Admin:          admin.String(),
				Address:        groupPolicyAddr,
				GroupId:        myGroupID,
				Metadata:       nil,
				Version:        2,
				DecisionPolicy: nil,
				CreatedAt:      s.blockTime,
			},
			expErr: true,
		},
		"correct data": {
			req: &group.MsgUpdateGroupPolicyAdmin{
				Admin:    admin.String(),
				Address:  groupPolicyAddr,
				NewAdmin: newAdmin.String(),
			},
			expGroupPolicy: &group.GroupPolicyInfo{
				Admin:          newAdmin.String(),
				Address:        groupPolicyAddr,
				GroupId:        myGroupID,
				Metadata:       nil,
				Version:        2,
				DecisionPolicy: nil,
				CreatedAt:      s.blockTime,
			},
			expErr: false,
		},
	}
	for msg, spec := range specs {
		spec := spec
		err := spec.expGroupPolicy.SetDecisionPolicy(policy)
		s.Require().NoError(err)

		s.Run(msg, func() {
			_, err := s.keeper.UpdateGroupPolicyAdmin(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			res, err := s.keeper.GroupPolicyInfo(s.ctx, &group.QueryGroupPolicyInfoRequest{
				Address: groupPolicyAddr,
			})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expGroupPolicy, res.Info)
		})
	}
}

func (s *TestSuite) TestUpdateGroupPolicyMetadata() {
	addrs := s.addrs
	addr1 := addrs[0]
	addr5 := addrs[4]

	admin := addr1
	groupPolicyAddr, myGroupID, policy := createGroupAndGroupPolicy(admin, s)

	specs := map[string]struct {
		req            *group.MsgUpdateGroupPolicyMetadata
		expGroupPolicy *group.GroupPolicyInfo
		expErr         bool
	}{
		"with wrong admin": {
			req: &group.MsgUpdateGroupPolicyMetadata{
				Admin:    addr5.String(),
				Address:  groupPolicyAddr,
				Metadata: []byte("hello"),
			},
			expGroupPolicy: &group.GroupPolicyInfo{},
			expErr:         true,
		},
		"with wrong group policy": {
			req: &group.MsgUpdateGroupPolicyMetadata{
				Admin:    admin.String(),
				Address:  addr5.String(),
				Metadata: []byte("hello"),
			},
			expGroupPolicy: &group.GroupPolicyInfo{},
			expErr:         true,
		},
		"with comment too long": {
			req: &group.MsgUpdateGroupPolicyMetadata{
				Admin:    admin.String(),
				Address:  addr5.String(),
				Metadata: []byte(strings.Repeat("a", 256)),
			},
			expGroupPolicy: &group.GroupPolicyInfo{},
			expErr:         true,
		},
		"correct data": {
			req: &group.MsgUpdateGroupPolicyMetadata{
				Admin:    admin.String(),
				Address:  groupPolicyAddr,
				Metadata: []byte("hello"),
			},
			expGroupPolicy: &group.GroupPolicyInfo{
				Admin:          admin.String(),
				Address:        groupPolicyAddr,
				GroupId:        myGroupID,
				Metadata:       []byte("hello"),
				Version:        2,
				DecisionPolicy: nil,
				CreatedAt:      s.blockTime,
			},
			expErr: false,
		},
	}
	for msg, spec := range specs {
		spec := spec
		err := spec.expGroupPolicy.SetDecisionPolicy(policy)
		s.Require().NoError(err)

		s.Run(msg, func() {
			_, err := s.keeper.UpdateGroupPolicyMetadata(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			res, err := s.keeper.GroupPolicyInfo(s.ctx, &group.QueryGroupPolicyInfoRequest{
				Address: groupPolicyAddr,
			})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expGroupPolicy, res.Info)
		})
	}
}

func (s *TestSuite) TestUpdateGroupPolicyDecisionPolicy() {
	addrs := s.addrs
	addr1 := addrs[0]
	addr5 := addrs[4]

	admin := addr1
	groupPolicyAddr, myGroupID, policy := createGroupAndGroupPolicy(admin, s)

	specs := map[string]struct {
		req            *group.MsgUpdateGroupPolicyDecisionPolicy
		policy         group.DecisionPolicy
		expGroupPolicy *group.GroupPolicyInfo
		expErr         bool
	}{
		"with wrong admin": {
			req: &group.MsgUpdateGroupPolicyDecisionPolicy{
				Admin:   addr5.String(),
				Address: groupPolicyAddr,
			},
			policy:         policy,
			expGroupPolicy: &group.GroupPolicyInfo{},
			expErr:         true,
		},
		"with wrong group policy": {
			req: &group.MsgUpdateGroupPolicyDecisionPolicy{
				Admin:   admin.String(),
				Address: addr5.String(),
			},
			policy:         policy,
			expGroupPolicy: &group.GroupPolicyInfo{},
			expErr:         true,
		},
		"correct data": {
			req: &group.MsgUpdateGroupPolicyDecisionPolicy{
				Admin:   admin.String(),
				Address: groupPolicyAddr,
			},
			policy: group.NewThresholdDecisionPolicy(
				"2",
				time.Duration(2)*time.Second,
			),
			expGroupPolicy: &group.GroupPolicyInfo{
				Admin:          admin.String(),
				Address:        groupPolicyAddr,
				GroupId:        myGroupID,
				Metadata:       nil,
				Version:        2,
				DecisionPolicy: nil,
				CreatedAt:      s.blockTime,
			},
			expErr: false,
		},
	}
	for msg, spec := range specs {
		spec := spec
		err := spec.expGroupPolicy.SetDecisionPolicy(spec.policy)
		s.Require().NoError(err)

		err = spec.req.SetDecisionPolicy(spec.policy)
		s.Require().NoError(err)

		s.Run(msg, func() {
			_, err := s.keeper.UpdateGroupPolicyDecisionPolicy(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			res, err := s.keeper.GroupPolicyInfo(s.ctx, &group.QueryGroupPolicyInfoRequest{
				Address: groupPolicyAddr,
			})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expGroupPolicy, res.Info)
		})
	}
}

func (s *TestSuite) TestGroupPoliciesByAdminOrGroup() {
	addrs := s.addrs
	addr2 := addrs[1]

	admin := addr2
	groupRes, err := s.keeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    admin.String(),
		Members:  nil,
		Metadata: nil,
	})
	s.Require().NoError(err)
	myGroupID := groupRes.GroupId

	policies := []group.DecisionPolicy{
		group.NewThresholdDecisionPolicy(
			"1",
			time.Second,
		),
		group.NewThresholdDecisionPolicy(
			"10",
			time.Second,
		),
	}

	count := 2
	expectAccs := make([]*group.GroupPolicyInfo, count)
	for i := range expectAccs {
		req := &group.MsgCreateGroupPolicy{
			Admin:    admin.String(),
			Metadata: nil,
			GroupId:  myGroupID,
		}
		err := req.SetDecisionPolicy(policies[i])
		s.Require().NoError(err)
		res, err := s.keeper.CreateGroupPolicy(s.ctx, req)
		s.Require().NoError(err)

		expectAcc := &group.GroupPolicyInfo{
			Address:   res.Address,
			Admin:     admin.String(),
			Metadata:  nil,
			GroupId:   myGroupID,
			Version:   uint64(1),
			CreatedAt: s.blockTime,
		}
		err = expectAcc.SetDecisionPolicy(policies[i])
		s.Require().NoError(err)
		expectAccs[i] = expectAcc
	}
	sort.Slice(expectAccs, func(i, j int) bool { return expectAccs[i].Address < expectAccs[j].Address })

	// query group policy by group
	policiesByGroupRes, err := s.keeper.GroupPoliciesByGroup(s.ctx, &group.QueryGroupPoliciesByGroupRequest{
		GroupId: myGroupID,
	})
	s.Require().NoError(err)
	policyAccs := policiesByGroupRes.GroupPolicies
	s.Require().Equal(len(policyAccs), count)
	// we reorder policyAccs by address to be able to compare them
	sort.Slice(policyAccs, func(i, j int) bool { return policyAccs[i].Address < policyAccs[j].Address })
	for i := range policyAccs {
		s.Assert().Equal(policyAccs[i].Address, expectAccs[i].Address)
		s.Assert().Equal(policyAccs[i].GroupId, expectAccs[i].GroupId)
		s.Assert().Equal(policyAccs[i].Admin, expectAccs[i].Admin)
		s.Assert().Equal(policyAccs[i].Metadata, expectAccs[i].Metadata)
		s.Assert().Equal(policyAccs[i].Version, expectAccs[i].Version)
		s.Assert().Equal(policyAccs[i].CreatedAt, expectAccs[i].CreatedAt)
		s.Assert().Equal(policyAccs[i].GetDecisionPolicy(), expectAccs[i].GetDecisionPolicy())
	}

	// query group policy by admin
	policiesByAdminRes, err := s.keeper.GroupPoliciesByAdmin(s.ctx, &group.QueryGroupPoliciesByAdminRequest{
		Admin: admin.String(),
	})
	s.Require().NoError(err)
	policyAccs = policiesByAdminRes.GroupPolicies
	s.Require().Equal(len(policyAccs), count)
	// we reorder policyAccs by address to be able to compare them
	sort.Slice(policyAccs, func(i, j int) bool { return policyAccs[i].Address < policyAccs[j].Address })
	for i := range policyAccs {
		s.Assert().Equal(policyAccs[i].Address, expectAccs[i].Address)
		s.Assert().Equal(policyAccs[i].GroupId, expectAccs[i].GroupId)
		s.Assert().Equal(policyAccs[i].Admin, expectAccs[i].Admin)
		s.Assert().Equal(policyAccs[i].Metadata, expectAccs[i].Metadata)
		s.Assert().Equal(policyAccs[i].Version, expectAccs[i].Version)
		s.Assert().Equal(policyAccs[i].CreatedAt, expectAccs[i].CreatedAt)
		s.Assert().Equal(policyAccs[i].GetDecisionPolicy(), expectAccs[i].GetDecisionPolicy())
	}
}

func (s *TestSuite) TestCreateProposal() {
	addrs := s.addrs
	addr1 := addrs[0]
	addr2 := addrs[1]
	addr4 := addrs[3]
	addr5 := addrs[4]

	myGroupID := s.groupID
	accountAddr := s.groupPolicyAddr

	msgSend := &banktypes.MsgSend{
		FromAddress: s.groupPolicyAddr.String(),
		ToAddress:   addr2.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}

	policyReq := &group.MsgCreateGroupPolicy{
		Admin:    addr1.String(),
		GroupId:  myGroupID,
		Metadata: nil,
	}
	policy := group.NewThresholdDecisionPolicy(
		"100",
		time.Second,
	)
	err := policyReq.SetDecisionPolicy(policy)
	s.Require().NoError(err)
	bigThresholdRes, err := s.keeper.CreateGroupPolicy(s.ctx, policyReq)
	s.Require().NoError(err)
	bigThresholdAddr := bigThresholdRes.Address

	defaultProposal := group.Proposal{
		Status: group.ProposalStatusSubmitted,
		Result: group.ProposalResultUnfinalized,
		VoteState: group.Tally{
			YesCount:     "0",
			NoCount:      "0",
			AbstainCount: "0",
			VetoCount:    "0",
		},
		ExecutorResult: group.ProposalExecutorResultNotRun,
	}
	specs := map[string]struct {
		req         *group.MsgCreateProposal
		msgs        []sdk.Msg
		expProposal group.Proposal
		expErr      bool
		postRun     func(sdkCtx sdk.Context)
	}{
		"all good with minimal fields set": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{addr2.String()},
			},
			expProposal: defaultProposal,
			postRun:     func(sdkCtx sdk.Context) {},
		},
		"all good with good msg payload": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{addr2.String()},
			},
			msgs: []sdk.Msg{&banktypes.MsgSend{
				FromAddress: accountAddr.String(),
				ToAddress:   addr2.String(),
				Amount:      sdk.Coins{sdk.NewInt64Coin("token", 100)},
			}},
			expProposal: defaultProposal,
			postRun:     func(sdkCtx sdk.Context) {},
		},
		"metadata too long": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Metadata:  bytes.Repeat([]byte{1}, 256),
				Proposers: []string{addr2.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"group policy required": {
			req: &group.MsgCreateProposal{
				Metadata:  nil,
				Proposers: []string{addr2.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"existing group policy required": {
			req: &group.MsgCreateProposal{
				Address:   addr1.String(),
				Proposers: []string{addr2.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"impossible case: decision policy threshold > total group weight": {
			req: &group.MsgCreateProposal{
				Address:   bigThresholdAddr,
				Proposers: []string{addr2.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"only group members can create a proposal": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{addr4.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"all proposers must be in group": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{addr2.String(), addr4.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"admin that is not a group member can not create proposal": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Metadata:  nil,
				Proposers: []string{addr1.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"reject msgs that are not authz by group policy": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Metadata:  nil,
				Proposers: []string{addr2.String()},
			},
			msgs:    []sdk.Msg{&testdata.TestMsg{Signers: []string{addr1.String()}}},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"with try exec": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{addr2.String()},
				Exec:      group.Exec_EXEC_TRY,
			},
			msgs: []sdk.Msg{msgSend},
			expProposal: group.Proposal{
				Status: group.ProposalStatusClosed,
				Result: group.ProposalResultAccepted,
				VoteState: group.Tally{
					YesCount:     "2",
					NoCount:      "0",
					AbstainCount: "0",
					VetoCount:    "0",
				},
				ExecutorResult: group.ProposalExecutorResultSuccess,
			},
			postRun: func(sdkCtx sdk.Context) {
				fromBalances := s.app.BankKeeper.GetAllBalances(sdkCtx, accountAddr)
				s.Require().Contains(fromBalances, sdk.NewInt64Coin("test", 9900))
				toBalances := s.app.BankKeeper.GetAllBalances(sdkCtx, addr2)
				s.Require().Contains(toBalances, sdk.NewInt64Coin("test", 100))
			},
		},
		"with try exec, not enough yes votes for proposal to pass": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{addr5.String()},
				Exec:      group.Exec_EXEC_TRY,
			},
			msgs: []sdk.Msg{msgSend},
			expProposal: group.Proposal{
				Status: group.ProposalStatusSubmitted,
				Result: group.ProposalResultUnfinalized,
				VoteState: group.Tally{
					YesCount:     "1",
					NoCount:      "0",
					AbstainCount: "0",
					VetoCount:    "0",
				},
				ExecutorResult: group.ProposalExecutorResultNotRun,
			},
			postRun: func(sdkCtx sdk.Context) {},
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			err := spec.req.SetMsgs(spec.msgs)
			s.Require().NoError(err)

			res, err := s.keeper.CreateProposal(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			id := res.ProposalId

			// then all data persisted
			proposalRes, err := s.keeper.Proposal(s.ctx, &group.QueryProposalRequest{ProposalId: id})
			s.Require().NoError(err)
			proposal := proposalRes.Proposal

			s.Assert().Equal(accountAddr.String(), proposal.Address)
			s.Assert().Equal(spec.req.Metadata, proposal.Metadata)
			s.Assert().Equal(spec.req.Proposers, proposal.Proposers)
			s.Assert().Equal(s.blockTime, proposal.SubmittedAt)
			s.Assert().Equal(uint64(1), proposal.GroupVersion)
			s.Assert().Equal(uint64(1), proposal.GroupPolicyVersion)
			s.Assert().Equal(spec.expProposal.Status, proposal.Status)
			s.Assert().Equal(spec.expProposal.Result, proposal.Result)
			s.Assert().Equal(spec.expProposal.VoteState, proposal.VoteState)
			s.Assert().Equal(spec.expProposal.ExecutorResult, proposal.ExecutorResult)
			s.Assert().Equal(s.blockTime.Add(time.Second), proposal.Timeout)

			if spec.msgs == nil { // then empty list is ok
				s.Assert().Len(proposal.GetMsgs(), 0)
			} else {
				s.Assert().Equal(spec.msgs, proposal.GetMsgs())
			}

			spec.postRun(s.sdkCtx)
		})
	}
}

func (s *TestSuite) TestWithdrawProposal() {
	addrs := s.addrs
	addr2 := addrs[1]
	addr5 := addrs[4]
	groupPolicy := s.groupPolicyAddr

	msgSend := &banktypes.MsgSend{
		FromAddress: s.groupPolicyAddr.String(),
		ToAddress:   addr2.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}

	proposers := []string{addr2.String()}
	proposalID := createProposal(s.ctx, s, []sdk.Msg{msgSend}, proposers)

	specs := map[string]struct {
		preRun     func(sdkCtx sdk.Context) uint64
		proposalId uint64
		admin      string
		expErrMsg  string
	}{
		"wrong admin": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				return createProposal(s.ctx, s, []sdk.Msg{msgSend}, proposers)
			},
			admin:     addr5.String(),
			expErrMsg: "unauthorized",
		},
		"wrong proposalId": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				return 1111
			},
			admin:     proposers[0],
			expErrMsg: "not found",
		},
		"happy case with proposer": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				return createProposal(s.ctx, s, []sdk.Msg{msgSend}, proposers)
			},
			proposalId: proposalID,
			admin:      proposers[0],
		},
		"already closed proposal": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				pId := createProposal(s.ctx, s, []sdk.Msg{msgSend}, proposers)
				_, err := s.keeper.WithdrawProposal(s.ctx, &group.MsgWithdrawProposal{
					ProposalId: pId,
					Address:    proposers[0],
				})
				s.Require().NoError(err)
				return pId
			},
			proposalId: proposalID,
			admin:      proposers[0],
			expErrMsg:  "cannot withdraw a proposal with the status of STATUS_WITHDRAWN",
		},
		"happy case with group admin address": {
			preRun: func(sdkCtx sdk.Context) uint64 {
				return createProposal(s.ctx, s, []sdk.Msg{msgSend}, proposers)
			},
			proposalId: proposalID,
			admin:      groupPolicy.String(),
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			pId := spec.preRun(s.sdkCtx)

			_, err := s.keeper.WithdrawProposal(s.ctx, &group.MsgWithdrawProposal{
				ProposalId: pId,
				Address:    spec.admin,
			})

			if spec.expErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), spec.expErrMsg)
				return
			}

			s.Require().NoError(err)
			resp, err := s.keeper.Proposal(s.ctx, &group.QueryProposalRequest{ProposalId: pId})
			s.Require().NoError(err)
			s.Require().Equal(resp.GetProposal().Status, group.ProposalStatusWithdrawn)
		})
	}
}

func (s *TestSuite) TestVote() {
	addrs := s.addrs
	addr1 := addrs[0]
	addr2 := addrs[1]
	addr3 := addrs[2]
	addr4 := addrs[3]
	addr5 := addrs[4]
	members := []group.Member{
		{Address: addr4.String(), Weight: "1", AddedAt: s.blockTime},
		{Address: addr3.String(), Weight: "2", AddedAt: s.blockTime},
	}
	groupRes, err := s.keeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    addr1.String(),
		Members:  members,
		Metadata: nil,
	})
	s.Require().NoError(err)
	myGroupID := groupRes.GroupId

	policy := group.NewThresholdDecisionPolicy(
		"2",
		time.Duration(2),
	)
	policyReq := &group.MsgCreateGroupPolicy{
		Admin:    addr1.String(),
		GroupId:  myGroupID,
		Metadata: nil,
	}
	err = policyReq.SetDecisionPolicy(policy)
	s.Require().NoError(err)
	policyRes, err := s.keeper.CreateGroupPolicy(s.ctx, policyReq)
	s.Require().NoError(err)
	accountAddr := policyRes.Address
	groupPolicy, err := sdk.AccAddressFromBech32(accountAddr)
	s.Require().NoError(err)
	s.Require().NotNil(groupPolicy)

	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.sdkCtx, groupPolicy, sdk.Coins{sdk.NewInt64Coin("test", 10000)}))

	req := &group.MsgCreateProposal{
		Address:   accountAddr,
		Metadata:  nil,
		Proposers: []string{addr4.String()},
		Msgs:      nil,
	}
	err = req.SetMsgs([]sdk.Msg{&banktypes.MsgSend{
		FromAddress: accountAddr,
		ToAddress:   addr5.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}})
	s.Require().NoError(err)

	proposalRes, err := s.keeper.CreateProposal(s.ctx, req)
	s.Require().NoError(err)
	myProposalID := proposalRes.ProposalId

	// proposals by group policy
	proposalsRes, err := s.keeper.ProposalsByGroupPolicy(s.ctx, &group.QueryProposalsByGroupPolicyRequest{
		Address: accountAddr,
	})
	s.Require().NoError(err)
	proposals := proposalsRes.Proposals
	s.Require().Equal(len(proposals), 1)
	s.Assert().Equal(req.Address, proposals[0].Address)
	s.Assert().Equal(req.Metadata, proposals[0].Metadata)
	s.Assert().Equal(req.Proposers, proposals[0].Proposers)
	s.Assert().Equal(s.blockTime, proposals[0].SubmittedAt)

	s.Assert().Equal(uint64(1), proposals[0].GroupVersion)
	s.Assert().Equal(uint64(1), proposals[0].GroupPolicyVersion)
	s.Assert().Equal(group.ProposalStatusSubmitted, proposals[0].Status)
	s.Assert().Equal(group.ProposalResultUnfinalized, proposals[0].Result)
	s.Assert().Equal(group.Tally{
		YesCount:     "0",
		NoCount:      "0",
		AbstainCount: "0",
		VetoCount:    "0",
	}, proposals[0].VoteState)

	specs := map[string]struct {
		srcCtx            sdk.Context
		expVoteState      group.Tally
		req               *group.MsgVote
		doBefore          func(ctx context.Context)
		postRun           func(sdkCtx sdk.Context)
		expProposalStatus group.Proposal_Status
		expResult         group.Proposal_Result
		expExecutorResult group.Proposal_ExecutorResult
		expErr            bool
	}{
		"vote yes": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Choice:     group.Choice_CHOICE_YES,
			},
			expVoteState: group.Tally{
				YesCount:     "1",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusSubmitted,
			expResult:         group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun:           func(sdkCtx sdk.Context) {},
		},
		"with try exec": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr3.String(),
				Choice:     group.Choice_CHOICE_YES,
				Exec:       group.Exec_EXEC_TRY,
			},
			expVoteState: group.Tally{
				YesCount:     "2",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusClosed,
			expResult:         group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultSuccess,
			postRun: func(sdkCtx sdk.Context) {
				fromBalances := s.app.BankKeeper.GetAllBalances(sdkCtx, groupPolicy)
				s.Require().Contains(fromBalances, sdk.NewInt64Coin("test", 9900))
				toBalances := s.app.BankKeeper.GetAllBalances(sdkCtx, addr5)
				s.Require().Contains(toBalances, sdk.NewInt64Coin("test", 100))
			},
		},
		"with try exec, not enough yes votes for proposal to pass": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Choice:     group.Choice_CHOICE_YES,
				Exec:       group.Exec_EXEC_TRY,
			},
			expVoteState: group.Tally{
				YesCount:     "1",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusSubmitted,
			expResult:         group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun:           func(sdkCtx sdk.Context) {},
		},
		"vote no": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			expVoteState: group.Tally{
				YesCount:     "0",
				NoCount:      "1",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusSubmitted,
			expResult:         group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun:           func(sdkCtx sdk.Context) {},
		},
		"vote abstain": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Choice:     group.Choice_CHOICE_ABSTAIN,
			},
			expVoteState: group.Tally{
				YesCount:     "0",
				NoCount:      "0",
				AbstainCount: "1",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusSubmitted,
			expResult:         group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun:           func(sdkCtx sdk.Context) {},
		},
		"vote veto": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Choice:     group.Choice_CHOICE_VETO,
			},
			expVoteState: group.Tally{
				YesCount:     "0",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "1",
			},
			expProposalStatus: group.ProposalStatusSubmitted,
			expResult:         group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun:           func(sdkCtx sdk.Context) {},
		},
		"apply decision policy early": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr3.String(),
				Choice:     group.Choice_CHOICE_YES,
			},
			expVoteState: group.Tally{
				YesCount:     "2",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusClosed,
			expResult:         group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun:           func(sdkCtx sdk.Context) {},
		},
		"reject new votes when final decision is made already": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Choice:     group.Choice_CHOICE_YES,
			},
			doBefore: func(ctx context.Context) {
				_, err := s.keeper.Vote(ctx, &group.MsgVote{
					ProposalId: myProposalID,
					Voter:      addr3.String(),
					Choice:     group.Choice_CHOICE_VETO,
				})
				s.Require().NoError(err)
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"metadata too long": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Metadata:   bytes.Repeat([]byte{1}, 256),
				Choice:     group.Choice_CHOICE_NO,
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"existing proposal required": {
			req: &group.MsgVote{
				ProposalId: 999,
				Voter:      addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"empty choice": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"invalid choice": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Choice:     5,
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"voter must be in group": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr2.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"admin that is not a group member can not vote": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr1.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"on timeout": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			srcCtx:  s.sdkCtx.WithBlockTime(s.blockTime.Add(time.Second)),
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"closed already": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			doBefore: func(ctx context.Context) {
				_, err := s.keeper.Vote(ctx, &group.MsgVote{
					ProposalId: myProposalID,
					Voter:      addr3.String(),
					Choice:     group.Choice_CHOICE_YES,
				})
				s.Require().NoError(err)
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"voted already": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			doBefore: func(ctx context.Context) {
				_, err := s.keeper.Vote(ctx, &group.MsgVote{
					ProposalId: myProposalID,
					Voter:      addr4.String(),
					Choice:     group.Choice_CHOICE_YES,
				})
				s.Require().NoError(err)
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"with group modified": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			doBefore: func(ctx context.Context) {
				_, err = s.keeper.UpdateGroupMetadata(ctx, &group.MsgUpdateGroupMetadata{
					GroupId:  myGroupID,
					Admin:    addr1.String(),
					Metadata: []byte{1, 2, 3},
				})
				s.Require().NoError(err)
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"with policy modified": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			doBefore: func(ctx context.Context) {
				m, err := group.NewMsgUpdateGroupPolicyDecisionPolicyRequest(
					addr1,
					groupPolicy,
					&group.ThresholdDecisionPolicy{
						Threshold: "1",
						Timeout:   time.Second,
					},
				)
				s.Require().NoError(err)

				_, err = s.keeper.UpdateGroupPolicyDecisionPolicy(ctx, m)
				s.Require().NoError(err)
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			sdkCtx := s.sdkCtx
			if !spec.srcCtx.IsZero() {
				sdkCtx = spec.srcCtx
			}
			sdkCtx, _ = sdkCtx.CacheContext()
			ctx := sdk.WrapSDKContext(sdkCtx)

			if spec.doBefore != nil {
				spec.doBefore(ctx)
			}
			_, err := s.keeper.Vote(ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			s.Require().NoError(err)
			// vote is stored and all data persisted
			res, err := s.keeper.VoteByProposalVoter(ctx, &group.QueryVoteByProposalVoterRequest{
				ProposalId: spec.req.ProposalId,
				Voter:      spec.req.Voter,
			})
			s.Require().NoError(err)
			loaded := res.Vote
			s.Assert().Equal(spec.req.ProposalId, loaded.ProposalId)
			s.Assert().Equal(spec.req.Voter, loaded.Voter)
			s.Assert().Equal(spec.req.Choice, loaded.Choice)
			s.Assert().Equal(spec.req.Metadata, loaded.Metadata)
			s.Assert().Equal(s.blockTime, loaded.SubmittedAt)

			// query votes by proposal
			votesByProposalRes, err := s.keeper.VotesByProposal(ctx, &group.QueryVotesByProposalRequest{
				ProposalId: spec.req.ProposalId,
			})
			s.Require().NoError(err)
			votesByProposal := votesByProposalRes.Votes
			s.Require().Equal(1, len(votesByProposal))
			vote := votesByProposal[0]
			s.Assert().Equal(spec.req.ProposalId, vote.ProposalId)
			s.Assert().Equal(spec.req.Voter, vote.Voter)
			s.Assert().Equal(spec.req.Choice, vote.Choice)
			s.Assert().Equal(spec.req.Metadata, vote.Metadata)
			s.Assert().Equal(s.blockTime, vote.SubmittedAt)

			// query votes by voter
			voter := spec.req.Voter
			votesByVoterRes, err := s.keeper.VotesByVoter(ctx, &group.QueryVotesByVoterRequest{
				Voter: voter,
			})
			s.Require().NoError(err)
			votesByVoter := votesByVoterRes.Votes
			s.Require().Equal(1, len(votesByVoter))
			s.Assert().Equal(spec.req.ProposalId, votesByVoter[0].ProposalId)
			s.Assert().Equal(voter, votesByVoter[0].Voter)
			s.Assert().Equal(spec.req.Choice, votesByVoter[0].Choice)
			s.Assert().Equal(spec.req.Metadata, votesByVoter[0].Metadata)
			s.Assert().Equal(s.blockTime, votesByVoter[0].SubmittedAt)

			// and proposal is updated
			proposalRes, err := s.keeper.Proposal(ctx, &group.QueryProposalRequest{
				ProposalId: spec.req.ProposalId,
			})
			s.Require().NoError(err)
			proposal := proposalRes.Proposal
			s.Assert().Equal(spec.expVoteState, proposal.VoteState)
			s.Assert().Equal(spec.expResult, proposal.Result)
			s.Assert().Equal(spec.expProposalStatus, proposal.Status)
			s.Assert().Equal(spec.expExecutorResult, proposal.ExecutorResult)

			spec.postRun(sdkCtx)
		})
	}
}

func (s *TestSuite) TestExecProposal() {
	addrs := s.addrs
	addr1 := addrs[0]
	addr2 := addrs[1]

	msgSend1 := &banktypes.MsgSend{
		FromAddress: s.groupPolicyAddr.String(),
		ToAddress:   addr2.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}
	msgSend2 := &banktypes.MsgSend{
		FromAddress: s.groupPolicyAddr.String(),
		ToAddress:   addr2.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 10001)},
	}
	proposers := []string{addr2.String()}

	specs := map[string]struct {
		srcBlockTime      time.Time
		setupProposal     func(ctx context.Context) uint64
		expErr            bool
		expProposalStatus group.Proposal_Status
		expProposalResult group.Proposal_Result
		expExecutorResult group.Proposal_ExecutorResult
		expBalance        bool
		expFromBalances   sdk.Coin
		expToBalances     sdk.Coin
	}{
		"proposal executed when accepted": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				return createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_YES)
			},
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultSuccess,
			expBalance:        true,
			expFromBalances:   sdk.NewInt64Coin("test", 9900),
			expToBalances:     sdk.NewInt64Coin("test", 100),
		},
		"proposal with multiple messages executed when accepted": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1, msgSend1}
				return createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_YES)
			},
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultSuccess,
			expBalance:        true,
			expFromBalances:   sdk.NewInt64Coin("test", 9800),
			expToBalances:     sdk.NewInt64Coin("test", 200),
		},
		"proposal not executed when rejected": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				return createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_NO)
			},
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultRejected,
			expExecutorResult: group.ProposalExecutorResultNotRun,
		},
		"open proposal must not fail": {
			setupProposal: func(ctx context.Context) uint64 {
				return createProposal(ctx, s, []sdk.Msg{msgSend1}, proposers)
			},
			expProposalStatus: group.ProposalStatusSubmitted,
			expProposalResult: group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
		},
		"existing proposal required": {
			setupProposal: func(ctx context.Context) uint64 {
				return 9999
			},
			expErr: true,
		},
		"Decision policy also applied on timeout": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				return createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_NO)
			},
			srcBlockTime:      s.blockTime.Add(time.Second),
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultRejected,
			expExecutorResult: group.ProposalExecutorResultNotRun,
		},
		"Decision policy also applied after timeout": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				return createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_NO)
			},
			srcBlockTime:      s.blockTime.Add(time.Second).Add(time.Millisecond),
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultRejected,
			expExecutorResult: group.ProposalExecutorResultNotRun,
		},
		"with group modified before tally": {
			setupProposal: func(ctx context.Context) uint64 {
				myProposalID := createProposal(ctx, s, []sdk.Msg{msgSend1}, proposers)

				// then modify group
				_, err := s.keeper.UpdateGroupMetadata(ctx, &group.MsgUpdateGroupMetadata{
					Admin:    addr1.String(),
					GroupId:  s.groupID,
					Metadata: []byte{1, 2, 3},
				})
				s.Require().NoError(err)
				return myProposalID
			},
			expProposalStatus: group.ProposalStatusAborted,
			expProposalResult: group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
		},
		"with group policy modified before tally": {
			setupProposal: func(ctx context.Context) uint64 {
				myProposalID := createProposal(ctx, s, []sdk.Msg{msgSend1}, proposers)
				_, err := s.keeper.UpdateGroupPolicyMetadata(ctx, &group.MsgUpdateGroupPolicyMetadata{
					Admin:    addr1.String(),
					Address:  s.groupPolicyAddr.String(),
					Metadata: []byte("group policy modified before tally"),
				})
				s.Require().NoError(err)
				return myProposalID
			},
			expProposalStatus: group.ProposalStatusAborted,
			expProposalResult: group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
		},
		"prevent double execution when successful": {
			setupProposal: func(ctx context.Context) uint64 {
				myProposalID := createProposalAndVote(ctx, s, []sdk.Msg{msgSend1}, proposers, group.Choice_CHOICE_YES)

				_, err := s.keeper.Exec(ctx, &group.MsgExec{Signer: addr1.String(), ProposalId: myProposalID})
				s.Require().NoError(err)
				return myProposalID
			},
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultSuccess,
			expBalance:        true,
			expFromBalances:   sdk.NewInt64Coin("test", 9900),
			expToBalances:     sdk.NewInt64Coin("test", 100),
		},
		"rollback all msg updates on failure": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1, msgSend2}
				return createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_YES)
			},
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultFailure,
		},
		"executable when failed before": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend2}
				myProposalID := createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_YES)

				_, err := s.keeper.Exec(ctx, &group.MsgExec{Signer: addr1.String(), ProposalId: myProposalID})
				s.Require().NoError(err)
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, sdkCtx, s.groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("test", 10002)}))

				return myProposalID
			},
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultSuccess,
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			sdkCtx, _ := s.sdkCtx.CacheContext()
			ctx := sdk.WrapSDKContext(sdkCtx)
			proposalID := spec.setupProposal(ctx)

			if !spec.srcBlockTime.IsZero() {
				sdkCtx = sdkCtx.WithBlockTime(spec.srcBlockTime)
			}

			ctx = sdk.WrapSDKContext(sdkCtx)
			_, err := s.keeper.Exec(ctx, &group.MsgExec{Signer: addr1.String(), ProposalId: proposalID})
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// and proposal is updated
			res, err := s.keeper.Proposal(ctx, &group.QueryProposalRequest{ProposalId: proposalID})
			s.Require().NoError(err)
			proposal := res.Proposal

			exp := group.Proposal_Result_name[int32(spec.expProposalResult)]
			got := group.Proposal_Result_name[int32(proposal.Result)]
			s.Assert().Equal(exp, got)

			exp = group.Proposal_Status_name[int32(spec.expProposalStatus)]
			got = group.Proposal_Status_name[int32(proposal.Status)]
			s.Assert().Equal(exp, got)

			exp = group.Proposal_ExecutorResult_name[int32(spec.expExecutorResult)]
			got = group.Proposal_ExecutorResult_name[int32(proposal.ExecutorResult)]
			s.Assert().Equal(exp, got)

			if spec.expBalance {
				fromBalances := s.app.BankKeeper.GetAllBalances(sdkCtx, s.groupPolicyAddr)
				s.Require().Contains(fromBalances, spec.expFromBalances)
				toBalances := s.app.BankKeeper.GetAllBalances(sdkCtx, addr2)
				s.Require().Contains(toBalances, spec.expToBalances)
			}
		})
	}
}

func createProposal(
	ctx context.Context, s *TestSuite, msgs []sdk.Msg,
	proposers []string) uint64 {
	proposalReq := &group.MsgCreateProposal{
		Address:   s.groupPolicyAddr.String(),
		Proposers: proposers,
		Metadata:  nil,
	}
	err := proposalReq.SetMsgs(msgs)
	s.Require().NoError(err)

	proposalRes, err := s.keeper.CreateProposal(ctx, proposalReq)
	s.Require().NoError(err)
	return proposalRes.ProposalId
}

func createProposalAndVote(
	ctx context.Context, s *TestSuite, msgs []sdk.Msg,
	proposers []string, choice group.Choice) uint64 {
	s.Require().Greater(len(proposers), 0)
	myProposalID := createProposal(ctx, s, msgs, proposers)

	_, err := s.keeper.Vote(ctx, &group.MsgVote{
		ProposalId: myProposalID,
		Voter:      proposers[0],
		Choice:     choice,
	})
	s.Require().NoError(err)
	return myProposalID
}

func createGroupAndGroupPolicy(
	admin sdk.AccAddress,
	s *TestSuite,
) (string, uint64, group.DecisionPolicy) {
	groupRes, err := s.keeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    admin.String(),
		Members:  nil,
		Metadata: nil,
	})
	s.Require().NoError(err)

	myGroupID := groupRes.GroupId
	groupPolicy := &group.MsgCreateGroupPolicy{
		Admin:    admin.String(),
		GroupId:  myGroupID,
		Metadata: nil,
	}

	policy := group.NewThresholdDecisionPolicy(
		"1",
		time.Second,
	)
	err = groupPolicy.SetDecisionPolicy(policy)
	s.Require().NoError(err)

	groupPolicyRes, err := s.keeper.CreateGroupPolicy(s.ctx, groupPolicy)
	s.Require().NoError(err)

	return groupPolicyRes.Address, myGroupID, policy
}
