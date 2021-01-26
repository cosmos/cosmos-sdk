package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

// SendCoinsFromModuleToAccount transfers coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist.
func (k Keeper) SendCoinsFromModuleToAccount(
	ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {
	if k.bk.BlacklistedAddr(recipientAddr) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "Address <%s> in blacklist is not allowed", recipientAddr)
	}
	senderAddr := k.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	return k.bk.SendCoins(ctx, senderAddr, recipientAddr, amt)
}

// SendCoinsFromModuleToModule transfers coins from a ModuleAccount to another.
// It will panic if either module account does not exist.
func (k Keeper) SendCoinsFromModuleToModule(
	ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins,
) error {

	senderAddr := k.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	recipientAcc := k.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	return k.bk.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// SendCoinsFromAccountToModule transfers coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k Keeper) SendCoinsFromAccountToModule(
	ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {

	recipientAcc := k.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	return k.bk.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// DelegateCoinsFromAccountToModule delegates coins and transfers them from a
// delegator account to a module account. It will panic if the module account
// does not exist or is unauthorized.
func (k Keeper) DelegateCoinsFromAccountToModule(
	ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {

	recipientAcc := k.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	if !recipientAcc.HasPermission(types.Staking) {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to receive delegated coins", recipientModule))
	}

	return k.bk.DelegateCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// UndelegateCoinsFromModuleToAccount undelegates the unbonding coins and transfers
// them from a module account to the delegator account. It will panic if the
// module account does not exist or is unauthorized.
func (k Keeper) UndelegateCoinsFromModuleToAccount(
	ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {

	acc := k.GetModuleAccount(ctx, senderModule)
	if acc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	if !acc.HasPermission(types.Staking) {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to undelegate coins", senderModule))
	}

	return k.bk.UndelegateCoins(ctx, acc.GetAddress(), recipientAddr, amt)
}

// MintCoins creates new coins from thin air and adds it to the module account.
// It will panic if the module account does not exist or is unauthorized.
func (k Keeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	acc := k.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}

	if !acc.HasPermission(types.Minter) {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to mint tokens", moduleName))
	}

	_, err := k.bk.AddCoins(ctx, acc.GetAddress(), amt)
	if err != nil {
		return err
	}

	// update supply
	for i := 0; i < len(amt); i++ {
		k.inflate(ctx, amt[i].Denom, amt[i].Amount)
	}

	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("minted %s from %s module account", amt.String(), moduleName))

	return nil
}

// BurnCoins burns coins deletes coins from the balance of the module account.
// It will panic if the module account does not exist or is unauthorized.
func (k Keeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	acc := k.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}

	if !acc.HasPermission(types.Burner) {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to burn tokens", moduleName))
	}

	_, err := k.bk.SubtractCoins(ctx, acc.GetAddress(), amt)
	if err != nil {
		return err
	}

	// update supply
	for i := 0; i < len(amt); i++ {
		if err = k.deflate(ctx, amt[i].Denom, amt[i].Amount); err != nil {
			return err
		}
	}

	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("burned %s from %s module account", amt.String(), moduleName))

	return nil
}

// GetSupplyByDenom gets the amount of a token supply from the store
func (k Keeper) GetSupplyByDenom(ctx sdk.Context, denom string) sdk.Dec {
	tokenSupplyAmount := sdk.ZeroDec()
	bytes := ctx.KVStore(k.storeKey).Get(getTokenSupplyKey(denom))
	if bytes == nil {
		return tokenSupplyAmount
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bytes, &tokenSupplyAmount)
	return tokenSupplyAmount
}

// inflate adds the amount of a token in the store
func (k Keeper) inflate(ctx sdk.Context, tokenSymbol string, amount sdk.Dec) {
	if amount.Equal(sdk.ZeroDec()) {
		return
	}

	originalSupplyAmount := k.GetSupplyByDenom(ctx, tokenSymbol)
	k.setTokenSupplyAmount(ctx, tokenSymbol, originalSupplyAmount.Add(amount))
}

// deflate subtracts the amount of a token from the original in the store
func (k Keeper) deflate(ctx sdk.Context, tokenSymbol string, deflationAmount sdk.Dec) error {
	currentSupplyAmount := k.GetSupplyByDenom(ctx, tokenSymbol)
	supplyAmount := currentSupplyAmount.Sub(deflationAmount)
	if supplyAmount.IsNegative() {
		return types.ErrInvalidDeflation
	}

	k.setTokenSupplyAmount(ctx, tokenSymbol, supplyAmount)
	return nil
}
