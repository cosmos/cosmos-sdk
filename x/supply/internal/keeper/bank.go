package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

// SendCoinsFromModuleToAccount transfers coins from a ModuleAccount to an AccAddress
func (k Keeper) SendCoinsFromModuleToAccount(
	ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {

	senderAddr := k.GetModuleAddress(senderModule)
	if senderAddr == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule)
	}

	return k.bk.SendCoins(ctx, senderAddr, recipientAddr, amt)
}

// SendCoinsFromModuleToModule transfers coins from a ModuleAccount to another
func (k Keeper) SendCoinsFromModuleToModule(
	ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins,
) error {

	senderAddr := k.GetModuleAddress(senderModule)
	if senderAddr == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule)
	}

	recipientAcc := k.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule)
	}

	return k.bk.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// SendCoinsFromAccountToModule transfers coins from an AccAddress to a ModuleAccount
func (k Keeper) SendCoinsFromAccountToModule(
	ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {

	recipientAcc := k.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule)
	}

	return k.bk.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// DelegateCoinsFromAccountToModule delegates coins and transfers
// them from a delegator account to a module account
func (k Keeper) DelegateCoinsFromAccountToModule(
	ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {

	recipientAcc := k.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule)
	}

	if !recipientAcc.HasPermission(types.Staking) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to receive delegated coins", recipientModule)
	}

	return k.bk.DelegateCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// UndelegateCoinsFromModuleToAccount undelegates the unbonding coins and transfers
// them from a module account to the delegator account
func (k Keeper) UndelegateCoinsFromModuleToAccount(
	ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {

	acc := k.GetModuleAccount(ctx, senderModule)
	if acc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule)
	}

	if !acc.HasPermission(types.Staking) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to undelegate coins", senderModule)
	}

	return k.bk.UndelegateCoins(ctx, acc.GetAddress(), recipientAddr, amt)
}

// MintCoins creates new coins from thin air and adds it to the module account.
// Panics if the name maps to a non-minter module account or if the amount is invalid.
func (k Keeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	acc := k.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName)
	}

	if !acc.HasPermission(types.Minter) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to mint tokens", moduleName)
	}

	_, err := k.bk.AddCoins(ctx, acc.GetAddress(), amt)
	if err != nil {
		return err
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply = supply.Inflate(amt)

	k.SetSupply(ctx, supply)

	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("minted %s from %s module account", amt.String(), moduleName))

	return nil
}

// BurnCoins burns coins deletes coins from the balance of the module account.
// Panics if the name maps to a non-burner module account or if the amount is invalid.
func (k Keeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	acc := k.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName)
	}

	if !acc.HasPermission(types.Burner) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to burn tokens", moduleName)
	}

	_, err := k.bk.SubtractCoins(ctx, acc.GetAddress(), amt)
	if err != nil {
		return err
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply = supply.Deflate(amt)
	k.SetSupply(ctx, supply)

	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("burned %s from %s module account", amt.String(), moduleName))

	return nil
}
