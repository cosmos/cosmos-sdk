package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"pgregory.net/rapid"
)

type RapidTestSuite struct {
	suite.Suite

	ctx sdk.Context

	queryClient   types.QueryClient
	accountKeeper keeper.AccountKeeper
	msgServer     types.MsgServer
	encCfg        moduletestutil.TestEncodingConfig
}

func TestRapidTestSuite(t *testing.T) {
	suite.Run(t, new(RapidTestSuite))
}

func (suite *RapidTestSuite) SetupTest() {
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

func (suite *RapidTestSuite) TestGRPCQueryAccounts() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		pkBz := rapid.SliceOfN(rapid.Byte(), 20, 20).Draw(t, "hex")
		addr := sdk.AccAddress(pkBz)
		suite.accountKeeper.SetAccount(suite.ctx,
			suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr))

		for i := 0; i < 1000; i++ {
			acc, err := suite.queryClient.Account(suite.ctx, &types.QueryAccountRequest{Address: addr.String()})
			suite.Require().NoError(err)
			suite.Require().NotNil(acc)
			var account types.AccountI

			err = suite.encCfg.InterfaceRegistry.UnpackAny(acc.Account, &account)
			suite.Require().NoError(err)
			suite.Require().Equal(account.GetAddress(), addr)
		}
	})
}
