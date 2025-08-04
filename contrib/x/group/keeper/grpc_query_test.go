package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	group2 "github.com/cosmos/cosmos-sdk/contrib/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/contrib/x/group/keeper"
	"github.com/cosmos/cosmos-sdk/contrib/x/group/module"
	grouptestutil "github.com/cosmos/cosmos-sdk/contrib/x/group/testutil"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type fixture struct {
	ctx          types.Context
	keeper       groupkeeper.Keeper
	queryClient  group2.QueryClient
	addrs        []types.AccAddress
	defaultGroup *group2.MsgCreateGroupWithPolicyResponse
}

func initKeeper(t *testing.T) *fixture {
	t.Helper()
	var (
		groupKeeper       groupkeeper.Keeper
		interfaceRegistry codectypes.InterfaceRegistry
	)

	key := storetypes.NewKVStoreKey(group2.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})

	ctx := testCtx.Ctx

	bApp := baseapp.NewBaseApp(
		"group",
		log.NewNopLogger(),
		testCtx.DB,
		encCfg.TxConfig.TxDecoder(),
	)

	addrs := simtestutil.CreateIncrementalAccounts(6)
	ctrl := gomock.NewController(t)
	accountKeeper := grouptestutil.NewMockAccountKeeper(ctrl)
	for _, addr := range addrs {
		accountKeeper.EXPECT().GetAccount(gomock.Any(), addr).Return(authtypes.NewBaseAccountWithAddress(addr)).AnyTimes()
	}
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	// group policy expected calls
	accountKeeper.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	accountKeeper.EXPECT().NewAccount(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	accountKeeper.EXPECT().SetAccount(gomock.Any(), gomock.Any()).AnyTimes()

	groupKeeper = groupkeeper.NewKeeper(key, encCfg.Codec, bApp.MsgServiceRouter(), accountKeeper, group2.DefaultConfig())
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry)
	group2.RegisterQueryServer(queryHelper, groupKeeper)
	queryClient := group2.NewQueryClient(queryHelper)

	msgGroupAndPolicy := &group2.MsgCreateGroupWithPolicy{
		Admin: addrs[0].String(),
		Members: []group2.MemberRequest{
			{Address: addrs[1].String(), Weight: "1"},
			{Address: addrs[3].String(), Weight: "2"},
		},
	}
	err := msgGroupAndPolicy.SetDecisionPolicy(group2.NewThresholdDecisionPolicy("2", time.Second, 20))
	require.NoError(t, err)

	resp, err := groupKeeper.CreateGroupWithPolicy(ctx, msgGroupAndPolicy)
	require.NoError(t, err)

	return &fixture{
		ctx:          ctx,
		keeper:       groupKeeper,
		queryClient:  queryClient,
		addrs:        addrs,
		defaultGroup: resp,
	}
}

func TestQueryGroupInfo(t *testing.T) {
	fixture := initKeeper(t)

	testCases := []struct {
		name      string
		req       group2.QueryGroupInfoRequest
		expErrMsg string
	}{
		{
			name:      "invalid req",
			expErrMsg: "group: not found",
		},
		{
			name:      "unknown id",
			req:       group2.QueryGroupInfoRequest{GroupId: 20},
			expErrMsg: "group: not found",
		},
		{
			name:      "valid id",
			req:       group2.QueryGroupInfoRequest{GroupId: 1},
			expErrMsg: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fixture.queryClient.GroupInfo(fixture.ctx, &tc.req)
			if tc.expErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestQueryGroupPolicyInfo(t *testing.T) {
	fixture := initKeeper(t)

	testCases := []struct {
		name      string
		req       group2.QueryGroupPolicyInfoRequest
		expErrMsg string
	}{
		{
			name:      "valid address",
			req:       group2.QueryGroupPolicyInfoRequest{Address: fixture.defaultGroup.GroupPolicyAddress},
			expErrMsg: "",
		},
		{
			name:      "unexisting address",
			req:       group2.QueryGroupPolicyInfoRequest{Address: fixture.addrs[5].String()},
			expErrMsg: "group policy: not found",
		},
		{
			name:      "invalid address",
			req:       group2.QueryGroupPolicyInfoRequest{Address: "invalid address"},
			expErrMsg: "decoding bech32 failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fixture.queryClient.GroupPolicyInfo(fixture.ctx, &tc.req)
			if tc.expErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestQueryGroupMembers(t *testing.T) {
	fixture := initKeeper(t)

	testCases := []struct {
		name      string
		req       group2.QueryGroupMembersRequest
		postRun   func(resp *group2.QueryGroupMembersResponse)
		expErrMsg string
	}{
		{
			name: "valid group",
			req:  group2.QueryGroupMembersRequest{GroupId: 1},
			postRun: func(resp *group2.QueryGroupMembersResponse) {
				require.Len(t, resp.Members, 2)
			},
			expErrMsg: "",
		},
		{
			name: "unexisting group",
			req:  group2.QueryGroupMembersRequest{GroupId: 20},
			postRun: func(resp *group2.QueryGroupMembersResponse) {
				require.Len(t, resp.Members, 0)
			},
			expErrMsg: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := fixture.queryClient.GroupMembers(fixture.ctx, &tc.req)
			if tc.expErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}

			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func TestQueryGroupsByAdmin(t *testing.T) {
	fixture := initKeeper(t)

	testCases := []struct {
		name      string
		req       group2.QueryGroupsByAdminRequest
		postRun   func(resp *group2.QueryGroupsByAdminResponse)
		expErrMsg string
	}{
		{
			name:      "valid admin",
			req:       group2.QueryGroupsByAdminRequest{Admin: fixture.addrs[0].String()},
			postRun:   func(resp *group2.QueryGroupsByAdminResponse) { require.Len(t, resp.Groups, 1) },
			expErrMsg: "",
		},
		{
			name:      "unexisting address",
			req:       group2.QueryGroupsByAdminRequest{Admin: fixture.addrs[5].String()},
			postRun:   func(resp *group2.QueryGroupsByAdminResponse) { require.Len(t, resp.Groups, 0) },
			expErrMsg: "",
		},
		{
			name:      "invalid address",
			req:       group2.QueryGroupsByAdminRequest{Admin: "invalid address"},
			expErrMsg: "decoding bech32 failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := fixture.queryClient.GroupsByAdmin(fixture.ctx, &tc.req)
			if tc.expErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}

			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func TestQueryGroupPoliciesByGroup(t *testing.T) {
	fixture := initKeeper(t)

	testCases := []struct {
		name      string
		req       group2.QueryGroupPoliciesByGroupRequest
		postRun   func(resp *group2.QueryGroupPoliciesByGroupResponse)
		expErrMsg string
	}{
		{
			name:      "valid group",
			req:       group2.QueryGroupPoliciesByGroupRequest{GroupId: 1},
			postRun:   func(resp *group2.QueryGroupPoliciesByGroupResponse) { require.Len(t, resp.GroupPolicies, 1) },
			expErrMsg: "",
		},
		{
			name:      "unexisting group",
			req:       group2.QueryGroupPoliciesByGroupRequest{GroupId: 20},
			postRun:   func(resp *group2.QueryGroupPoliciesByGroupResponse) { require.Len(t, resp.GroupPolicies, 0) },
			expErrMsg: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := fixture.keeper.GroupPoliciesByGroup(fixture.ctx, &tc.req)
			if tc.expErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}

			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func TestQueryGroupPoliciesByAdmin(t *testing.T) {
	fixture := initKeeper(t)

	testCases := []struct {
		name      string
		req       group2.QueryGroupPoliciesByAdminRequest
		postRun   func(resp *group2.QueryGroupPoliciesByAdminResponse)
		expErrMsg string
	}{
		{
			name:      "valid admin",
			req:       group2.QueryGroupPoliciesByAdminRequest{Admin: fixture.addrs[0].String()},
			postRun:   func(resp *group2.QueryGroupPoliciesByAdminResponse) { require.Len(t, resp.GroupPolicies, 1) },
			expErrMsg: "",
		},
		{
			name:      "unexisting address",
			req:       group2.QueryGroupPoliciesByAdminRequest{Admin: fixture.addrs[5].String()},
			postRun:   func(resp *group2.QueryGroupPoliciesByAdminResponse) { require.Len(t, resp.GroupPolicies, 0) },
			expErrMsg: "",
		},
		{
			name:      "invalid address",
			req:       group2.QueryGroupPoliciesByAdminRequest{Admin: "invalid address"},
			expErrMsg: "decoding bech32 failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := fixture.keeper.GroupPoliciesByAdmin(fixture.ctx, &tc.req)
			if tc.expErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}

			if tc.postRun != nil {
				tc.postRun(resp)
			}
		})
	}
}

func TestQueryGroupsByMember(t *testing.T) {
	fixture := initKeeper(t)

	members := []group2.MemberRequest{
		{Address: fixture.addrs[3].String(), Weight: "1"}, {Address: fixture.addrs[4].String(), Weight: "2"},
	}
	_, err := fixture.keeper.CreateGroup(fixture.ctx, &group2.MsgCreateGroup{
		Admin:   fixture.addrs[1].String(),
		Members: members,
	})
	require.NoError(t, err)

	// not part of any group
	resp, err := fixture.queryClient.GroupsByMember(context.Background(), &group2.QueryGroupsByMemberRequest{
		Address: fixture.addrs[5].String(),
	})
	require.NoError(t, err)
	require.Len(t, resp.Groups, 0)

	// expect one group
	resp, err = fixture.queryClient.GroupsByMember(context.Background(), &group2.QueryGroupsByMemberRequest{
		Address: fixture.addrs[4].String(),
	})
	require.NoError(t, err)
	require.Len(t, resp.Groups, 1)

	// expect two groups
	resp, err = fixture.queryClient.GroupsByMember(context.Background(), &group2.QueryGroupsByMemberRequest{
		Address: fixture.addrs[3].String(),
	})
	require.NoError(t, err)
	require.Len(t, resp.Groups, 2)
}

func TestQueryGroups(t *testing.T) {
	fixture := initKeeper(t)

	members := []group2.MemberRequest{
		{Address: fixture.addrs[3].String(), Weight: "1"},
	}
	_, err := fixture.keeper.CreateGroup(fixture.ctx, &group2.MsgCreateGroup{
		Admin:   fixture.addrs[2].String(),
		Members: members,
	})
	require.NoError(t, err)

	testCases := []struct {
		name         string
		expErr       bool
		expLen       int
		itemsPerPage uint64
	}{
		{
			name:         "success case, without pagination",
			expErr:       false,
			expLen:       2,
			itemsPerPage: 10,
		},
		{
			name:         "success case, with pagination",
			expErr:       false,
			expLen:       1,
			itemsPerPage: 1,
		},
		{
			name:   "success without pagination",
			expErr: false,
			expLen: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := fixture.queryClient.Groups(context.Background(), &group2.QueryGroupsRequest{
				Pagination: &query.PageRequest{
					Limit: tc.itemsPerPage,
				},
			})

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(resp.Groups), tc.expLen)
			}
		})
	}
}
