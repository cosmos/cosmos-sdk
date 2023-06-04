package keeper_test

import (
	"bytes"
	"testing"

	cmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/circuit/keeper"
	"cosmossdk.io/x/circuit/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type fixture struct {
	ctx        sdk.Context
	keeper     keeper.Keeper
	mockAddr   []byte
	mockPerms  types.Permissions
	mockMsgURL string
}

func initFixture(t *testing.T) *fixture {
	ac := addresscodec.NewBech32Codec("cosmos")
	mockStoreKey := storetypes.NewKVStoreKey("test")
	k := keeper.NewKeeper(mockStoreKey, authtypes.NewModuleAddress("gov").String(), ac)

	bz, err := ac.StringToBytes(authtypes.NewModuleAddress("gov").String())
	require.NoError(t, err)

	return &fixture{
		ctx:      testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test")).Ctx.WithBlockHeader(cmproto.Header{}),
		keeper:   k,
		mockAddr: bz,
		mockPerms: types.Permissions{
			Level: 3,
		},
		mockMsgURL: "mock_url",
	}
}

func TestGetAuthority(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	authority := f.keeper.GetAuthority()
	require.True(t, bytes.Equal(f.mockAddr, authority))
}

func TestGetAndSetPermissions(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	// Set the permissions for the mock address.

	err := f.keeper.SetPermissions(f.ctx, f.mockAddr, &f.mockPerms)
	require.NoError(t, err)

	// Retrieve the permissions for the mock address.
	perms, err := f.keeper.GetPermissions(f.ctx, f.mockAddr)
	require.NoError(t, err)

	//// Assert that the retrieved permissions match the expected value.
	require.Equal(t, &f.mockPerms, perms)
}

func TestIteratePermissions(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	// Define a set of mock permissions
	mockPerms := []types.Permissions{
		{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{"url1", "url2"}},
		{Level: types.Permissions_LEVEL_ALL_MSGS},
		{Level: types.Permissions_LEVEL_NONE_UNSPECIFIED},
	}

	// Set the permissions for a set of mock addresses
	mockAddrs := [][]byte{
		[]byte("mock_address_1"),
		[]byte("mock_address_2"),
		[]byte("mock_address_3"),
	}
	for i, addr := range mockAddrs {
		f.keeper.SetPermissions(f.ctx, addr, &mockPerms[i])
	}

	// Define a variable to store the returned permissions
	var returnedPerms []types.Permissions

	// Iterate through the permissions and append them to the returnedPerms slice
	f.keeper.IteratePermissions(f.ctx, func(address []byte, perms types.Permissions) (stop bool) {
		returnedPerms = append(returnedPerms, perms)
		return false
	})

	// Assert that the returned permissions match the set mock permissions
	require.Equal(t, mockPerms, returnedPerms)
}

func TestIterateDisabledList(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// Set the permissions for a set of mock addresses
	mockMsgs := []string{
		"mockUrl1",
		"mockUrl2",
		"mockUrl3",
	}

	for _, url := range mockMsgs {
		f.keeper.DisableMsg(f.ctx, url)
	}

	// Define a variable to store the returned disabled URLs
	var returnedDisabled []string

	f.keeper.IterateDisableLists(f.ctx, func(url string) bool {
		returnedDisabled = append(returnedDisabled, url)
		return false
	})

	// Assert that the returned disabled URLs match the set mock disabled URLs
	require.Equal(t, mockMsgs[0], returnedDisabled[0])
	require.Equal(t, mockMsgs[1], returnedDisabled[1])
	require.Equal(t, mockMsgs[2], returnedDisabled[2])

	// re-enable mockMsgs[0]
	f.keeper.EnableMsg(f.ctx, mockMsgs[0])

	returnedDisabled = []string{}
	f.keeper.IterateDisableLists(f.ctx, func(url string) bool {
		returnedDisabled = append(returnedDisabled, url)
		return false
	})

	// Assert that the returned disabled URLs match the set mock disabled URLs
	require.Len(t, returnedDisabled, 2)
	require.Equal(t, mockMsgs[1], returnedDisabled[0])
	require.Equal(t, mockMsgs[2], returnedDisabled[1])
}
