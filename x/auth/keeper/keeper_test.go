package keeper_test

import (
	"testing"

	massert "github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	dbm "github.com/tendermint/tm-db"
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
	suite.app, suite.ctx = createTestApp(suite.T(), true)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.AccountKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func TestAccountMapperGetSet(t *testing.T) {
	app, ctx := createTestApp(t, true)
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
	app, ctx := createTestApp(t, true)
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
	app, ctx := createTestApp(t, true)
	params := types.DefaultParams()

	app.AccountKeeper.SetParams(ctx, params)

	actualParams := app.AccountKeeper.GetParams(ctx)
	require.Equal(t, params, actualParams)
}

func TestSupply_ValidatePermissions(t *testing.T) {
	app, _ := createTestApp(t, true)

	// add module accounts to supply keeper
	maccPerms := simapp.GetMaccPerms()
	maccPerms[holder] = nil
	maccPerms[types.Burner] = []string{types.Burner}
	maccPerms[types.Minter] = []string{types.Minter}
	maccPerms[multiPerm] = []string{types.Burner, types.Minter, types.Staking}
	maccPerms[randomPerm] = []string{"random"}

	cdc := simapp.MakeTestEncodingConfig().Codec
	keeper := keeper.NewAccountKeeper(
		cdc, app.GetKey(types.StoreKey), app.GetSubspace(types.ModuleName),
		types.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
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

func TestInitGenesis(tt *testing.T) {
	authKey := sdk.NewKVStoreKey(types.StoreKey)
	paramsKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	paramsTKey := sdk.NewTransientStoreKey(paramstypes.TStoreKey)
	newCtx := func(t *testing.T) sdk.Context {
		db := dbm.NewMemDB()
		cms := store.NewCommitMultiStore(db)
		cms.MountStoreWithDB(authKey, storetypes.StoreTypeIAVL, db)
		cms.MountStoreWithDB(paramsKey, storetypes.StoreTypeIAVL, db)
		cms.MountStoreWithDB(paramsTKey, storetypes.StoreTypeTransient, db)
		err := cms.LoadLatestVersion()
		if err != nil {
			panic(err)
		}
		return sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())
	}
	newAccountKeeper := func(t *testing.T) keeper.AccountKeeper {
		encConf := simapp.MakeTestEncodingConfig()
		paramsKeeper := paramskeeper.NewKeeper(encConf.Codec, encConf.Amino, paramsKey, paramsTKey)
		return keeper.NewAccountKeeper(
			encConf.Codec, authKey, paramsKeeper.Subspace(types.ModuleName),
			types.ProtoBaseAccount, map[string][]string{}, sdk.Bech32MainPrefix,
		)
	}

	tt.Run("params are set", func(t *testing.T) {
		genState := types.GenesisState{
			Params: types.Params{
				MaxMemoCharacters:      types.DefaultMaxMemoCharacters + 1,
				TxSigLimit:             types.DefaultTxSigLimit + 1,
				TxSizeCostPerByte:      types.DefaultTxSizeCostPerByte + 1,
				SigVerifyCostED25519:   types.DefaultSigVerifyCostED25519 + 1,
				SigVerifyCostSecp256k1: types.DefaultSigVerifyCostSecp256k1 + 1,
			},
			Accounts: []*codectypes.Any{},
		}

		accKeeper := newAccountKeeper(t)
		ctx := newCtx(t)

		accKeeper.InitGenesis(ctx, genState)

		params := accKeeper.GetParams(ctx)
		massert.Equal(t, genState.Params.MaxMemoCharacters, params.MaxMemoCharacters, "MaxMemoCharacters")
		massert.Equal(t, genState.Params.TxSigLimit, params.TxSigLimit, "TxSigLimit")
		massert.Equal(t, genState.Params.TxSizeCostPerByte, params.TxSizeCostPerByte, "TxSizeCostPerByte")
		massert.Equal(t, genState.Params.SigVerifyCostED25519, params.SigVerifyCostED25519, "SigVerifyCostED25519")
		massert.Equal(t, genState.Params.SigVerifyCostSecp256k1, params.SigVerifyCostSecp256k1, "SigVerifyCostSecp256k1")
	})

	tt.Run("duplicate account numbers are fixed", func(t *testing.T) {
		pubKey1 := ed25519.GenPrivKey().PubKey()
		pubKey2 := ed25519.GenPrivKey().PubKey()
		accts := []types.AccountI{
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
		genState := types.GenesisState{
			Params:   types.DefaultParams(),
			Accounts: nil,
		}
		for _, acct := range accts {
			genState.Accounts = append(genState.Accounts, codectypes.UnsafePackAny(acct))
		}

		accKeeper := newAccountKeeper(t)
		ctx := newCtx(t)

		accKeeper.InitGenesis(ctx, genState)

		keeperAccts := accKeeper.GetAllAccounts(ctx)
		assert.GreaterOrEqual(t, len(keeperAccts), len(accts), "number of accounts in the keeper vs in genesis state")
		for i, genAcct := range accts {
			genAcctAddr := genAcct.GetAddress()
			var keeperAcct types.AccountI
			for _, kacct := range keeperAccts {
				if genAcctAddr.Equals(kacct.GetAddress()) {
					keeperAcct = kacct
					break
				}
			}
			if assert.NotNilf(t, keeperAcct, "genesis account %s not in keeper accounts", genAcctAddr) {
				assert.Equal(t, genAcct.GetPubKey(), keeperAcct.GetPubKey())
				assert.Equal(t, genAcct.GetSequence(), keeperAcct.GetSequence())
				if i == 1 {
					assert.Equal(t, 1, int(keeperAcct.GetAccountNumber()))
				} else {
					assert.Equal(t, genAcct.GetSequence(), keeperAcct.GetSequence())
				}
			}
		}

		// The 3rd account has account number 5, so the next should be 6.
		nextNum := accKeeper.GetNextAccountNumber(ctx)
		assert.Equal(t, 6, int(nextNum))
	})

	tt.Run("one zero account still sets global account number", func(t *testing.T) {
		pubKey1 := ed25519.GenPrivKey().PubKey()
		genState := types.GenesisState{
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

		accKeeper := newAccountKeeper(t)
		ctx := newCtx(t)

		accKeeper.InitGenesis(ctx, genState)

		nextNum := accKeeper.GetNextAccountNumber(ctx)
		assert.Equal(t, 1, int(nextNum))
	})
}
