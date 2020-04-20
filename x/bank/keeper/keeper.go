package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ Keeper = (*BaseKeeper)(nil)

// Keeper defines a module interface that facilitates the transfer of coins
// between accounts.
type Keeper interface {
	SendKeeper

	GetSupply(ctx sdk.Context) exported.SupplyI
	SetSupply(ctx sdk.Context, supply exported.SupplyI)

	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	DelegateCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	UndelegateCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error

	DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error
	UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error
}

// BaseKeeper manages transfers between accounts. It implements the Keeper interface.
type BaseKeeper struct {
	BaseSendKeeper

	ak         types.AccountKeeper
	cdc        types.Codec
	storeKey   sdk.StoreKey
	paramSpace paramtypes.Subspace
}

func NewBaseKeeper(
	cdc types.Codec, storeKey sdk.StoreKey, ak types.AccountKeeper, paramSpace paramtypes.Subspace,
	blacklistedAddrs map[string]bool,
) BaseKeeper {

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return BaseKeeper{
		BaseSendKeeper: NewBaseSendKeeper(cdc, storeKey, ak, paramSpace, blacklistedAddrs),
		ak:             ak,
		cdc:            cdc,
		storeKey:       storeKey,
		paramSpace:     paramSpace,
	}
}

// DelegateCoins performs delegation by deducting amt coins from an account with
// address addr. For vesting accounts, delegations amounts are tracked for both
// vesting and vested coins. The coins are then transferred from the delegator
// address to a ModuleAccount address. If any of the delegation amounts are negative,
// an error is returned.
func (k BaseKeeper) DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error {
	moduleAcc := k.ak.GetAccount(ctx, moduleAccAddr)
	if moduleAcc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleAccAddr)
	}

	if !amt.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	balances := sdk.NewCoins()

	for _, coin := range amt {
		balance := k.GetBalance(ctx, delegatorAddr, coin.Denom)
		if balance.IsLT(coin) {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInsufficientFunds, "failed to delegate; %s is smaller than %s", balance, amt,
			)
		}

		balances = balances.Add(balance)
		err := k.SetBalance(ctx, delegatorAddr, balance.Sub(coin))
		if err != nil {
			return err
		}
	}

	if err := k.trackDelegation(ctx, delegatorAddr, ctx.BlockHeader().Time, balances, amt); err != nil {
		return sdkerrors.Wrap(err, "failed to track delegation")
	}

	_, err := k.AddCoins(ctx, moduleAccAddr, amt)
	if err != nil {
		return err
	}

	return nil
}

// UndelegateCoins performs undelegation by crediting amt coins to an account with
// address addr. For vesting accounts, undelegation amounts are tracked for both
// vesting and vested coins. The coins are then transferred from a ModuleAccount
// address to the delegator address. If any of the undelegation amounts are
// negative, an error is returned.
func (k BaseKeeper) UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error {
	moduleAcc := k.ak.GetAccount(ctx, moduleAccAddr)
	if moduleAcc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleAccAddr)
	}

	if !amt.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	_, err := k.SubtractCoins(ctx, moduleAccAddr, amt)
	if err != nil {
		return err
	}

	if err := k.trackUndelegation(ctx, delegatorAddr, amt); err != nil {
		return sdkerrors.Wrap(err, "failed to track undelegation")
	}

	_, err = k.AddCoins(ctx, delegatorAddr, amt)
	if err != nil {
		return err
	}

	return nil
}

// GetSupply retrieves the Supply from store
func (k BaseKeeper) GetSupply(ctx sdk.Context) exported.SupplyI {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.SupplyKey)
	if bz == nil {
		panic("stored supply should not have been nil")
	}

	supply, err := k.cdc.UnmarshalSupply(bz)
	if err != nil {
		panic(err)
	}

	return supply
}

// SetSupply sets the Supply to store
func (k BaseKeeper) SetSupply(ctx sdk.Context, supply exported.SupplyI) {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalSupply(supply)
	if err != nil {
		panic(err)
	}

	store.Set(types.SupplyKey, bz)
}

// SendCoinsFromModuleToAccount transfers coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist.
func (k BaseKeeper) SendCoinsFromModuleToAccount(
	ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {

	senderAddr := k.ak.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	return k.SendCoins(ctx, senderAddr, recipientAddr, amt)
}

// SendCoinsFromModuleToModule transfers coins from a ModuleAccount to another.
// It will panic if either module account does not exist.
func (k BaseKeeper) SendCoinsFromModuleToModule(
	ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins,
) error {

	senderAddr := k.ak.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	recipientAcc := k.ak.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	return k.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// SendCoinsFromAccountToModule transfers coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k BaseKeeper) SendCoinsFromAccountToModule(
	ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {

	recipientAcc := k.ak.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	return k.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// DelegateCoinsFromAccountToModule delegates coins and transfers them from a
// delegator account to a module account. It will panic if the module account
// does not exist or is unauthorized.
func (k BaseKeeper) DelegateCoinsFromAccountToModule(
	ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {

	recipientAcc := k.ak.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	if !recipientAcc.HasPermission(auth.Staking) {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to receive delegated coins", recipientModule))
	}

	return k.DelegateCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// UndelegateCoinsFromModuleToAccount undelegates the unbonding coins and transfers
// them from a module account to the delegator account. It will panic if the
// module account does not exist or is unauthorized.
func (k BaseKeeper) UndelegateCoinsFromModuleToAccount(
	ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {

	acc := k.ak.GetModuleAccount(ctx, senderModule)
	if acc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	if !acc.HasPermission(auth.Staking) {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to undelegate coins", senderModule))
	}

	return k.UndelegateCoins(ctx, acc.GetAddress(), recipientAddr, amt)
}

// MintCoins creates new coins from thin air and adds it to the module account.
// It will panic if the module account does not exist or is unauthorized.
func (k BaseKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	acc := k.ak.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}

	if !acc.HasPermission(auth.Minter) {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to mint tokens", moduleName))
	}

	_, err := k.AddCoins(ctx, acc.GetAddress(), amt)
	if err != nil {
		return err
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply.Inflate(amt)

	k.SetSupply(ctx, supply)

	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("minted %s from %s module account", amt.String(), moduleName))

	return nil
}

// BurnCoins burns coins deletes coins from the balance of the module account.
// It will panic if the module account does not exist or is unauthorized.
func (k BaseKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	acc := k.ak.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}

	if !acc.HasPermission(auth.Burner) {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to burn tokens", moduleName))
	}

	_, err := k.SubtractCoins(ctx, acc.GetAddress(), amt)
	if err != nil {
		return err
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply.Deflate(amt)
	k.SetSupply(ctx, supply)

	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("burned %s from %s module account", amt.String(), moduleName))

	return nil
}

func (k BaseKeeper) trackDelegation(ctx sdk.Context, addr sdk.AccAddress, blockTime time.Time, balance, amt sdk.Coins) error {
	acc := k.ak.GetAccount(ctx, addr)
	if acc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	vacc, ok := acc.(vestexported.VestingAccount)
	if ok {
		// TODO: return error on account.TrackDelegation
		vacc.TrackDelegation(blockTime, balance, amt)
	}

	return nil
}

func (k BaseKeeper) trackUndelegation(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	acc := k.ak.GetAccount(ctx, addr)
	if acc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	vacc, ok := acc.(vestexported.VestingAccount)
	if ok {
		// TODO: return error on account.TrackUndelegation
		vacc.TrackUndelegation(amt)
	}

	return nil
}
