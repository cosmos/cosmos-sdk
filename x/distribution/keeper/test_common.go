package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

//nolint: deadcode unused
var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delPk2   = ed25519.GenPrivKey().PubKey()
	delPk3   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
	delAddr2 = sdk.AccAddress(delPk2.Address())
	delAddr3 = sdk.AccAddress(delPk3.Address())

	valOpPk1    = ed25519.GenPrivKey().PubKey()
	valOpPk2    = ed25519.GenPrivKey().PubKey()
	valOpPk3    = ed25519.GenPrivKey().PubKey()
	valOpAddr1  = sdk.ValAddress(valOpPk1.Address())
	valOpAddr2  = sdk.ValAddress(valOpPk2.Address())
	valOpAddr3  = sdk.ValAddress(valOpPk3.Address())
	valAccAddr1 = sdk.AccAddress(valOpPk1.Address()) // generate acc addresses for these validator keys too
	valAccAddr2 = sdk.AccAddress(valOpPk2.Address())
	valAccAddr3 = sdk.AccAddress(valOpPk3.Address())

	valConsPk1   = ed25519.GenPrivKey().PubKey()
	valConsPk2   = ed25519.GenPrivKey().PubKey()
	valConsPk3   = ed25519.GenPrivKey().PubKey()
	valConsAddr1 = sdk.ConsAddress(valConsPk1.Address())
	valConsAddr2 = sdk.ConsAddress(valConsPk2.Address())
	valConsAddr3 = sdk.ConsAddress(valConsPk3.Address())

	// TODO move to common testing package for all modules
	// test addresses
	TestAddrs = []sdk.AccAddress{
		delAddr1, delAddr2, delAddr3,
		valAccAddr1, valAccAddr2, valAccAddr3,
	}

	emptyDelAddr sdk.AccAddress
	emptyValAddr sdk.ValAddress
	emptyPubkey  crypto.PubKey
)

// create a codec used only for testing
func MakeTestCodec() *codec.Codec {
	var cdc = codec.New()
	bank.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	types.RegisterCodec(cdc) // distr
	return cdc
}

// test input with default values
func CreateTestInputDefault(t *testing.T, isCheckTx bool, initPower int64) (
	sdk.Context, auth.AccountKeeper, Keeper, staking.Keeper, types.SupplyKeeper) {

	communityTax := sdk.NewDecWithPrec(2, 2)

	ctx, ak, _, dk, sk, _, supplyKeeper := CreateTestInputAdvanced(t, isCheckTx, initPower, communityTax)
	return ctx, ak, dk, sk, supplyKeeper
}

// hogpodge of all sorts of input required for testing
func CreateTestInputAdvanced(t *testing.T, isCheckTx bool, initPower int64,
	communityTax sdk.Dec) (sdk.Context, auth.AccountKeeper, bank.Keeper,
	Keeper, staking.Keeper, params.Keeper, types.SupplyKeeper) {

	initTokens := sdk.TokensFromConsensusPower(initPower)

	keyDistr := sdk.NewKVStoreKey(types.StoreKey)
	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	ms.MountStoreWithDB(keyDistr, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyStaking, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	cdc := MakeTestCodec()
	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, log.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountKeeper, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	supplyKeeper := supply.NewKeeper(cdc, keySupply, accountKeeper, bankKeeper, supply.DefaultCodespace,
		[]string{auth.FeeCollectorName, types.ModuleName}, []string{}, []string{staking.NotBondedPoolName, staking.BondedPoolName})

	sk := staking.NewKeeper(cdc, keyStaking, tkeyStaking, supplyKeeper, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	sk.SetParams(ctx, staking.DefaultParams())

	keeper := NewKeeper(cdc, keyDistr, pk.Subspace(DefaultParamspace), sk, supplyKeeper, types.DefaultCodespace, auth.FeeCollectorName)

	initCoins := sdk.NewCoins(sdk.NewCoin(sk.BondDenom(ctx), initTokens))
	totalSupply := sdk.NewCoins(sdk.NewCoin(sk.BondDenom(ctx), initTokens.MulRaw(int64(len(TestAddrs)))))
	supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))

	// fill all the addresses with some coins, set the loose pool tokens simultaneously
	for _, addr := range TestAddrs {
		_, err := bankKeeper.AddCoins(ctx, addr, initCoins)
		require.Nil(t, err)
	}

	// create module accounts
	feeCollectorAcc := supply.NewEmptyModuleAccount(auth.FeeCollectorName, supply.Basic)
	notBondedPool := supply.NewEmptyModuleAccount(staking.NotBondedPoolName, supply.Burner)
	bondPool := supply.NewEmptyModuleAccount(staking.BondedPoolName, supply.Burner)
	distrAcc := supply.NewEmptyModuleAccount(types.ModuleName, supply.Basic)

	keeper.supplyKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	keeper.supplyKeeper.SetModuleAccount(ctx, notBondedPool)
	keeper.supplyKeeper.SetModuleAccount(ctx, bondPool)
	keeper.supplyKeeper.SetModuleAccount(ctx, distrAcc)

	// set the distribution hooks on staking
	sk.SetHooks(keeper.Hooks())

	// set genesis items required for distribution
	keeper.SetFeePool(ctx, types.InitialFeePool())
	keeper.SetCommunityTax(ctx, communityTax)
	keeper.SetBaseProposerReward(ctx, sdk.NewDecWithPrec(1, 2))
	keeper.SetBonusProposerReward(ctx, sdk.NewDecWithPrec(4, 2))

	return ctx, accountKeeper, bankKeeper, keeper, sk, pk, supplyKeeper
}
