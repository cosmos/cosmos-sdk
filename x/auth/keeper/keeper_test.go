package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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

	ctx sdk.Context

	queryClient   types.QueryClient
	accountKeeper keeper.AccountKeeper
	msgServer     types.MsgServer
	encCfg        moduletestutil.TestEncodingConfig
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.encCfg = moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})

	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{})

	maccPerms := map[string][]string{
		"fee_collector":          nil,
		"mint":                   {"minter"},
		"bonded_tokens_pool":     {"burner", "staking"},
		"not_bonded_tokens_pool": {"burner", "staking"},
		multiPerm:                {"burner", "minter", "staking"},
		randomPerm:               {"random"},
	}

	suite.accountKeeper = keeper.NewAccountKeeper(
		suite.encCfg.Codec,
		key,
		types.ProtoBaseAccount,
		maccPerms,
		"cosmos",
		types.NewModuleAddress("gov").String(),
	)

	suite.msgServer = keeper.NewMsgServerImpl(suite.accountKeeper)
	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, suite.accountKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestAccountMapperGetSet() {
	ctx := suite.ctx
	addr := sdk.AccAddress([]byte("some---------address"))

	// no account before its created
	acc := suite.accountKeeper.GetAccount(ctx, addr)
	suite.Require().Nil(acc)

	// create account and check default values
	acc = suite.accountKeeper.NewAccountWithAddress(ctx, addr)
	suite.Require().NotNil(acc)
	suite.Require().Equal(addr, acc.GetAddress())
	suite.Require().EqualValues(nil, acc.GetPubKey())
	suite.Require().EqualValues(0, acc.GetSequence())

	// NewAccount doesn't call Set, so it's still nil
	suite.Require().Nil(suite.accountKeeper.GetAccount(ctx, addr))

	// set some values on the account and save it
	newSequence := uint64(20)
	err := acc.SetSequence(newSequence)
	suite.Require().NoError(err)
	suite.accountKeeper.SetAccount(ctx, acc)

	// check the new values
	acc = suite.accountKeeper.GetAccount(ctx, addr)
	suite.Require().NotNil(acc)
	suite.Require().Equal(newSequence, acc.GetSequence())
}

func (suite *KeeperTestSuite) TestAccountMapperRemoveAccount() {
	ctx := suite.ctx
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))

	// create accounts
	acc1 := suite.accountKeeper.NewAccountWithAddress(ctx, addr1)
	acc2 := suite.accountKeeper.NewAccountWithAddress(ctx, addr2)

	accSeq1 := uint64(20)
	accSeq2 := uint64(40)

	err := acc1.SetSequence(accSeq1)
	suite.Require().NoError(err)
	err = acc2.SetSequence(accSeq2)
	suite.Require().NoError(err)
	suite.accountKeeper.SetAccount(ctx, acc1)
	suite.accountKeeper.SetAccount(ctx, acc2)

	acc1 = suite.accountKeeper.GetAccount(ctx, addr1)
	suite.Require().NotNil(acc1)
	suite.Require().Equal(accSeq1, acc1.GetSequence())

	// remove one account
	suite.accountKeeper.RemoveAccount(ctx, acc1)
	acc1 = suite.accountKeeper.GetAccount(ctx, addr1)
	suite.Require().Nil(acc1)

	acc2 = suite.accountKeeper.GetAccount(ctx, addr2)
	suite.Require().NotNil(acc2)
	suite.Require().Equal(accSeq2, acc2.GetSequence())
}

func (suite *KeeperTestSuite) TestGetSetParams() {
	ctx := suite.ctx
	params := types.DefaultParams()

	err := suite.accountKeeper.SetParams(ctx, params)
	suite.Require().NoError(err)

	actualParams := suite.accountKeeper.GetParams(ctx)
	suite.Require().Equal(params, actualParams)
}

func (suite *KeeperTestSuite) TestSupply_ValidatePermissions() {
	err := suite.accountKeeper.ValidatePermissions(multiPermAcc)
	suite.Require().NoError(err)

	err = suite.accountKeeper.ValidatePermissions(randomPermAcc)
	suite.Require().NoError(err)

	// unregistered permissions
	otherAcc := types.NewEmptyModuleAccount("other", "other")
	err = suite.accountKeeper.ValidatePermissions(otherAcc)
	suite.Require().Error(err)
}
