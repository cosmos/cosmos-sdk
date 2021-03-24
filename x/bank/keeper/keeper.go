package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

	InitGenesis(sdk.Context, *types.GenesisState)
	ExportGenesis(sdk.Context) *types.GenesisState

	GetSupply(ctx sdk.Context) exported.SupplyI
	SetSupply(ctx sdk.Context, supply exported.SupplyI)

	GetDenomMetaData(ctx sdk.Context, denom string) types.Metadata
	SetDenomMetaData(ctx sdk.Context, denomMetaData types.Metadata)
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
	MarshalSupply(supplyI exported.SupplyI) ([]byte, error)
	UnmarshalSupply(bz []byte) (exported.SupplyI, error)

	types.QueryServer
}

// BaseKeeper manages transfers between accounts. It implements the Keeper interface.
type BaseKeeper struct {
	BaseSendKeeper

	ak         types.AccountKeeper
	cdc        codec.BinaryMarshaler
	storeKey   sdk.StoreKey
	paramSpace paramtypes.Subspace
}

func NewBaseKeeper(
	cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, ak types.AccountKeeper, paramSpace paramtypes.Subspace,
	blockedAddrs map[string]bool,
) BaseKeeper {

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return BaseKeeper{
		BaseSendKeeper: NewBaseSendKeeper(cdc, storeKey, ak, paramSpace, blockedAddrs),
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

	err := k.AddCoins(ctx, moduleAccAddr, amt)
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

	err := k.SubtractCoins(ctx, moduleAccAddr, amt)
	if err != nil {
		return err
	}

	if err := k.trackUndelegation(ctx, delegatorAddr, amt); err != nil {
		return sdkerrors.Wrap(err, "failed to track undelegation")
	}

	err = k.AddCoins(ctx, delegatorAddr, amt)
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

	supply, err := k.UnmarshalSupply(bz)
	if err != nil {
		panic(err)
	}

	return supply
}

// SetSupply sets the Supply to store
func (k BaseKeeper) SetSupply(ctx sdk.Context, supply exported.SupplyI) {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.MarshalSupply(supply)
	if err != nil {
		panic(err)
	}

	store.Set(types.SupplyKey, bz)
}

// GetDenomMetaData retrieves the denomination metadata
func (k BaseKeeper) GetDenomMetaData(ctx sdk.Context, denom string) types.Metadata {
	store := ctx.KVStore(k.storeKey)
	store = prefix.NewStore(store, types.DenomMetadataKey(denom))

	bz := store.Get([]byte(denom))
	if bz == nil {
		return types.Metadata{}
	}

	var metadata types.Metadata
	k.cdc.MustUnmarshalBinaryBare(bz, &metadata)

	return metadata
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
	store := ctx.KVStore(k.storeKey)
	denomMetaDataStore := prefix.NewStore(store, types.DenomMetadataPrefix)

	iterator := denomMetaDataStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var metadata types.Metadata
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &metadata)

		if cb(metadata) {
			break
		}
	}
}

// SetDenomMetaData sets the denominations metadata
func (k BaseKeeper) SetDenomMetaData(ctx sdk.Context, denomMetaData types.Metadata) {
	store := ctx.KVStore(k.storeKey)
	denomMetaDataStore := prefix.NewStore(store, types.DenomMetadataKey(denomMetaData.Base))

	m := k.cdc.MustMarshalBinaryBare(&denomMetaData)
	denomMetaDataStore.Set([]byte(denomMetaData.Base), m)
}

// SendCoinsFromModuleToAccount transfers coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if
// the recipient address is black-listed or if sending the tokens fails.
func (k BaseKeeper) SendCoinsFromModuleToAccount(
	ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {

	senderAddr := k.ak.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	if k.BlockedAddr(recipientAddr) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", recipientAddr)
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

	if !recipientAcc.HasPermission(authtypes.Staking) {
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

	if !acc.HasPermission(authtypes.Staking) {
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

	if !acc.HasPermission(authtypes.Minter) {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to mint tokens", moduleName))
	}

	err := k.AddCoins(ctx, acc.GetAddress(), amt)
	if err != nil {
		return err
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply.Inflate(amt)

	k.SetSupply(ctx, supply)

	logger := k.Logger(ctx)
	logger.Info("minted coins from module account", "amount", amt.String(), "from", moduleName)

	return nil
}

// BurnCoins burns coins deletes coins from the balance of the module account.
// It will panic if the module account does not exist or is unauthorized.
func (k BaseKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	acc := k.ak.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}

	if !acc.HasPermission(authtypes.Burner) {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to burn tokens", moduleName))
	}

	err := k.SubtractCoins(ctx, acc.GetAddress(), amt)
	if err != nil {
		return err
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply.Deflate(amt)
	k.SetSupply(ctx, supply)

	logger := k.Logger(ctx)
	logger.Info("burned tokens from module account", "amount", amt.String(), "from", moduleName)

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

// MarshalSupply protobuf serializes a Supply interface
func (k BaseKeeper) MarshalSupply(supplyI exported.SupplyI) ([]byte, error) {
	return k.cdc.MarshalInterface(supplyI)
}

// UnmarshalSupply returns a Supply interface from raw encoded supply
// bytes of a Proto-based Supply type
func (k BaseKeeper) UnmarshalSupply(bz []byte) (exported.SupplyI, error) {
	var evi exported.SupplyI
	return evi, k.cdc.UnmarshalInterface(bz, &evi)
}
