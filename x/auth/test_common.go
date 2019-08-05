// nolint
package auth

import (
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/cosmos/cosmos-sdk/x/supply/exported"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	ak  AccountKeeper
	sk  types.SupplyKeeper
}

// moduleAccount defines an account for modules that holds coins on a pool
type moduleAccount struct {
	*types.BaseAccount
	name        string   `json:"name" yaml:"name"`              // name of the module
	permissions []string `json:"permissions" yaml"permissions"` // permissions of module account
}

// HasPermission returns whether or not the module account has permission.
func (ma moduleAccount) HasPermission(permission string) bool {
	for _, perm := range ma.permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// GetName returns the the name of the holder's module
func (ma moduleAccount) GetName() string {
	return ma.name
}

// GetPermissions returns permissions granted to the module account
func (ma moduleAccount) GetPermissions() []string {
	return ma.permissions
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	types.RegisterCodec(cdc)
	cdc.RegisterInterface((*exported.ModuleAccountI)(nil), nil)
	cdc.RegisterConcrete(&moduleAccount{}, "cosmos-sdk/ModuleAccount", nil)
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
		if ok {
			return macc
		}
	}

	moduleAddress := sk.GetModuleAddress(moduleName)
	baseAcc := types.NewBaseAccountWithAddress(moduleAddress)

	// create a new module account
	macc := &moduleAccount{
		BaseAccount: &baseAcc,
		name:        moduleName,
		permissions: []string{"basic"},
	}

	maccI := (sk.ak.NewAccount(ctx, macc)).(exported.ModuleAccountI)
	sk.ak.SetAccount(ctx, maccI)
	return maccI
}

// GetModuleAddress for dummy supply keeper
func (sk DummySupplyKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(moduleName)))
}
