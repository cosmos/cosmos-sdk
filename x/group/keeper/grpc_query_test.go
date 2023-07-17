package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	"github.com/cosmos/cosmos-sdk/x/group/module"
	grouptestutil "github.com/cosmos/cosmos-sdk/x/group/testutil"
)

func initKeeper(t *testing.T) (types.Context, groupkeeper.Keeper, []types.AccAddress, group.QueryClient) {
	t.Helper()
	var (
		groupKeeper       groupkeeper.Keeper
		interfaceRegistry codectypes.InterfaceRegistry
	)

	key := storetypes.NewKVStoreKey(group.StoreKey)
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

	groupKeeper = groupkeeper.NewKeeper(key, encCfg.Codec, bApp.MsgServiceRouter(), accountKeeper, group.DefaultConfig())
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry)
	group.RegisterQueryServer(queryHelper, groupKeeper)
	queryClient := group.NewQueryClient(queryHelper)

	msgGroupAndPolicy := &group.MsgCreateGroupWithPolicy{
		Admin: addrs[0].String(),
		Members: []group.MemberRequest{
			{Address: addrs[1].String(), Weight: "1"},
			{Address: addrs[3].String(), Weight: "2"},
		},
	}
	msgGroupAndPolicy.SetDecisionPolicy(group.NewThresholdDecisionPolicy("2", time.Second, 20))

	_, err := groupKeeper.CreateGroupWithPolicy(ctx, msgGroupAndPolicy)
	require.NoError(t, err)

	return ctx, groupKeeper, addrs, queryClient
}

func TestQueryGroupInfo(t *testing.T) {
	ctx, _, _, queryClient := initKeeper(t)

	testCases := []struct {
		name      string
		groupID   uint64
		expErrMsg string
	}{
		{
			name:      "unknown id",
			groupID:   20,
			expErrMsg: "group: not found",
		},
		{
			name:      "valid id",
			groupID:   1,
			expErrMsg: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := queryClient.GroupInfo(ctx, &group.QueryGroupInfoRequest{GroupId: tc.groupID})
			if tc.expErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestQueryGroupsByMember(t *testing.T) {
	ctx, groupKeeper, addrs, queryClient := initKeeper(t)

	members := []group.MemberRequest{
		{Address: addrs[3].String(), Weight: "1"}, {Address: addrs[4].String(), Weight: "2"},
	}
	_, err := groupKeeper.CreateGroup(ctx, &group.MsgCreateGroup{
		Admin:   addrs[1].String(),
		Members: members,
	})
	require.NoError(t, err)

	// not part of any group
	resp, err := queryClient.GroupsByMember(context.Background(), &group.QueryGroupsByMemberRequest{
		Address: addrs[5].String(),
	})
	require.NoError(t, err)
	require.Len(t, resp.Groups, 0)

	// expect one group
	resp, err = queryClient.GroupsByMember(context.Background(), &group.QueryGroupsByMemberRequest{
		Address: addrs[4].String(),
	})
	require.NoError(t, err)
	require.Len(t, resp.Groups, 1)

	// expect two groups
	resp, err = queryClient.GroupsByMember(context.Background(), &group.QueryGroupsByMemberRequest{
		Address: addrs[3].String(),
	})
	require.NoError(t, err)
	require.Len(t, resp.Groups, 2)
}

func TestQueryGroups(t *testing.T) {
	ctx, groupKeeper, addrs, queryClient := initKeeper(t)

	members := []group.MemberRequest{
		{Address: addrs[3].String(), Weight: "1"},
	}
	_, err := groupKeeper.CreateGroup(ctx, &group.MsgCreateGroup{
		Admin:   addrs[2].String(),
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
			resp, err := queryClient.Groups(context.Background(), &group.QueryGroupsRequest{
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
