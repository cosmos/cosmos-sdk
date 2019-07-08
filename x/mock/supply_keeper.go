package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/supply/exported"

)

// ModuleAccount defines an account for modules that holds coins on a pool
type ModuleAccount struct {
	*types.BaseAccount
	Name       string `json:"name"`       // name of the module
	Permission string `json:"permission"` // permission of module account (minter/burner/holder)
}


// GetName returns the the name of the holder's module
func (ma ModuleAccount) GetName() string {
	return ma.Name
}

// GetPermission returns permission granted to the module account (holder/minter/burner)
func (ma ModuleAccount) GetPermission() string {
	return ma.Permission
}


// DummySupplyKeeper defines a supply keeper used only for testing to avoid
// circle dependencies
type DummySupplyKeeper struct {
	ak auth.AccountKeeper
}

// NewDummySupplyKeeper creates a DummySupplyKeeper instance
func NewDummySupplyKeeper(ak auth.AccountKeeper) DummySupplyKeeper {
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

// Send Coins from module to account
func (sk DummySupplyKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	moduleAcc := sk.GetModuleAccount(ctx, senderModule)
	toAcc := sk.ak.GetAccount(ctx, toAddr)

	newFromCoins, hasNeg := moduleAcc.GetCoins().SafeSub(amt)
	if hasNeg {
		return sdk.ErrInsufficientCoins(moduleAcc.GetCoins().String())
	}

	newToCoins := toAcc.GetCoins().Add(amt)

	if err := moduleAcc.SetCoins(newFromCoins); err != nil {
		return sdk.ErrInternal(err.Error())
	}

	if err := toAcc.SetCoins(newToCoins); err != nil {
		return sdk.ErrInternal(err.Error())
	}

	sk.ak.SetAccount(ctx, moduleAcc)
	sk.ak.SetAccount(ctx, toAcc)

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
	macc := &ModuleAccount{
		BaseAccount: &baseAcc,
		Name:        moduleName,
		Permission:  "basic",
	}

	maccI := (sk.ak.NewAccount(ctx, macc)).(exported.ModuleAccountI)
	sk.ak.SetAccount(ctx, maccI)
	return maccI
}

// GetModuleAddress for dummy supply keeper
func (sk DummySupplyKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(moduleName)))
}
