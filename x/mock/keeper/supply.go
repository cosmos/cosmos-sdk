package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/mock/types"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// MockSupplyKeeper defines a supply keeper used only for testing to avoid
// circle dependencies
type MockSupplyKeeper struct {
	ak types.AccountKeeper
}

// NewMockSupplyKeeper creates a MockSupplyKeeper instance
func NewMockSupplyKeeper(ak types.AccountKeeper) MockSupplyKeeper {
	return MockSupplyKeeper{ak}
}

// SendCoinsFromAccountToModule for the mock supply keeper
func (sk MockSupplyKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, fromAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error {

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

// GetModuleAccount for mock supply keeper
func (sk MockSupplyKeeper) GetModuleAccount(ctx sdk.Context, moduleName string) supplyexported.ModuleAccountI {
	addr := sk.GetModuleAddress(moduleName)

	acc := sk.ak.GetAccount(ctx, addr)
	if acc != nil {
		macc, ok := acc.(supplyexported.ModuleAccountI)
		if ok {
			return macc
		}
	}

	moduleAddress := sk.GetModuleAddress(moduleName)
	baseAcc := NewBaseAccountWithAddress(moduleAddress)

	// create a new module account
	macc := &moduleAccount{
		BaseAccount: &baseAcc,
		name:        moduleName,
		permissions: nil,
	}

	maccI := (sk.ak.NewAccount(ctx, macc)).(supplyexported.ModuleAccountI)
	sk.ak.SetAccount(ctx, maccI)
	return maccI
}

// GetModuleAddress for mock supply keeper
func (sk MockSupplyKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(moduleName)))
}
