package bank

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

const (
	costGetCoins      sdk.Gas = 10
	costHasCoins      sdk.Gas = 10
	costSetCoins      sdk.Gas = 100
	costSubtractCoins sdk.Gas = 10
	costAddCoins      sdk.Gas = 10

	costSetDenomMetadata sdk.Gas = 10
	costGetDenomMetadata sdk.Gas = 10
	costMintCoins        sdk.Gas = 10
	costBurnCoins        sdk.Gas = 10
)

//-----------------------------------------------------------------------------
// Keeper

var _ Keeper = (*BaseKeeper)(nil)

// Keeper defines a module interface that facilitates the transfer of coins
// between accounts.
type Keeper interface {
	SendKeeper

	MintCoins(ctx sdk.Context, amt sdk.Coins) sdk.Error
	BurnCoins(ctx sdk.Context, amt sdk.Coins) sdk.Error
	MintAddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error
	BurnSubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error
	SetDenomSupply(ctx sdk.Context, totalSupply sdk.Coin) sdk.Error
	SetDenomDecimals(ctx sdk.Context, denom string, decimals uint8) sdk.Error
}

// BaseKeeper manages transfers between accounts. It implements the Keeper
// interface.
type BaseKeeper struct {
	BaseSendKeeper

	ak auth.AccountKeeper
}

// NewBaseKeeper returns a new BaseKeeper
func NewBaseKeeper(cdc *codec.Codec, ak auth.AccountKeeper, storeKey sdk.StoreKey) BaseKeeper {
	return BaseKeeper{
		BaseSendKeeper: NewBaseSendKeeper(cdc, ak, storeKey),
		ak:             ak,
	}
}

// Increases the total supply of a specific denom
func (keeper BaseKeeper) MintCoins(ctx sdk.Context, amt sdk.Coins) sdk.Error {
	return mintCoins(ctx, keeper.cdc, keeper.metadataStoreKey, amt)
}

// Decreases the total supply of a specific denom
func (keeper BaseKeeper) BurnCoins(ctx sdk.Context, amt sdk.Coins) sdk.Error {
	return burnCoins(ctx, keeper.cdc, keeper.metadataStoreKey, amt)
}

// Adds coins to an account and increases the total supply of a specific denom
func (keeper BaseKeeper) MintAddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	err := keeper.MintCoins(ctx, amt)
	if err != nil {
		return err
	}
	_, _, err = keeper.AddCoins(ctx, addr, amt)
	return err
}

// Subtracts coins to an account and decreases the total supply of a specific denom
func (keeper BaseKeeper) BurnSubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	err := keeper.BurnCoins(ctx, amt)
	if err != nil {
		return err
	}
	_, _, err = keeper.SubtractCoins(ctx, addr, amt)
	return err
}

// SetDenomSupply sets the total supply of a specific denom
func (keeper BaseViewKeeper) SetDenomSupply(ctx sdk.Context, totalSupply sdk.Coin) sdk.Error {
	return setDenomSupply(ctx, keeper.cdc, keeper.metadataStoreKey, totalSupply)
}

// SetDenomDecimals sets the decimals of a specific denom
func (keeper BaseViewKeeper) SetDenomDecimals(ctx sdk.Context, denom string, decimals uint8) sdk.Error {
	return setDenomDecimals(ctx, keeper.cdc, keeper.metadataStoreKey, denom, decimals)
}

// InputOutputCoins handles a list of inputs and outputs
func (keeper BaseKeeper) InputOutputCoins(
	ctx sdk.Context, inputs []Input, outputs []Output,
) (sdk.Tags, sdk.Error) {

	return inputOutputCoins(ctx, keeper.ak, inputs, outputs)
}

//-----------------------------------------------------------------------------
// Send Keeper

// SendKeeper defines a module interface that facilitates the transfer of coins
// between accounts without the possibility of creating coins.
type SendKeeper interface {
	ViewKeeper

	SetCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
	InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) (sdk.Tags, sdk.Error)
}

var _ SendKeeper = (*BaseSendKeeper)(nil)

// SendKeeper only allows transfers between accounts without the possibility of
// creating coins. It implements the SendKeeper interface.
type BaseSendKeeper struct {
	BaseViewKeeper

	ak auth.AccountKeeper
}

// NewBaseSendKeeper returns a new BaseSendKeeper.
func NewBaseSendKeeper(cdc *codec.Codec, ak auth.AccountKeeper, storeKey sdk.StoreKey) BaseSendKeeper {
	return BaseSendKeeper{
		BaseViewKeeper: NewBaseViewKeeper(cdc, ak, storeKey),
		ak:             ak,
	}
}

// SetCoins sets the coins at the addr.
func (keeper BaseSendKeeper) SetCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return setCoins(ctx, keeper.ak, addr, amt)
}

// SubtractCoins subtracts amt from the coins at the addr.
func (keeper BaseSendKeeper) SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error) {
	return subtractCoins(ctx, keeper.ak, addr, amt)
}

// AddCoins adds amt to the coins at the addr.
func (keeper BaseSendKeeper) AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error) {
	return addCoins(ctx, keeper.ak, addr, amt)
}

// SendCoins moves coins from one account to another
func (keeper BaseSendKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error) {
	return sendCoins(ctx, keeper.ak, fromAddr, toAddr, amt)
}

// InputOutputCoins handles a list of inputs and outputs
func (keeper BaseSendKeeper) InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) (sdk.Tags, sdk.Error) {
	return inputOutputCoins(ctx, keeper.ak, inputs, outputs)
}

//-----------------------------------------------------------------------------
// View Keeper

var _ ViewKeeper = (*BaseViewKeeper)(nil)

// ViewKeeper defines a module interface that facilitates read only access to
// account balances.
type ViewKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool

	GetDenomSupply(ctx sdk.Context, denom string) (sdk.Coin, sdk.Error)
	GetDenomDecimals(ctx sdk.Context, denom string) (uint8, sdk.Error)
}

// BaseViewKeeper implements a read only keeper implementation of ViewKeeper.
type BaseViewKeeper struct {
	ak auth.AccountKeeper
	// The wire codec for binary encoding/decoding.
	cdc *codec.Codec
	// The (unexposed) keys used to access the stores from the Context.
	metadataStoreKey sdk.StoreKey
}

// NewBaseViewKeeper returns a new BaseViewKeeper.
func NewBaseViewKeeper(cdc *codec.Codec, ak auth.AccountKeeper, storeKey sdk.StoreKey) BaseViewKeeper {
	return BaseViewKeeper{
		cdc:              cdc,
		ak:               ak,
		metadataStoreKey: storeKey,
	}
}

// GetCoins returns the coins at the addr.
func (keeper BaseViewKeeper) GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return getCoins(ctx, keeper.ak, addr)
}

// HasCoins returns whether or not an account has at least amt coins.
func (keeper BaseViewKeeper) HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool {
	return hasCoins(ctx, keeper.ak, addr, amt)
}

// GetDenomSupply returns the total supply of a specific denom
func (keeper BaseViewKeeper) GetDenomSupply(ctx sdk.Context, denom string) (sdk.Coin, sdk.Error) {
	return getDenomSupply(ctx, keeper.cdc, keeper.metadataStoreKey, denom)
}

// GetDenomDecimals returns the decimals of a specific denom
func (keeper BaseViewKeeper) GetDenomDecimals(ctx sdk.Context, denom string) (uint8, sdk.Error) {
	return getDenomDecimals(ctx, keeper.cdc, keeper.metadataStoreKey, denom)
}

//-----------------------------------------------------------------------------

func getCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress) sdk.Coins {
	ctx.GasMeter().ConsumeGas(costGetCoins, "getCoins")
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		return sdk.Coins{}
	}
	return acc.GetCoins()
}

func setCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	ctx.GasMeter().ConsumeGas(costSetCoins, "setCoins")
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		acc = am.NewAccountWithAddress(ctx, addr)
	}
	err := acc.SetCoins(amt)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	am.SetAccount(ctx, acc)
	return nil
}

// HasCoins returns whether or not an account has at least amt coins.
func hasCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress, amt sdk.Coins) bool {
	ctx.GasMeter().ConsumeGas(costHasCoins, "hasCoins")
	return getCoins(ctx, am, addr).IsAllGTE(amt)
}

// SubtractCoins subtracts amt from the coins at the addr.
func subtractCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error) {
	ctx.GasMeter().ConsumeGas(costSubtractCoins, "subtractCoins")

	oldCoins := getCoins(ctx, am, addr)
	newCoins, hasNeg := oldCoins.SafeMinus(amt)
	if hasNeg {
		return amt, nil, sdk.ErrInsufficientCoins(fmt.Sprintf("%s < %s", oldCoins, amt))
	}

	err := setCoins(ctx, am, addr, newCoins)
	tags := sdk.NewTags("sender", []byte(addr.String()))
	return newCoins, tags, err
}

// AddCoins adds amt to the coins at the addr.
func addCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error) {
	ctx.GasMeter().ConsumeGas(costAddCoins, "addCoins")
	oldCoins := getCoins(ctx, am, addr)
	newCoins := oldCoins.Plus(amt)
	if !newCoins.IsNotNegative() {
		return amt, nil, sdk.ErrInsufficientCoins(fmt.Sprintf("%s < %s", oldCoins, amt))
	}
	err := setCoins(ctx, am, addr, newCoins)
	tags := sdk.NewTags("recipient", []byte(addr.String()))
	return newCoins, tags, err
}

// SendCoins moves coins from one account to another
// NOTE: Make sure to revert state changes from tx on error
func sendCoins(ctx sdk.Context, am auth.AccountKeeper, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error) {
	_, subTags, err := subtractCoins(ctx, am, fromAddr, amt)
	if err != nil {
		return nil, err
	}

	_, addTags, err := addCoins(ctx, am, toAddr, amt)
	if err != nil {
		return nil, err
	}

	return subTags.AppendTags(addTags), nil
}

// InputOutputCoins handles a list of inputs and outputs
// NOTE: Make sure to revert state changes from tx on error
func inputOutputCoins(ctx sdk.Context, am auth.AccountKeeper, inputs []Input, outputs []Output) (sdk.Tags, sdk.Error) {
	allTags := sdk.EmptyTags()

	for _, in := range inputs {
		_, tags, err := subtractCoins(ctx, am, in.Address, in.Coins)
		if err != nil {
			return nil, err
		}
		allTags = allTags.AppendTags(tags)
	}

	for _, out := range outputs {
		_, tags, err := addCoins(ctx, am, out.Address, out.Coins)
		if err != nil {
			return nil, err
		}
		allTags = allTags.AppendTags(tags)
	}

	return allTags, nil
}

func mintCoins(ctx sdk.Context, cdc *codec.Codec, storeKey sdk.StoreKey, amt sdk.Coins) sdk.Error {
	ctx.GasMeter().ConsumeGas(costMintCoins, "mintCoins")
	for _, coin := range amt {
		supply, err := getDenomSupply(ctx, cdc, storeKey, coin.Denom)
		if err != nil {
			return err
		}
		supply = supply.Plus(coin)
		err = setDenomSupply(ctx, cdc, storeKey, supply)
		if err != nil {
			return err
		}
	}
	return nil
}

func burnCoins(ctx sdk.Context, cdc *codec.Codec, storeKey sdk.StoreKey, amt sdk.Coins) sdk.Error {
	ctx.GasMeter().ConsumeGas(costBurnCoins, "burnCoins")
	for _, coin := range amt {
		supply, err := getDenomSupply(ctx, cdc, storeKey, coin.Denom)
		if err != nil {
			return err
		}
		supply = supply.Minus(coin)
		err = setDenomSupply(ctx, cdc, storeKey, supply)
		if err != nil {
			return err
		}
	}
	return nil
}

func getDenomSupply(ctx sdk.Context, cdc *codec.Codec, storeKey sdk.StoreKey, denom string) (totalSupplyCoin sdk.Coin, err sdk.Error) {
	store := ctx.KVStore(storeKey)
	bz := store.Get(KeySupply(denom))
	if bz == nil {
		return totalSupplyCoin, sdk.ErrInvalidCoins("nonexistent denom")
	}
	cdc.MustUnmarshalBinaryBare(bz, &totalSupplyCoin)
	return totalSupplyCoin, nil
}

func setDenomSupply(ctx sdk.Context, cdc *codec.Codec, storeKey sdk.StoreKey, totalSupply sdk.Coin) sdk.Error {
	store := ctx.KVStore(storeKey)
	bz := cdc.MustMarshalBinaryBare(totalSupply)
	store.Set(KeySupply(totalSupply.Denom), bz)
	return nil
}

func getDenomDecimals(ctx sdk.Context, cdc *codec.Codec, storeKey sdk.StoreKey, denom string) (decimals uint8, err sdk.Error) {
	store := ctx.KVStore(storeKey)
	bz := store.Get(KeyDecimals(denom))
	if bz == nil {
		return decimals, sdk.ErrInvalidCoins("nonexistent denom")
	}
	cdc.MustUnmarshalBinaryBare(bz, &decimals)
	return decimals, nil
}

func setDenomDecimals(ctx sdk.Context, cdc *codec.Codec, storeKey sdk.StoreKey, denom string, decimals uint8) sdk.Error {
	store := ctx.KVStore(storeKey)
	bz := cdc.MustMarshalBinaryBare(decimals)
	store.Set(KeySupply(denom), bz)
	return nil
}
