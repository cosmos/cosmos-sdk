package keeper_test

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/circuit"
	"github.com/cosmos/cosmos-sdk/x/circuit/keeper"
	"github.com/cosmos/cosmos-sdk/x/circuit/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// Define a test suite for the circuit breaker keeper.
type KeeperTestSuite struct {
	suite.Suite
	cdc          codec.Codec
	ctx          sdk.Context
	keeper       keeper.Keeper
	mockAddr     sdk.AccAddress
	mockPerms    types.Permissions
	mockMsgURL   string
	mockCtx      sdk.Context
	mockStoreKey storetypes.StoreKey
}

func (suite *KeeperTestSuite) SetupTest() {
	// Define a mock store key.
	mockStoreKey := sdk.NewKVStoreKey("test")

	// Define a mock authority address.
	mockAddr := sdk.AccAddress("mock_address")

	// Define a mock set of permissions.
	mockPerms := types.Permissions{
		Level: types.Permissions_LEVEL_SUPER_ADMIN,
	}
	// Define a mock context.
	mockCtx := testutil.DefaultContextWithDB(suite.T(), mockStoreKey, sdk.NewTransientStoreKey("transient_test"))

	ctx := mockCtx.Ctx.WithBlockHeader(tmproto.Header{})

	encCfg := moduletestutil.MakeTestEncodingConfig(circuit.AppModuleBasic{})
	//Define codec

	// Define a mock message URL.
	mockMsgURL := "mock_url"

	// Create a new keeper instance.
	keeper := keeper.NewKeeper(encCfg.Codec, mockStoreKey, mockAddr.String())

	// Set the test suite variables.
	suite.ctx = ctx
	suite.cdc = encCfg.Codec
	suite.keeper = keeper
	suite.mockAddr = mockAddr
	suite.mockPerms = mockPerms
	suite.mockMsgURL = mockMsgURL
	suite.mockStoreKey = mockStoreKey
}

func (suite *KeeperTestSuite) TestGetAuthority() {
	require.Equal(suite.T(), suite.mockAddr.String(), suite.keeper.GetAuthority())
}

func (suite *KeeperTestSuite) TestGetPermissions() {
	// Define a mock set of permissions to set.
	mockPerms := types.Permissions{
		Level: types.Permissions_LEVEL_SUPER_ADMIN,
	}

	// Set the permissions for the mock address.
	err := suite.keeper.SetPermissions(suite.ctx, suite.mockAddr, &mockPerms)
	require.NoError(suite.T(), err)

	// Retrieve the set permissions from the store and check that they match the mock permissions.
	perms, err := suite.keeper.GetPermissions(suite.ctx, suite.mockAddr.String())
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), mockPerms, perms)
}

func (suite *KeeperTestSuite) TestSetPermissions() {
	err := suite.keeper.SetPermissions(suite.mockCtx, suite.mockAddr, &suite.mockPerms)
	require.NoError(suite.T(), err)
	perms, _ := suite.keeper.GetPermissions(suite.mockCtx, suite.mockAddr.String())
	require.Equal(suite.T(), suite.mockPerms, perms)
}

func (suite *KeeperTestSuite) TestIteratePermissions() {
	// Define a set of mock permissions
	mockPerms := []types.Permissions{
		{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{"url1", "url2"}},
		{Level: types.Permissions_LEVEL_ALL_MSGS},
		{Level: types.Permissions_LEVEL_NONE_UNSPECIFIED},
	}

	// Set the permissions for a set of mock addresses
	mockAddrs := []sdk.AccAddress{
		sdk.AccAddress("mock_address_1"),
		sdk.AccAddress("mock_address_2"),
		sdk.AccAddress("mock_address_3"),
	}
	for i, addr := range mockAddrs {
		suite.keeper.SetPermissions(suite.mockCtx, addr, &mockPerms[i])
	}

	// Define a variable to store the returned permissions
	var returnedPerms []types.Permissions

	// Iterate through the permissions and append them to the returnedPerms slice
	suite.keeper.IteratePermissions(suite.mockCtx, func(address []byte, perms types.Permissions) (stop bool) {
		returnedPerms = append(returnedPerms, perms)
		return false
	})

	// Assert that the returned permissions match the set mock permissions
	require.Equal(suite.T(), mockPerms, returnedPerms)
}
