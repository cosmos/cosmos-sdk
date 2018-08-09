package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

const (
	costGetCoins      sdk.Gas = 10
	costHasCoins      sdk.Gas = 10
	costSetCoins      sdk.Gas = 100
	costSubtractCoins sdk.Gas = 10
	costAddCoins      sdk.Gas = 10

	costSetDenomMetadata sdk.Gas = 0
	costGetDenomMetadata sdk.Gas = 0
	costMintCoins        sdk.Gas = 0
	costBurnCoins        sdk.Gas = 0
)

// Keeper manages transfers between accounts
type Keeper struct {
	am auth.AccountMapper

	// The (unexposed) keys used to access the stores from the Context.
	metadataStoreKey sdk.StoreKey

	// The wire codec for binary encoding/decoding.
	cdc *wire.Codec

	// Reserved codespace
	codespace sdk.CodespaceType
}

// NewKeeper returns a new Keeper
func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, am auth.AccountMapper, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		am:               am,
		metadataStoreKey: key,
		cdc:              cdc,
		codespace:        codespace,
	}
}

// GetCoins returns the coins at the addr.
func (keeper Keeper) GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	ctx.GasMeter().ConsumeGas(costGetCoins, "getCoins")
	acc := keeper.am.GetAccount(ctx, addr)
	if acc == nil {
		return sdk.Coins{}
	}
	return acc.GetCoins()
}

// SetCoins sets the coins at the addr.
func (keeper Keeper) SetCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	ctx.GasMeter().ConsumeGas(costSetCoins, "setCoins")
	acc := keeper.am.GetAccount(ctx, addr)
	if acc == nil {
		acc = keeper.am.NewAccountWithAddress(ctx, addr)
	}
	err := acc.SetCoins(amt)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	keeper.am.SetAccount(ctx, acc)
	return nil
}

// HasCoins returns whether or not an account has at least amt coins.
func (keeper Keeper) HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool {
	ctx.GasMeter().ConsumeGas(costHasCoins, "hasCoins")
	return keeper.GetCoins(ctx, addr).IsGTE(amt)
}

// SubtractCoins subtracts amt from the coins at the addr.
func (keeper Keeper) SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error) {
	ctx.GasMeter().ConsumeGas(costSubtractCoins, "subtractCoins")
	oldCoins := keeper.GetCoins(ctx, addr)
	newCoins := oldCoins.Minus(amt)
	if !newCoins.IsNotNegative() {
		return amt, sdk.ErrInsufficientCoins(fmt.Sprintf("%s < %s", oldCoins, amt))
	}
	err := keeper.SetCoins(ctx, addr, newCoins)
	return newCoins, err
}

// AddCoins adds amt to the coins at the addr.
func (keeper Keeper) AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error) {
	ctx.GasMeter().ConsumeGas(costAddCoins, "addCoins")
	oldCoins := keeper.GetCoins(ctx, addr)
	newCoins := oldCoins.Plus(amt)
	if !newCoins.IsNotNegative() {
		return amt, sdk.ErrInsufficientCoins(fmt.Sprintf("%s < %s", oldCoins, amt))
	}
	err := keeper.SetCoins(ctx, addr, newCoins)
	return newCoins, err
}

// SendCoins moves coins from one account to another
// NOTE: Make sure to revert state changes from tx on error
func (keeper Keeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	_, err := keeper.SubtractCoins(ctx, fromAddr, amt)
	if err != nil {
		return err
	}

	_, err = keeper.AddCoins(ctx, toAddr, amt)
	if err != nil {
		return err
	}

	return nil
}

// InputOutputCoins handles a list of inputs and outputs
// NOTE: Make sure to revert state changes from tx on error
func (keeper Keeper) InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) sdk.Error {
	for _, in := range inputs {
		_, err := keeper.SubtractCoins(ctx, in.Address, in.Coins)
		if err != nil {
			return err
		}
	}

	for _, out := range outputs {
		_, err := keeper.AddCoins(ctx, out.Address, out.Coins)
		if err != nil {
			return err
		}
	}

	return nil
}

// Returns the metadata for a specific denom
func (keeper Keeper) GetDenomMetadata(ctx sdk.Context, denomName string) (DenomMetadata, sdk.Error) {
	store := ctx.KVStore(keeper.metadataStoreKey)
	bz := store.Get([]byte(denomName))
	var denomMetadata DenomMetadata
	keeper.cdc.MustUnmarshalBinary(bz, denomMetadata)
	return denomMetadata, nil
}

// Sets the metadata for a specific denom
func (keeper Keeper) SetDenomMetadata(ctx sdk.Context, denomName string, denomMetadata DenomMetadata) sdk.Error {
	store := ctx.KVStore(keeper.metadataStoreKey)
	bz := keeper.cdc.MustMarshalBinary(denomMetadata)
	store.Set([]byte(denomName), bz)
	return nil
}

// Increases the total supply of a specific denom
func (keeper Keeper) MintCoins(ctx sdk.Context, amt sdk.Coins) sdk.Error {
	for _, coin := range amt {
		metadata, err := keeper.GetDenomMetadata(ctx, coin.Denom)
		if err != nil {
			return err
		}
		metadata.TotalSupply = metadata.TotalSupply.Add(coin.Amount)
		err = keeper.SetDenomMetadata(ctx, coin.Denom, metadata)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decreases the total supply of a specific denom
func (keeper Keeper) BurnCoins(ctx sdk.Context, amt sdk.Coins) sdk.Error {
	for _, coin := range amt {
		metadata, err := keeper.GetDenomMetadata(ctx, coin.Denom)
		if err != nil {
			return err
		}
		metadata.TotalSupply = metadata.TotalSupply.Sub(coin.Amount)
		err = keeper.SetDenomMetadata(ctx, coin.Denom, metadata)
		if err != nil {
			return err
		}
	}
	return nil
}

// Adds coins to an account and increases the total supply of a specific denom
func (keeper Keeper) MintAddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	err := keeper.MintCoins(ctx, amt)
	if err != nil {
		return err
	}
	_, err = keeper.AddCoins(ctx, addr, amt)
	if err != nil {
		return err
	}
	return nil
}

// Subtracts coins to an account and decreases the total supply of a specific denom
func (keeper Keeper) BurnSubCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	err := keeper.BurnCoins(ctx, amt)
	if err != nil {
		return err
	}
	_, err = keeper.SubtractCoins(ctx, addr, amt)
	if err != nil {
		return err
	}
	return nil
}

//______________________________________________________________________________________________

// SendKeeper only allows transfers between accounts, without the possibility of creating coins
type SendKeeper struct {
	keeper Keeper
}

// NewSendKeeper returns a new Keeper
func NewSendKeeper(keeper Keeper) SendKeeper {
	return SendKeeper{keeper: keeper}
}

// Exposes function in main Keeper
func (sendKeeper SendKeeper) GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return sendKeeper.keeper.GetCoins(ctx, addr)
}

// Exposes function in main Keeper
func (sendKeeper SendKeeper) SetCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return sendKeeper.keeper.SetCoins(ctx, addr, amt)
}

// Exposes function in main Keeper
func (sendKeeper SendKeeper) HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool {
	return sendKeeper.keeper.HasCoins(ctx, addr, amt)
}

// Exposes function in main Keeper
func (sendKeeper SendKeeper) SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error) {
	return sendKeeper.keeper.SubtractCoins(ctx, addr, amt)
}

// Exposes function in main Keeper
func (sendKeeper SendKeeper) AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error) {
	return sendKeeper.keeper.AddCoins(ctx, addr, amt)
}

// Exposes function in main Keeper
func (sendKeeper SendKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return sendKeeper.keeper.SendCoins(ctx, fromAddr, toAddr, amt)
}

// Exposes function in main Keeper
func (sendKeeper SendKeeper) InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) sdk.Error {
	return sendKeeper.keeper.InputOutputCoins(ctx, inputs, outputs)
}

// Exposes function in main Keeper
func (sendKeeper SendKeeper) GetDenomMetadata(ctx sdk.Context, denomName string) (DenomMetadata, sdk.Error) {
	return sendKeeper.keeper.GetDenomMetadata(ctx, denomName)
}

//______________________________________________________________________________________________

// ViewKeeper only allows reading of balances
type ViewKeeper struct {
	keeper Keeper
}

// NewViewKeeper returns a new Keeper
func NewViewKeeper(keeper Keeper) ViewKeeper {
	return ViewKeeper{keeper: keeper}
}

// Exposes function in main Keeper
func (viewKeeper ViewKeeper) GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return viewKeeper.keeper.GetCoins(ctx, addr)
}

// Exposes function in main Keeper
func (viewKeeper ViewKeeper) HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool {
	return viewKeeper.keeper.HasCoins(ctx, addr, amt)
}

// Exposes function in main Keeper
func (viewKeeper ViewKeeper) GetDenomMetadata(ctx sdk.Context, denomName string) (DenomMetadata, sdk.Error) {
	return viewKeeper.keeper.GetDenomMetadata(ctx, denomName)
}
