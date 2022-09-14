package keeper_test

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/btcutil/base58"
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
	"github.com/tendermint/tendermint/crypto"
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

func (suite *DeterministicTestSuite) runIterations(addr sdk.AccAddress, prevRes types.AccountI) {
	for i := 0; i < 1000; i++ {
		acc, err := suite.queryClient.Account(suite.ctx, &types.QueryAccountRequest{Address: addr.String()})
		suite.Require().NoError(err)
		suite.Require().NotNil(acc)
		var account types.AccountI

		err = suite.encCfg.InterfaceRegistry.UnpackAny(acc.Account, &account)
		suite.Require().NoError(err)
		suite.Require().Equal(account.GetAddress(), addr)

		if prevRes != nil {
			suite.Require().Equal(account, prevRes)
		}

		prevRes = account
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryAccounts() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		addr := testdata.AddressGenerator(t).Draw(t, "address")
		acc1 := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
		suite.accountKeeper.SetAccount(suite.ctx, acc1)

		suite.runIterations(addr, acc1)
	})

	addrBbz, _, err := base58.CheckDecode("1CKZ9Nx4zgds8tU7nJHotKSDr4a9bYJCa3")
	suite.Require().NoError(err)
	addr1 := sdk.AccAddress(crypto.Address(addrBbz))

	pub, err := hex.DecodeString("02950e1cdfcb133d6024109fd489f734eeb4502418e538c28481f22bce276f248c")
	suite.Require().NoError(err)

	acc1 := types.NewBaseAccount(addr1, &secp256k1.PubKey{Key: pub}, uint64(10087), uint64(0))

	suite.accountKeeper.SetAccount(suite.ctx, acc1)
	suite.runIterations(addr1, acc1)
}
