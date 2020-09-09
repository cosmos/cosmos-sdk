package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	holder     = "holder"
	multiPerm  = "multiple permissions account"
	randomPerm = "random permission"
)

var (
	multiPermAcc  = types.NewEmptyModuleAccount(multiPerm, types.Burner, types.Minter, types.Staking)
	randomPermAcc = types.NewEmptyModuleAccount(randomPerm, "random")
)

type KeeperTestSuite struct {
	suite.Suite

	app *simapp.SimApp
	ctx sdk.Context

	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app, suite.ctx = createTestApp(true)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.AccountKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func TestAccountMapperGetSet(t *testing.T) {
	app, ctx := createTestApp(true)
	addr := sdk.AccAddress([]byte("some---------address"))

	// no account before its created
	acc := app.AccountKeeper.GetAccount(ctx, addr)
	require.Nil(t, acc)

	// create account and check default values
	acc = app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	require.NotNil(t, acc)
	require.Equal(t, addr, acc.GetAddress())
	require.EqualValues(t, nil, acc.GetPubKey())
	require.EqualValues(t, 0, acc.GetSequence())

	// NewAccount doesn't call Set, so it's still nil
	require.Nil(t, app.AccountKeeper.GetAccount(ctx, addr))

	// set some values on the account and save it
	newSequence := uint64(20)
	err := acc.SetSequence(newSequence)
	require.NoError(t, err)
	app.AccountKeeper.SetAccount(ctx, acc)

	// check the new values
	acc = app.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	require.Equal(t, newSequence, acc.GetSequence())
}

func TestAccountMapperRemoveAccount(t *testing.T) {
	app, ctx := createTestApp(true)
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))

	// create accounts
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	accSeq1 := uint64(20)
	accSeq2 := uint64(40)

	err := acc1.SetSequence(accSeq1)
	require.NoError(t, err)
	err = acc2.SetSequence(accSeq2)
	require.NoError(t, err)
	app.AccountKeeper.SetAccount(ctx, acc1)
	app.AccountKeeper.SetAccount(ctx, acc2)

	acc1 = app.AccountKeeper.GetAccount(ctx, addr1)
	require.NotNil(t, acc1)
	require.Equal(t, accSeq1, acc1.GetSequence())

	// remove one account
	app.AccountKeeper.RemoveAccount(ctx, acc1)
	acc1 = app.AccountKeeper.GetAccount(ctx, addr1)
	require.Nil(t, acc1)

	acc2 = app.AccountKeeper.GetAccount(ctx, addr2)
	require.NotNil(t, acc2)
	require.Equal(t, accSeq2, acc2.GetSequence())
}

func TestGetSetParams(t *testing.T) {
	app, ctx := createTestApp(true)
	params := types.DefaultParams()

	app.AccountKeeper.SetParams(ctx, params)

	actualParams := app.AccountKeeper.GetParams(ctx)
	require.Equal(t, params, actualParams)
}

func TestSupply_ValidatePermissions(t *testing.T) {
	app, _ := createTestApp(true)

	// add module accounts to supply keeper
	maccPerms := simapp.GetMaccPerms()
	maccPerms[holder] = nil
	maccPerms[types.Burner] = []string{types.Burner}
	maccPerms[types.Minter] = []string{types.Minter}
	maccPerms[multiPerm] = []string{types.Burner, types.Minter, types.Staking}
	maccPerms[randomPerm] = []string{"random"}

	cdc, _ := simapp.MakeCodecs()
	keeper := keeper.NewAccountKeeper(
		cdc, app.GetKey(types.StoreKey), app.GetSubspace(types.ModuleName),
		types.ProtoBaseAccount, maccPerms,
	)

	err := keeper.ValidatePermissions(multiPermAcc)
	require.NoError(t, err)

	err = keeper.ValidatePermissions(randomPermAcc)
	require.NoError(t, err)

	// unregistered permissions
	otherAcc := types.NewEmptyModuleAccount("other", "other")
	err = app.AccountKeeper.ValidatePermissions(otherAcc)
	require.Error(t, err)
}
