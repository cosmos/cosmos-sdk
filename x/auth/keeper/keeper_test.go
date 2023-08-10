package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/header"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
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

	ctx sdk.Context

	queryClient   types.QueryClient
	accountKeeper keeper.AccountKeeper
	msgServer     types.MsgServer
	encCfg        moduletestutil.TestEncodingConfig
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.encCfg = moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})

	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithHeaderInfo(header.Info{})

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
		storeService,
		types.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec("cosmos"),
		"cosmos",
		types.NewModuleAddress("gov").String(),
	)
	suite.msgServer = keeper.NewMsgServerImpl(suite.accountKeeper)
	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(suite.accountKeeper))
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
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

func (suite *KeeperTestSuite) TestInitGenesis() {
	suite.SetupTest() // reset

	// Check if params are set
	genState := types.GenesisState{
		Params: types.Params{
			MaxMemoCharacters:      types.DefaultMaxMemoCharacters + 1,
			TxSigLimit:             types.DefaultTxSigLimit + 1,
			TxSizeCostPerByte:      types.DefaultTxSizeCostPerByte + 1,
			SigVerifyCostED25519:   types.DefaultSigVerifyCostED25519 + 1,
			SigVerifyCostSecp256k1: types.DefaultSigVerifyCostSecp256k1 + 1,
		},
	}

	ctx := suite.ctx
	suite.accountKeeper.InitGenesis(ctx, genState)

	params := suite.accountKeeper.GetParams(ctx)
	suite.Require().Equal(genState.Params.MaxMemoCharacters, params.MaxMemoCharacters, "MaxMemoCharacters")
	suite.Require().Equal(genState.Params.TxSigLimit, params.TxSigLimit, "TxSigLimit")
	suite.Require().Equal(genState.Params.TxSizeCostPerByte, params.TxSizeCostPerByte, "TxSizeCostPerByte")
	suite.Require().Equal(genState.Params.SigVerifyCostED25519, params.SigVerifyCostED25519, "SigVerifyCostED25519")
	suite.Require().Equal(genState.Params.SigVerifyCostSecp256k1, params.SigVerifyCostSecp256k1, "SigVerifyCostSecp256k1")

	suite.SetupTest() // reset
	ctx = suite.ctx
	// Fix duplicate account numbers
	pubKey1 := ed25519.GenPrivKey().PubKey()
	pubKey2 := ed25519.GenPrivKey().PubKey()
	accts := []sdk.AccountI{
		&types.BaseAccount{
			Address:       sdk.AccAddress(pubKey1.Address()).String(),
			PubKey:        codectypes.UnsafePackAny(pubKey1),
			AccountNumber: 0,
			Sequence:      5,
		},
		&types.ModuleAccount{
			BaseAccount: &types.BaseAccount{
				Address:       types.NewModuleAddress("testing").String(),
				PubKey:        nil,
				AccountNumber: 0,
				Sequence:      6,
			},
			Name:        "testing",
			Permissions: nil,
		},
		&types.BaseAccount{
			Address:       sdk.AccAddress(pubKey2.Address()).String(),
			PubKey:        codectypes.UnsafePackAny(pubKey2),
			AccountNumber: 5,
			Sequence:      7,
		},
	}
	genState = types.GenesisState{
		Params:   types.DefaultParams(),
		Accounts: nil,
	}
	for _, acct := range accts {
		genState.Accounts = append(genState.Accounts, codectypes.UnsafePackAny(acct))
	}

	suite.accountKeeper.InitGenesis(ctx, genState)

	keeperAccts := suite.accountKeeper.GetAllAccounts(ctx)
	// len(accts)+1 because we initialize fee_collector account after the genState accounts
	suite.Require().Equal(len(keeperAccts), len(accts)+1, "number of accounts in the keeper vs in genesis state")
	for i, genAcct := range accts {
		genAcctAddr := genAcct.GetAddress()
		var keeperAcct sdk.AccountI
		for _, kacct := range keeperAccts {
			if genAcctAddr.Equals(kacct.GetAddress()) {
				keeperAcct = kacct
				break
			}
		}
		suite.Require().NotNilf(keeperAcct, "genesis account %s not in keeper accounts", genAcctAddr)
		suite.Require().Equal(genAcct.GetPubKey(), keeperAcct.GetPubKey())
		suite.Require().Equal(genAcct.GetSequence(), keeperAcct.GetSequence())
		if i == 1 {
			suite.Require().Equalf(1, int(keeperAcct.GetAccountNumber()), genAcctAddr.String())
		} else {
			suite.Require().Equal(genAcct.GetSequence(), keeperAcct.GetSequence())
		}
	}

	// fee_collector's is the last account to be set, so it has +1 of the highest in the accounts list
	feeCollector := suite.accountKeeper.GetModuleAccount(ctx, "fee_collector")
	suite.Require().Equal(6, int(feeCollector.GetAccountNumber()))

	// The 3rd account has account number 5, but because the FeeCollector account gets initialized last, the next should be 7.
	nextNum := suite.accountKeeper.NextAccountNumber(ctx)
	suite.Require().Equal(7, int(nextNum))

	suite.SetupTest() // reset
	ctx = suite.ctx
	// one zero account still sets global account number
	genState = types.GenesisState{
		Params: types.DefaultParams(),
		Accounts: []*codectypes.Any{
			codectypes.UnsafePackAny(&types.BaseAccount{
				Address:       sdk.AccAddress(pubKey1.Address()).String(),
				PubKey:        codectypes.UnsafePackAny(pubKey1),
				AccountNumber: 0,
				Sequence:      5,
			}),
		},
	}

	suite.accountKeeper.InitGenesis(ctx, genState)

	keeperAccts = suite.accountKeeper.GetAllAccounts(ctx)
	// len(genState.Accounts)+1 because we initialize fee_collector as account number 1 (last)
	suite.Require().Equal(len(keeperAccts), len(genState.Accounts)+1, "number of accounts in the keeper vs in genesis state")

	// Check both accounts account numbers
	suite.Require().Equal(0, int(suite.accountKeeper.GetAccount(ctx, sdk.AccAddress(pubKey1.Address())).GetAccountNumber()))
	feeCollector = suite.accountKeeper.GetModuleAccount(ctx, "fee_collector")
	suite.Require().Equal(1, int(feeCollector.GetAccountNumber()))

	nextNum = suite.accountKeeper.NextAccountNumber(ctx)
	// we expect nextNum to be 2 because we initialize fee_collector as account number 1
	suite.Require().Equal(2, int(nextNum))
}
