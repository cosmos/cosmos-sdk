package keeper_test

import (
	"bytes"
	context "context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/circuit"
	"cosmossdk.io/x/circuit/keeper"
	"cosmossdk.io/x/circuit/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

var addresses = []string{
	"cosmos1zglwfu6xjzvzagqcmvzewyzjp9xwqw5qwrr8n9",
	"cosmos1p8s0p6gqc6c9gt77lgr2qqujz49huhu6a80smx",
	"cosmos1qasf9ehx8m7cnat39ndc74rx3fg7z66u8lw0fd",
	"cosmos1uxrdj5zfuudhypsmmjxnj4gpu432ycht06a05a",
	"cosmos1wn7k8a7fwpmrwnm94ndj0germfnxnhl6hs8spj",
}

type fixture struct {
	ctx        context.Context
	keeper     keeper.Keeper
	mockAddr   []byte
	mockPerms  types.Permissions
	mockMsgURL string
	ac         address.Codec
}

func initFixture(t *testing.T) *fixture {
	t.Helper()
	encCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, circuit.AppModule{})
	ac := addresscodec.NewBech32Codec("cosmos")
	mockStoreKey := storetypes.NewKVStoreKey("test")

	env := runtime.NewEnvironment(runtime.NewKVStoreService(mockStoreKey), log.NewNopLogger())
	k := keeper.NewKeeper(env, encCfg.Codec, authtypes.NewModuleAddress("gov").String(), ac)

	bz, err := ac.StringToBytes(authtypes.NewModuleAddress("gov").String())
	require.NoError(t, err)

	return &fixture{
		ctx:      testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test")).Ctx,
		keeper:   k,
		mockAddr: bz,
		mockPerms: types.Permissions{
			Level:         3,
			LimitTypeUrls: []string{"test"},
		},
		mockMsgURL: "mock_url",
		ac:         ac,
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

	err := f.keeper.Permissions.Set(f.ctx, f.mockAddr, f.mockPerms)
	require.NoError(t, err)

	// Retrieve the permissions for the mock address.
	perms, err := f.keeper.Permissions.Get(f.ctx, f.mockAddr)
	require.NoError(t, err)

	//// Assert that the retrieved permissions match the expected value.
	require.Equal(t, f.mockPerms, perms)
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
		err := f.keeper.Permissions.Set(f.ctx, addr, mockPerms[i])
		require.NoError(t, err)
	}

	// Define a variable to store the returned permissions
	var returnedPerms []types.Permissions

	// Iterate through the permissions and append them to the returnedPerms slice
	err := f.keeper.Permissions.Walk(f.ctx, nil, func(address []byte, perms types.Permissions) (stop bool, err error) {
		returnedPerms = append(returnedPerms, perms)
		return false, nil
	})
	require.NoError(t, err)

	// Assert that the returned permissions match the set mock permissions
	require.Equal(t, mockPerms, returnedPerms)
}

func TestIterateDisabledList(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	mockMsgs := []string{
		"mockUrl1",
		"mockUrl2",
		"mockUrl3",
	}

	for _, url := range mockMsgs {
		require.NoError(t, f.keeper.DisableList.Set(f.ctx, url))
	}

	// Define a variable to store the returned disabled URLs
	var returnedDisabled []string

	err := f.keeper.DisableList.Walk(f.ctx, nil, func(msgUrl string) (bool, error) {
		returnedDisabled = append(returnedDisabled, msgUrl)
		return false, nil
	})
	require.NoError(t, err)

	// Assert that the returned disabled URLs match the set mock disabled URLs
	require.Equal(t, mockMsgs[0], returnedDisabled[0])
	require.Equal(t, mockMsgs[1], returnedDisabled[1])
	require.Equal(t, mockMsgs[2], returnedDisabled[2])

	// re-enable mockMsgs[0]
	require.NoError(t, f.keeper.DisableList.Remove(f.ctx, mockMsgs[0]))
	returnedDisabled = []string{}

	err = f.keeper.DisableList.Walk(f.ctx, nil, func(msgUrl string) (bool, error) {
		returnedDisabled = append(returnedDisabled, msgUrl)
		return false, nil
	})
	require.NoError(t, err)

	require.Len(t, returnedDisabled, 2)
	require.Equal(t, mockMsgs[1], returnedDisabled[0])
	require.Equal(t, mockMsgs[2], returnedDisabled[1])
}
