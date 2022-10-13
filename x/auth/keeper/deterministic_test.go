package keeper_test

import (
	"context"
	"encoding/hex"
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"google.golang.org/grpc"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

type DeterministicTestSuite struct {
	suite.Suite

	key           *storetypes.KVStoreKey
	ctx           sdk.Context
	queryClient   types.QueryClient
	accountKeeper keeper.AccountKeeper
	encCfg        moduletestutil.TestEncodingConfig
	maccPerms     map[string][]string
}

var (
	iterCount = 1000
	addr      = sdk.MustAccAddressFromBech32("cosmos1j364pjm8jkxxmujj0vp2xjg0y7w8tyveuamfm6")
	pub, _    = hex.DecodeString("01090C02812F010C25200ED40E004105160196E801F70005070EA21603FF06001E")
)

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

	suite.key = key
	suite.maccPerms = maccPerms
}

type QueryRequest interface {
	*types.QueryAccountRequest | *types.QueryAccountsRequest
}

type QueryResponse interface {
	*types.QueryAccountResponse | *types.QueryAccountsResponse
}

func queryReq[request QueryRequest, response QueryResponse](
	suite *DeterministicTestSuite,
	req request, prevRes response,
	grpcFn func(context.Context, request, ...grpc.CallOption) (response, error),
	gasConsumed uint64,
) {

	for i := 0; i < iterCount; i++ {
		before := suite.ctx.GasMeter().GasConsumed()
		res, err := grpcFn(suite.ctx, req)
		suite.Require().Equal(suite.ctx.GasMeter().GasConsumed()-before, gasConsumed)
		suite.Require().NoError(err)
		suite.Require().Equal(res, prevRes)
	}
}

// createAndSetAccount creates a random account and sets to the keeper store.
func (suite *DeterministicTestSuite) createAndSetAccounts(t *rapid.T, count int) []types.AccountI {
	accs := make([]types.AccountI, 0, count)

	// We need all generated account-numbers unique
	accNums := rapid.SliceOfNDistinct(rapid.Uint64(), count, count, func(i uint64) uint64 { return i }).Draw(t, "acc-numss")

	for i := 0; i < count; i++ {
		pub := pubkeyGenerator(t).Draw(t, "pubkey")
		addr := sdk.AccAddress(pub.Address())
		accNum := accNums[i]
		seq := rapid.Uint64().Draw(t, "sequence")

		acc1 := types.NewBaseAccount(addr, &pub, accNum, seq)
		suite.accountKeeper.SetAccount(suite.ctx, acc1)
		accs = append(accs, acc1)
	}

	return accs
}

func (suite *DeterministicTestSuite) TestGRPCQueryAccount() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		accs := suite.createAndSetAccounts(t, 1)

		req := &types.QueryAccountRequest{Address: accs[0].GetAddress().String()}
		before := suite.ctx.GasMeter().GasConsumed()
		res, err := suite.queryClient.Account(suite.ctx, req)
		suite.Require().NoError(err)

		queryReq(suite, req, res, suite.queryClient.Account, suite.ctx.GasMeter().GasConsumed()-before)
	})

	// Regression tests
	accNum := uint64(10087)
	seq := uint64(98)

	acc1 := types.NewBaseAccount(addr, &secp256k1.PubKey{Key: pub}, accNum, seq)
	suite.accountKeeper.SetAccount(suite.ctx, acc1)

	req := &types.QueryAccountRequest{Address: acc1.GetAddress().String()}
	res, err := suite.queryClient.Account(suite.ctx, req)
	suite.Require().NoError(err)

	queryReq(suite, req, res, suite.queryClient.Account, 1543)
}

// pubkeyGenerator creates and returns a random pubkey generator using rapid.
func pubkeyGenerator(t *rapid.T) *rapid.Generator[secp256k1.PubKey] {
	return rapid.Custom(func(t *rapid.T) secp256k1.PubKey {
		pkBz := rapid.SliceOfN(rapid.Byte(), 33, 33).Draw(t, "hex")
		return secp256k1.PubKey{Key: pkBz}
	})
}

func (suite *DeterministicTestSuite) TestGRPCQueryAccounts() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		numAccs := rapid.IntRange(1, 10).Draw(t, "accounts")
		accs := suite.createAndSetAccounts(t, numAccs)

		req := &types.QueryAccountsRequest{
			Pagination: testdata.PaginationGenerator(t, uint64(numAccs)).Draw(t, "accounts"),
		}

		before := suite.ctx.GasMeter().GasConsumed()
		res, err := suite.queryClient.Accounts(suite.ctx, req)
		suite.Require().NoError(err)

		queryReq(suite, req, res, suite.queryClient.Accounts, suite.ctx.GasMeter().GasConsumed()-before)

		for i := 0; i < numAccs; i++ {
			suite.accountKeeper.RemoveAccount(suite.ctx, accs[i])
		}
	})

	// Regression test
	addr1, err := sdk.AccAddressFromBech32("cosmos1892yr6fzlj7ud0kfkah2ctrav3a4p4n060ze8f")
	suite.Require().NoError(err)
	pub1, err := hex.DecodeString("D1002E1B019000010BB7034500E71F011F1CA90D5B000E134BFB0F3603030D0303")
	suite.Require().NoError(err)
	accNum1 := uint64(107)
	seq1 := uint64(10001)

	accNum2 := uint64(100)
	seq2 := uint64(10)

	acc1 := types.NewBaseAccount(addr1, &secp256k1.PubKey{Key: pub1}, accNum1, seq1)
	acc2 := types.NewBaseAccount(addr, &secp256k1.PubKey{Key: pub}, accNum2, seq2)

	suite.accountKeeper.SetAccount(suite.ctx, acc1)
	suite.accountKeeper.SetAccount(suite.ctx, acc2)

	req := &types.QueryAccountsRequest{}
	res, err := suite.queryClient.Accounts(suite.ctx, req)
	suite.Require().NoError(err)

	queryReq(suite, req, res, suite.queryClient.Accounts, 1716)
}

func (suite *DeterministicTestSuite) runAccountAddressByIDIterations(id int64, prevRes string) {
	for i := 0; i < iterCount; i++ {
		res, err := suite.queryClient.AccountAddressByID(suite.ctx, &types.QueryAccountAddressByIDRequest{Id: id})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().Equal(res.AccountAddress, prevRes)
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryAccountAddressByID() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		pub := pubkeyGenerator(t).Draw(t, "pubkey")
		addr := sdk.AccAddress(pub.Address())

		accNum := rapid.Int64Min(0).Draw(t, "account-number")
		seq := rapid.Uint64().Draw(t, "sequence")

		acc1 := types.NewBaseAccount(addr, &pub, uint64(accNum), seq)
		suite.accountKeeper.SetAccount(suite.ctx, acc1)

		suite.runAccountAddressByIDIterations(accNum, addr.String())
	})

	// Regression test
	accNum := int64(10087)
	seq := uint64(0)

	acc1 := types.NewBaseAccount(addr, &secp256k1.PubKey{Key: pub}, uint64(accNum), seq)

	suite.accountKeeper.SetAccount(suite.ctx, acc1)
	suite.runAccountAddressByIDIterations(accNum, addr.String())
}

func (suite *DeterministicTestSuite) runParamsIterations(prevRes types.Params) {
	for i := 0; i < iterCount; i++ {
		res, err := suite.queryClient.Params(suite.ctx, &types.QueryParamsRequest{})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.Params, prevRes)
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryParameters() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		params := types.NewParams(
			rapid.Uint64Min(1).Draw(t, "max-memo-characters"),
			rapid.Uint64Min(1).Draw(t, "tx-sig-limit"),
			rapid.Uint64Min(1).Draw(t, "tx-size-cost-per-byte"),
			rapid.Uint64Min(1).Draw(t, "sig-verify-cost-ed25519"),
			rapid.Uint64Min(1).Draw(t, "sig-verify-cost-Secp256k1"),
		)

		err := suite.accountKeeper.SetParams(suite.ctx, params)
		suite.Require().NoError(err)

		suite.runParamsIterations(params)
	})

	// Regression test
	params := types.NewParams(15, 167, 100, 1, 21457)

	err := suite.accountKeeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	suite.runParamsIterations(params)
}

func (suite *DeterministicTestSuite) runAccountInfoIterations(addr sdk.AccAddress, prevRes *types.BaseAccount) {
	for i := 0; i < iterCount; i++ {
		res, err := suite.queryClient.AccountInfo(suite.ctx, &types.QueryAccountInfoRequest{Address: addr.String()})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().NotNil(res.Info)

		suite.Require().Equal(res.GetInfo().Address, prevRes.Address)
		suite.Require().True(res.GetInfo().PubKey.Equal(prevRes.PubKey))
		suite.Require().Equal(res.GetInfo().AccountNumber, prevRes.AccountNumber)
		suite.Require().Equal(res.GetInfo().Sequence, prevRes.Sequence)

	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryAccountInfo() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		accs := suite.createAndSetAccounts(t, 1)
		suite.Require().Len(accs, 1)

		acc, ok := accs[0].(*types.BaseAccount)
		suite.Require().True(ok)

		suite.runAccountInfoIterations(acc.GetAddress(), acc)
	})

	// Regression test
	accNum := uint64(10087)
	seq := uint64(10)

	acc1 := types.NewBaseAccount(addr, &secp256k1.PubKey{Key: pub}, accNum, seq)

	suite.accountKeeper.SetAccount(suite.ctx, acc1)
	suite.runAccountInfoIterations(addr, acc1)
}

func (suite *DeterministicTestSuite) createAndReturnQueryClient(ak keeper.AccountKeeper) types.QueryClient {
	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, ak)
	return types.NewQueryClient(queryHelper)
}

func (suite *DeterministicTestSuite) runBech32PrefixIterations(ak keeper.AccountKeeper, preRes string) {
	queryClient := suite.createAndReturnQueryClient(ak)

	for i := 0; i < iterCount; i++ {
		res, err := queryClient.Bech32Prefix(suite.ctx, &types.Bech32PrefixRequest{})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.Bech32Prefix, preRes)
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryBech32Prefix() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		prefix := rapid.StringMatching(`[a-zA-Z]+[1-9a-zA-Z]*`).Draw(t, "prefix")
		ak := keeper.NewAccountKeeper(
			suite.encCfg.Codec,
			suite.key,
			types.ProtoBaseAccount,
			nil,
			prefix,
			types.NewModuleAddress("gov").String(),
		)

		suite.runBech32PrefixIterations(ak, prefix)
	})

	suite.runBech32PrefixIterations(suite.accountKeeper, "cosmos")
}

func (suite *DeterministicTestSuite) runAddressBytesToStringIterations(addressBytes []byte, prevRes string) {
	for i := 0; i < iterCount; i++ {
		res, err := suite.queryClient.AddressBytesToString(suite.ctx, &types.AddressBytesToStringRequest{
			AddressBytes: addressBytes,
		})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.AddressString, prevRes)
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryAddressBytesToString() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		address := testdata.AddressGenerator(t).Draw(t, "address-bytes")
		suite.runAddressBytesToStringIterations(address.Bytes(), address.String())
	})

	suite.runAddressBytesToStringIterations(addr.Bytes(), addr.String())
}

func (suite *DeterministicTestSuite) runStringToAddressBytesIterations(addressString string, prevRes []byte) {
	for i := 0; i < iterCount; i++ {
		res, err := suite.queryClient.AddressStringToBytes(suite.ctx, &types.AddressStringToBytesRequest{
			AddressString: addressString,
		})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.AddressBytes, prevRes)
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryAddressStringToBytes() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		address := testdata.AddressGenerator(t).Draw(t, "address-string")
		suite.runStringToAddressBytesIterations(address.String(), address.Bytes())
	})

	suite.runStringToAddressBytesIterations(addr.String(), addr.Bytes())
}

func (suite *DeterministicTestSuite) setModuleAccounts(
	ctx sdk.Context, ak keeper.AccountKeeper, maccs []string) []types.AccountI {
	sort.Strings(maccs)
	moduleAccounts := make([]types.AccountI, 0, len(maccs))
	for _, m := range maccs {
		acc, _ := ak.GetModuleAccountAndPermissions(ctx, m)
		acc1, ok := acc.(types.AccountI)
		suite.Require().True(ok)
		moduleAccounts = append(moduleAccounts, acc1)
	}

	return moduleAccounts
}

func (suite *DeterministicTestSuite) runModuleAccountsIterations(ak keeper.AccountKeeper, prevRes []types.AccountI) {
	queryClient := suite.createAndReturnQueryClient(ak)
	for i := 0; i < iterCount; i++ {
		res, err := queryClient.ModuleAccounts(suite.ctx, &types.QueryModuleAccountsRequest{})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().NotNil(res.Accounts)
		suite.Require().Len(res.Accounts, len(prevRes))

		unpackedAccs := make([]types.AccountI, len(res.Accounts))
		for i := 0; i < len(res.Accounts); i++ {
			var account types.AccountI
			err = suite.encCfg.InterfaceRegistry.UnpackAny(res.Accounts[i], &account)
			suite.Require().NoError(err)

			unpackedAccs[i] = account
		}

		if prevRes != nil {
			for i := 0; i < len(prevRes); i++ {
				suite.Require().Equal(unpackedAccs[i].GetAddress(), prevRes[i].GetAddress())
				suite.Require().Equal(unpackedAccs[i].GetAccountNumber(), prevRes[i].GetAccountNumber())
				suite.Require().Equal(unpackedAccs[i].GetSequence(), prevRes[i].GetSequence())
			}
		}
	}

}

func (suite *DeterministicTestSuite) TestGRPCQueryModuleAccounts() {
	permissions := []string{"burner", "minter", "staking", "random"}

	rapid.Check(suite.T(), func(t *rapid.T) {
		maccsCount := rapid.IntRange(1, 10).Draw(t, "accounts")
		maccs := make([]string, maccsCount)

		for i := 0; i < maccsCount; i++ {
			maccs[i] = rapid.StringMatching(`[a-z]{5,}`).Draw(t, "module-name")
		}

		maccPerms := make(map[string][]string)
		for i := 0; i < maccsCount; i++ {
			mPerms := make([]string, 0, 4)
			for _, permission := range permissions {
				if rapid.Bool().Draw(t, "permissions") {
					mPerms = append(mPerms, permission)
				}
			}

			if len(mPerms) == 0 {
				num := rapid.IntRange(0, 3).Draw(t, "num")
				mPerms = append(mPerms, permissions[num])
			}

			maccPerms[maccs[i]] = mPerms
		}

		ak := keeper.NewAccountKeeper(
			suite.encCfg.Codec,
			suite.key,
			types.ProtoBaseAccount,
			maccPerms,
			"cosmos",
			types.NewModuleAddress("gov").String(),
		)

		storedMaccs := suite.setModuleAccounts(suite.ctx, ak, maccs)

		suite.runModuleAccountsIterations(ak, storedMaccs)
		for i := 0; i < len(storedMaccs); i++ {
			suite.accountKeeper.RemoveAccount(suite.ctx, storedMaccs[i])
		}
	})

	maccs := make([]string, 0, len(suite.maccPerms))
	for k := range suite.maccPerms {
		maccs = append(maccs, k)
	}

	sort.Strings(maccs)
	storedMaccs := suite.setModuleAccounts(suite.ctx, suite.accountKeeper, maccs)
	suite.runModuleAccountsIterations(suite.accountKeeper, storedMaccs)
}
