package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"pgregory.net/rapid"
)

type DeterministicTestSuite struct {
	suite.Suite

	ctx sdk.Context

	queryClient   types.QueryClient
	accountKeeper keeper.AccountKeeper
	encCfg        moduletestutil.TestEncodingConfig
}

func TestDeterministicTestSuite(t *testing.T) {
	suite.Run(t, new(DeterministicTestSuite))
}

func (suite *DeterministicTestSuite) SetupTest() {
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

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, suite.accountKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func (suite *DeterministicTestSuite) TestGRPCQueryAccounts() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		addr := testdata.AddrTestDeterministic(t)
		suite.accountKeeper.SetAccount(suite.ctx,
			suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr))

		acc, err := suite.queryClient.Account(suite.ctx, &types.QueryAccountRequest{Address: addr.String()})
		suite.Require().NoError(err)
		suite.Require().NotNil(acc)

		var prevRes types.AccountI
		err = suite.encCfg.InterfaceRegistry.UnpackAny(acc.Account, &prevRes)
		suite.Require().NoError(err)

		for i := 0; i < 1000; i++ {
			acc, err := suite.queryClient.Account(suite.ctx, &types.QueryAccountRequest{Address: addr.String()})
			suite.Require().NoError(err)
			suite.Require().NotNil(acc)
			var account types.AccountI

			err = suite.encCfg.InterfaceRegistry.UnpackAny(acc.Account, &account)
			suite.Require().NoError(err)
			suite.Require().Equal(account.GetAddress(), addr)

			// check with previous response too.
			suite.Require().Equal(account.GetAddress(), prevRes.GetAddress())
			suite.Require().Equal(account.GetPubKey(), prevRes.GetPubKey())
			suite.Require().Equal(account.GetSequence(), prevRes.GetSequence())
			suite.Require().Equal(account.GetAccountNumber(), prevRes.GetAccountNumber())
			prevRes = account
		}
	})

	priv := secp256k1.GenPrivKey()
	pub := priv.PubKey()
	addr1 := sdk.AccAddress(priv.PubKey().Address())
	// randAccNumber :=
	accNum := uint64(rand.Intn(100) + 10000) // range 10000 to 10100
	seq := uint64(0)

	acc1 := types.NewBaseAccount(addr1, pub, accNum, seq)
	suite.accountKeeper.SetAccount(suite.ctx, acc1)

	rapid.Check(suite.T(), func(t *rapid.T) {
		acc, err := suite.queryClient.Account(suite.ctx, &types.QueryAccountRequest{Address: addr1.String()})
		suite.Require().NoError(err)
		suite.Require().NotNil(acc)
		var account types.AccountI

		err = suite.encCfg.InterfaceRegistry.UnpackAny(acc.Account, &account)
		suite.Require().NoError(err)

		suite.Require().Equal(account.GetAddress(), addr1)
		suite.Require().Equal(account.GetPubKey(), pub)
		suite.Require().Equal(account.GetAccountNumber(), accNum)
		suite.Require().Equal(account.GetSequence(), seq)
	})
}
