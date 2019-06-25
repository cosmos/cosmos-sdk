// nolint
package auth

import (
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/cosmos/cosmos-sdk/x/supply/exported"
	supplytypes "github.com/cosmos/cosmos-sdk/x/supply/types"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	ak  AccountKeeper
	sk  types.SupplyKeeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	types.RegisterCodec(cdc)
	supplytypes.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	keyParams := sdk.NewKVStoreKey("subspace")
	tkeyParams := sdk.NewTransientStoreKey("transient_subspace")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	ps := subspace.NewSubspace(cdc, keyParams, tkeyParams, types.DefaultParamspace)
	ak := NewAccountKeeper(cdc, authCapKey, ps, types.ProtoBaseAccount)
	sk := NewDummySupplyKeeper(ak)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	//feeCollector := ak.NewAccountWithAddress(ctx, types.FeeCollectorAddr)
	//ak.SetAccount(ctx, feeCollector)

	ak.SetParams(ctx, types.DefaultParams())

	return testInput{cdc: cdc, ctx: ctx, ak: ak, sk: sk}
}

// DummySupplyKeeper defines a supply keeper used only for testing to avoid
// circle dependencies
type DummySupplyKeeper struct {
	ak AccountKeeper
}

// NewDummySupplyKeeper creates a DummySupplyKeeper instance
func NewDummySupplyKeeper(ak AccountKeeper) DummySupplyKeeper {
	return DummySupplyKeeper{ak}
}

// SendCoinsFromAccountToModule for the dummy supply keeper
func (sk DummySupplyKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, fromAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error {

	fromAcc := sk.ak.GetAccount(ctx, fromAddr)
	moduleAcc := sk.GetModuleAccount(ctx, recipientModule)

	newFromCoins, hasNeg := fromAcc.GetCoins().SafeSub(amt)
	if hasNeg {
		return sdk.ErrInsufficientCoins(fromAcc.GetCoins().String())
	}

	newToCoins := moduleAcc.GetCoins().Add(amt)

	if err := fromAcc.SetCoins(newFromCoins); err != nil {
		return sdk.ErrInternal(err.Error())
	}

	if err := moduleAcc.SetCoins(newToCoins); err != nil {
		return sdk.ErrInternal(err.Error())
	}

	sk.ak.SetAccount(ctx, fromAcc)
	sk.ak.SetAccount(ctx, moduleAcc)

	return nil
}

// GetModuleAccount for dummy supply keeper
func (sk DummySupplyKeeper) GetModuleAccount(ctx sdk.Context, moduleName string) exported.ModuleAccountI {
	addr := sk.GetModuleAddress(moduleName)

	acc := sk.ak.GetAccount(ctx, addr)
	if acc != nil {
		macc, ok := acc.(exported.ModuleAccountI)
		if !ok {
			return nil
		}
		return macc
	}

	// create a new module account
	macc := supplytypes.NewModuleAccount(moduleName, "holder")
	return (sk.ak.NewAccount(ctx, macc)).(exported.ModuleAccountI) // set the account number
}

// GetModuleAddress for dummy supply keeper
func (sk DummySupplyKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return supplytypes.NewModuleAddress(moduleName)
}
