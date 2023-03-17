package keeper

import (
	"fmt"

	"cosmossdk.io/math"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ Keeper = (*BaseKeeper)(nil)

// Keeper defines a module interface that facilitates the transfer of coins
// between accounts.
type Keeper interface {
	SendKeeper
	WithMintCoinsRestriction(MintingRestrictionFn) BaseKeeper

	InitGenesis(sdk.Context, *types.GenesisState)
	ExportGenesis(sdk.Context) *types.GenesisState

	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	HasSupply(ctx sdk.Context, denom string) bool
	GetPaginatedTotalSupply(ctx sdk.Context, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error)
	IterateTotalSupply(ctx sdk.Context, cb func(sdk.Coin) bool)
	GetDenomMetaData(ctx sdk.Context, denom string) (types.Metadata, bool)
	HasDenomMetaData(ctx sdk.Context, denom string) bool
	SetDenomMetaData(ctx sdk.Context, denomMetaData types.Metadata)
	GetAllDenomMetaData(ctx sdk.Context) []types.Metadata
	IterateAllDenomMetaData(ctx sdk.Context, cb func(types.Metadata) bool)

	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	DelegateCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	UndelegateCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error

	DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error
	UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error

	types.QueryServer
}

// BaseKeeper manages transfers between accounts. It implements the Keeper interface.
type BaseKeeper struct {
	BaseSendKeeper

	ak                     types.AccountKeeper
	cdc                    codec.BinaryCodec
	storeKey               storetypes.StoreKey
	mintCoinsRestrictionFn MintingRestrictionFn
}

type MintingRestrictionFn func(ctx sdk.Context, coins sdk.Coins) error

// GetPaginatedTotalSupply queries for the supply, ignoring 0 coins, with a given pagination
func (k BaseKeeper) GetPaginatedTotalSupply(ctx sdk.Context, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error) {
	results, pageResp, err := query.CollectionPaginate[string, math.Int](ctx, k.Supply, pagination)
	if err != nil {
		return nil, nil, err
	}
	coins := sdk.NewCoins()
	for _, res := range results {
		coins = coins.Add(sdk.NewCoin(res.Key, res.Value))
	}

	return coins, pageResp, nil
}

// NewBaseKeeper returns a new BaseKeeper object with a given codec, dedicated
// store key, an AccountKeeper implementation, and a parameter Subspace used to
// store and fetch module parameters. The BaseKeeper also accepts a
// blocklist map. This blocklist describes the set of addresses that are not allowed
// to receive funds through direct and explicit actions, for example, by using a MsgSend or
// by using a SendCoinsFromModuleToAccount execution.
func NewBaseKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ak types.AccountKeeper,
	blockedAddrs map[string]bool,
	authority string,
) BaseKeeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Errorf("invalid bank authority address: %w", err))
	}

	return BaseKeeper{
		BaseSendKeeper:         NewBaseSendKeeper(cdc, storeKey, ak, blockedAddrs, authority),
		ak:                     ak,
		cdc:                    cdc,
		storeKey:               storeKey,
		mintCoinsRestrictionFn: func(ctx sdk.Context, coins sdk.Coins) error { return nil },
	}
}

// WithMintCoinsRestriction restricts the bank Keeper used within a specific module to
// have restricted permissions on minting via function passed in parameter.
// Previous restriction functions can be nested as such:
//
//	bankKeeper.WithMintCoinsRestriction(restriction1).WithMintCoinsRestriction(restriction2)
func (k BaseKeeper) WithMintCoinsRestriction(check MintingRestrictionFn) BaseKeeper {
	oldRestrictionFn := k.mintCoinsRestrictionFn
	k.mintCoinsRestrictionFn = func(ctx sdk.Context, coins sdk.Coins) error {
		err := check(ctx, coins)
		if err != nil {
			return err
		}
		err = oldRestrictionFn(ctx, coins)
		if err != nil {
			return err
		}
		return nil
	}
	return k
}

// DelegateCoins performs delegation by deducting amt coins from an account with
// address addr. For vesting accounts, delegations amounts are tracked for both
// vesting and vested coins. The coins are then transferred from the delegator
// address to a ModuleAccount address. If any of the delegation amounts are negative,
// an error is returned.
func (k BaseKeeper) DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error {
	moduleAcc := k.ak.GetAccount(ctx, moduleAccAddr)
	if moduleAcc == nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleAccAddr)
	}

	if !amt.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	balances := sdk.NewCoins()

	for _, coin := range amt {
		balance := k.GetBalance(ctx, delegatorAddr, coin.GetDenom())
		if balance.IsLT(coin) {
			return errorsmod.Wrapf(
				sdkerrors.ErrInsufficientFunds, "failed to delegate; %s is smaller than %s", balance, amt,
			)
		}

		balances = balances.Add(balance)
		err := k.setBalance(ctx, delegatorAddr, balance.Sub(coin))
		if err != nil {
			return err
		}
	}

	if err := k.trackDelegation(ctx, delegatorAddr, balances, amt); err != nil {
		return errorsmod.Wrap(err, "failed to track delegation")
	}
	// emit coin spent event
	ctx.EventManager().EmitEvent(
		types.NewCoinSpentEvent(delegatorAddr, amt),
	)

	err := k.addCoins(ctx, moduleAccAddr, amt)
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
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleAccAddr)
	}

	if !amt.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	err := k.subUnlockedCoins(ctx, moduleAccAddr, amt)
	if err != nil {
		return err
	}

	if err := k.trackUndelegation(ctx, delegatorAddr, amt); err != nil {
		return errorsmod.Wrap(err, "failed to track undelegation")
	}

	err = k.addCoins(ctx, delegatorAddr, amt)
	if err != nil {
		return err
	}

	return nil
}

// GetSupply retrieves the Supply from store
func (k BaseKeeper) GetSupply(ctx sdk.Context, denom string) sdk.Coin {
	amt, err := k.Supply.Get(ctx, denom)
	if err != nil {
		return sdk.NewCoin(denom, math.ZeroInt())
	}
	return sdk.NewCoin(denom, amt)
}

// HasSupply checks if the supply coin exists in store.
func (k BaseKeeper) HasSupply(ctx sdk.Context, denom string) bool {
	has, err := k.Supply.Has(ctx, denom)
	return has && err == nil
}

// GetDenomMetaData retrieves the denomination metadata. returns the metadata and true if the denom exists,
// false otherwise.
func (k BaseKeeper) GetDenomMetaData(ctx sdk.Context, denom string) (types.Metadata, bool) {
	m, err := k.BaseViewKeeper.DenomMetadata.Get(ctx, denom)
	return m, err == nil
}

// HasDenomMetaData checks if the denomination metadata exists in store.
func (k BaseKeeper) HasDenomMetaData(ctx sdk.Context, denom string) bool {
	has, err := k.BaseViewKeeper.DenomMetadata.Has(ctx, denom)
	return has && err == nil
}

// GetAllDenomMetaData retrieves all denominations metadata
func (k BaseKeeper) GetAllDenomMetaData(ctx sdk.Context) []types.Metadata {
	denomMetaData := make([]types.Metadata, 0)
	k.IterateAllDenomMetaData(ctx, func(metadata types.Metadata) bool {
		denomMetaData = append(denomMetaData, metadata)
		return false
	})

	return denomMetaData
}

// IterateAllDenomMetaData iterates over all the denominations metadata and
// provides the metadata to a callback. If true is returned from the
// callback, iteration is halted.
func (k BaseKeeper) IterateAllDenomMetaData(ctx sdk.Context, cb func(types.Metadata) bool) {
	_ = k.BaseViewKeeper.DenomMetadata.Walk(ctx, nil, func(_ string, metadata types.Metadata) bool {
		return cb(metadata)
	})
}

// SetDenomMetaData sets the denominations metadata
func (k BaseKeeper) SetDenomMetaData(ctx sdk.Context, denomMetaData types.Metadata) {
	_ = k.BaseViewKeeper.DenomMetadata.Set(ctx, denomMetaData.Base, denomMetaData)
}

// SendCoinsFromModuleToAccount transfers coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if
// the recipient address is black-listed or if sending the tokens fails.
func (k BaseKeeper) SendCoinsFromModuleToAccount(
	ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {
	senderAddr := k.ak.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	if k.BlockedAddr(recipientAddr) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", recipientAddr)
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
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	recipientAcc := k.ak.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
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
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
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
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	if !recipientAcc.HasPermission(authtypes.Staking) {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to receive delegated coins", recipientModule))
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
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	if !acc.HasPermission(authtypes.Staking) {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to undelegate coins", senderModule))
	}

	return k.UndelegateCoins(ctx, acc.GetAddress(), recipientAddr, amt)
}

// MintCoins creates new coins from thin air and adds it to the module account.
// It will panic if the module account does not exist or is unauthorized.
func (k BaseKeeper) MintCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error {
	err := k.mintCoinsRestrictionFn(ctx, amounts)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Module %q attempted to mint coins %s it doesn't have permission for, error %v", moduleName, amounts, err))
		return err
	}
	acc := k.ak.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}

	if !acc.HasPermission(authtypes.Minter) {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to mint tokens", moduleName))
	}

	err = k.addCoins(ctx, acc.GetAddress(), amounts)
	if err != nil {
		return err
	}

	for _, amount := range amounts {
		supply := k.GetSupply(ctx, amount.GetDenom())
		supply = supply.Add(amount)
		k.setSupply(ctx, supply)
	}

	logger := k.Logger(ctx)
	logger.Debug("minted coins from module account", "amount", amounts.String(), "from", moduleName)

	// emit mint event
	ctx.EventManager().EmitEvent(
		types.NewCoinMintEvent(acc.GetAddress(), amounts),
	)

	return nil
}

// BurnCoins burns coins deletes coins from the balance of the module account.
// It will panic if the module account does not exist or is unauthorized.
func (k BaseKeeper) BurnCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error {
	acc := k.ak.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}

	if !acc.HasPermission(authtypes.Burner) {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to burn tokens", moduleName))
	}

	err := k.subUnlockedCoins(ctx, acc.GetAddress(), amounts)
	if err != nil {
		return err
	}

	for _, amount := range amounts {
		supply := k.GetSupply(ctx, amount.GetDenom())
		supply = supply.Sub(amount)
		k.setSupply(ctx, supply)
	}

	logger := k.Logger(ctx)
	logger.Debug("burned tokens from module account", "amount", amounts.String(), "from", moduleName)

	// emit burn event
	ctx.EventManager().EmitEvent(
		types.NewCoinBurnEvent(acc.GetAddress(), amounts),
	)

	return nil
}

// setSupply sets the supply for the given coin
func (k BaseKeeper) setSupply(ctx sdk.Context, coin sdk.Coin) {
	// Bank invariants and IBC requires to remove zero coins.
	if coin.IsZero() {
		_ = k.Supply.Remove(ctx, coin.Denom)
	} else {
		_ = k.Supply.Set(ctx, coin.Denom, coin.Amount)
	}
}

// trackDelegation tracks the delegation of the given account if it is a vesting account
func (k BaseKeeper) trackDelegation(ctx sdk.Context, addr sdk.AccAddress, balance, amt sdk.Coins) error {
	acc := k.ak.GetAccount(ctx, addr)
	if acc == nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	vacc, ok := acc.(types.VestingAccount)
	if ok {
		// TODO: return error on account.TrackDelegation
		vacc.TrackDelegation(ctx.BlockHeader().Time, balance, amt)
		k.ak.SetAccount(ctx, acc)
	}

	return nil
}

// trackUndelegation trakcs undelegation of the given account if it is a vesting account
func (k BaseKeeper) trackUndelegation(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	acc := k.ak.GetAccount(ctx, addr)
	if acc == nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	vacc, ok := acc.(types.VestingAccount)
	if ok {
		// TODO: return error on account.TrackUndelegation
		vacc.TrackUndelegation(amt)
		k.ak.SetAccount(ctx, acc)
	}

	return nil
}

// IterateTotalSupply iterates over the total supply calling the given cb (callback) function
// with the balance of each coin.
// The iteration stops if the callback returns true.
func (k BaseViewKeeper) IterateTotalSupply(ctx sdk.Context, cb func(sdk.Coin) bool) {
	_ = k.Supply.Walk(ctx, nil, func(s string, m math.Int) bool {
		return cb(sdk.NewCoin(s, m))
	})
}
