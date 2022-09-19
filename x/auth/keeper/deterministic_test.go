package keeper_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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

type DeterministicTestSuite struct {
	suite.Suite

	ctx           sdk.Context
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

func (suite *DeterministicTestSuite) runAccountIterations(addr sdk.AccAddress, prevRes types.AccountI) {
	for i := 0; i < 1000; i++ {
		acc, err := suite.queryClient.Account(suite.ctx, &types.QueryAccountRequest{Address: addr.String()})
		suite.Require().NoError(err)
		suite.Require().NotNil(acc)

		var account types.AccountI
		err = suite.encCfg.InterfaceRegistry.UnpackAny(acc.Account, &account)
		suite.Require().NoError(err)
		suite.Require().Equal(account.GetAddress(), addr)

		if prevRes != nil {
			any, err := codectypes.NewAnyWithValue(prevRes)
			suite.Require().NoError(err)

			suite.Require().Equal(acc.Account, any)
			suite.Require().Equal(account, prevRes)
		}

		prevRes = account
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryAccount() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		pub := pubkeyGenerator(t).Draw(t, "pubkey")
		addr := sdk.AccAddress(pub.Address())
		accNum := rapid.Uint64().Draw(t, "account-number")
		seq := rapid.Uint64().Draw(t, "sequence")

		fmt.Println(pub.String())
		fmt.Println(addr.String())
		suite.Require().NotNil(nil)

		acc1 := types.NewBaseAccount(addr, &pub, accNum, seq)
		suite.accountKeeper.SetAccount(suite.ctx, acc1)

		suite.runAccountIterations(addr, acc1)
	})

	// Regression test
	addr1 := sdk.MustAccAddressFromBech32("cosmos1j364pjm8jkxxmujj0vp2xjg0y7w8tyveuamfm6")
	pub, err := hex.DecodeString("01090C02812F010C25200ED40E004105160196E801F70005070EA21603FF06001E")
	suite.Require().NoError(err)

	accNum := uint64(10087)
	seq := uint64(0)

	acc1 := types.NewBaseAccount(addr1, &secp256k1.PubKey{Key: pub}, accNum, seq)

	suite.accountKeeper.SetAccount(suite.ctx, acc1)
	suite.runAccountIterations(addr1, acc1)
}

// pubkeyGenerator creates and returns a random pubkey generator using rapid.
func pubkeyGenerator(t *rapid.T) *rapid.Generator[secp256k1.PubKey] {
	return rapid.Custom(func(t *rapid.T) secp256k1.PubKey {
		pkBz := rapid.SliceOfN(rapid.Byte(), 33, 33).Draw(t, "hex")
		return secp256k1.PubKey{Key: pkBz}
	})
}

func (suite *DeterministicTestSuite) runAccountsIterations(prevRes []types.AccountI) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.Accounts(suite.ctx, &types.QueryAccountsRequest{})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().NotNil(res.Accounts)
		suite.Require().Len(res.Accounts, len(prevRes))

		uppackedAccs := make([]types.AccountI, len(res.Accounts))
		for i := 0; i < len(res.Accounts); i++ {
			var account types.AccountI
			err = suite.encCfg.InterfaceRegistry.UnpackAny(res.Accounts[i], &account)
			suite.Require().NoError(err)

			uppackedAccs[i] = account
		}

		sort.Slice(uppackedAccs, func(i2, j int) bool {
			return uppackedAccs[i2].GetAccountNumber() < uppackedAccs[j].GetAccountNumber()
		})

		sort.Slice(prevRes, func(i2, j int) bool {
			return prevRes[i2].GetAccountNumber() < prevRes[j].GetAccountNumber()
		})

		if prevRes != nil {
			suite.Require().Equal(uppackedAccs, prevRes)
		}

		prevRes = uppackedAccs
	}

	for i := 0; i < len(prevRes); i++ {
		suite.accountKeeper.RemoveAccount(suite.ctx, prevRes[i])
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryAccounts() {

	rapid.Check(suite.T(), func(t *rapid.T) {
		numAccs := rand.Intn(10) + 1
		accs := make([]types.AccountI, numAccs)

		for i := 0; i < numAccs; i++ {
			pub := pubkeyGenerator(t).Draw(t, "pubkey")
			addr := sdk.AccAddress(pub.Address())
			accNum := uint64(i*10 + rand.Intn(10)) // to avoid collisions
			seq := rapid.Uint64().Draw(t, "sequence")

			acc1 := types.NewBaseAccount(addr, &pub, accNum, seq)
			suite.accountKeeper.SetAccount(suite.ctx, acc1)
			accs[i] = acc1
		}

		suite.runAccountsIterations(accs)
		for i := 0; i < numAccs; i++ {
			suite.accountKeeper.RemoveAccount(suite.ctx, accs[i])
		}
	})

	// Regression test
	addr1 := sdk.MustAccAddressFromBech32("cosmos1892yr6fzlj7ud0kfkah2ctrav3a4p4n060ze8f")
	pub1, err := hex.DecodeString("D1002E1B019000010BB7034500E71F011F1CA90D5B000E134BFB0F3603030D0303")
	suite.Require().NoError(err)
	accNum1 := uint64(107)
	seq1 := uint64(0)

	addr2 := sdk.MustAccAddressFromBech32("cosmos1j364pjm8jkxxmujj0vp2xjg0y7w8tyveuamfm6")
	pub2, err := hex.DecodeString("01090C02812F010C25200ED40E004105160196E801F70005070EA21603FF06001E")
	suite.Require().NoError(err)

	accNum2 := uint64(100)
	seq2 := uint64(10)

	acc1 := types.NewBaseAccount(addr1, &secp256k1.PubKey{Key: pub1}, accNum1, seq1)
	acc2 := types.NewBaseAccount(addr2, &secp256k1.PubKey{Key: pub2}, accNum2, seq2)

	suite.accountKeeper.SetAccount(suite.ctx, acc1)
	suite.accountKeeper.SetAccount(suite.ctx, acc2)

	suite.runAccountsIterations([]types.AccountI{acc1, acc2})
}
