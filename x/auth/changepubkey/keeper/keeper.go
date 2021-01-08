package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

// Keeper manages history of public keys per account
type Keeper struct {
	key sdk.StoreKey
	cdc codec.BinaryMarshaler
	ak  authkeeper.AccountKeeper
}

// NewKeeper returns a new keeper which manages pubkey history per account.
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, ak authkeeper.AccountKeeper,
) Keeper {

	return Keeper{key, cdc, ak}
}

// "Everytime a key for an address is changed, we will store a log of this change in the state of the chain,
// thus creating a stack of all previous keys for an address and the time intervals for which they were active.
// This allows dapps and clients to easily query past keys for an account which may be useful for features
// such as verifying timestamped off-chain signed messages."

// Logger returns a module-specific logger.
func (pk Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetPubKeyHistory Returns the PubKey history of the account at address by time: involves current pubkey
func (pk Keeper) GetPubKeyHistory(ctx sdk.Context, addr sdk.AccAddress) ([]types.PubKeyHistory, error) {
	entries := []types.PubKeyHistory{}
	if pk.ak.GetAccount(ctx, addr) == nil {
		return entries, fmt.Errorf("account %s not found", addr.String())
	}
	iterator := pk.PubKeyHistoryIterator(ctx, addr)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		entry := types.DecodeHistoricalEntry(iterator.Value())
		entries = append(entries, entry)
	}
	currentEntry := pk.GetCurrentPubKeyEntry(ctx, addr)
	entries = append(entries, currentEntry)
	return entries, nil
}

// GetPubKeyHistoricalEntry Returns the PubKey historical entry at a specific time: involves current pubkey
func (pk Keeper) GetPubKeyHistoricalEntry(ctx sdk.Context, addr sdk.AccAddress, time time.Time) types.PubKeyHistory {
	iterator := pk.PubKeyHistoryIteratorAfter(ctx, addr, time)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		entry := types.DecodeHistoricalEntry(iterator.Value())
		if entry.EndTime.After(time) || entry.EndTime.Equal(time) { // TODO: is this inclusive?
			return entry
		}
	}

	return pk.GetCurrentPubKeyEntry(ctx, addr)
}

// GetLastPubKeyHistoricalEntry Returns the PubKey historical entry of last pubkey: does not involve current pubkey
func (pk Keeper) GetLastPubKeyHistoricalEntry(ctx sdk.Context, addr sdk.AccAddress) types.PubKeyHistory {
	iterator := pk.PubKeyHistoryReverseIterator(ctx, addr)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		entry := types.DecodeHistoricalEntry(iterator.Value())
		return entry
	}
	return types.PubKeyHistory{
		StartTime: time.Time{},
		EndTime:   time.Time{},
	}
}

// GetCurrentPubKeyEntry Returns the PubKey entry of current time
func (pk Keeper) GetCurrentPubKeyEntry(ctx sdk.Context, addr sdk.AccAddress) types.PubKeyHistory {
	time := ctx.BlockTime() // TODO: ctx.BlockTime() is correct for endTime?
	acc := pk.ak.GetAccount(ctx, addr)
	lastEntry := pk.GetLastPubKeyHistoricalEntry(ctx, addr)
	st := lastEntry.EndTime
	if st.Equal(time) {
		st = lastEntry.GetStartTime()
	}
	var pubkey crypto.PubKey
	if acc != nil {
		pubkey = acc.GetPubKey()
	}

	return types.PubKeyHistory{
		PubKey:    types.EncodePubKey(pk.cdc, pubkey),
		StartTime: st,
		EndTime:   time,
	}
}

// StoreLastPubKey Store pubkey of an account at the time of changepubkey action
func (pk Keeper) StoreLastPubKey(ctx sdk.Context, addr sdk.AccAddress, time time.Time, pubkey crypto.PubKey) error {
	store := ctx.KVStore(pk.key)
	prefixStore := prefix.NewStore(store, addr) // prefix store for specific account
	key := types.GetPubKeyHistoryKey(time)
	lastEntry := pk.GetLastPubKeyHistoricalEntry(ctx, addr)
	st := lastEntry.EndTime
	if st.Equal(time) {
		st = lastEntry.GetStartTime()
	}
	prefixStore.Set(key, types.EncodeHistoricalEntry(types.PubKeyHistory{
		PubKey:    types.EncodePubKey(pk.cdc, pubkey),
		StartTime: st,
		EndTime:   time,
	}))
	return nil
}

// PubKeyHistoryIteratorAfter returns the iterator used for getting a set of history
// where pubkey endTime is after a specific time
func (pk Keeper) PubKeyHistoryIteratorAfter(ctx sdk.Context, addr sdk.AccAddress, time time.Time) sdk.Iterator {
	store := ctx.KVStore(pk.key)
	prefixStore := prefix.NewStore(store, addr) // prefix store for specific account
	startKey := types.GetPubKeyHistoryKey(time)
	endKey := storetypes.PrefixEndBytes(types.KeyPrefixPubKeyHistory)
	return prefixStore.Iterator(startKey, endKey)
}

// PubKeyHistoryIterator returns the iterator used for getting a full history
func (pk Keeper) PubKeyHistoryIterator(ctx sdk.Context, addr sdk.AccAddress) sdk.Iterator {
	store := ctx.KVStore(pk.key)
	prefixStore := prefix.NewStore(store, addr) // prefix store for specific account
	return sdk.KVStorePrefixIterator(prefixStore, types.KeyPrefixPubKeyHistory)
}

// PubKeyHistoryReverseIterator returns the iterator used for getting a full history in reverse order
func (pk Keeper) PubKeyHistoryReverseIterator(ctx sdk.Context, addr sdk.AccAddress) sdk.Iterator {
	store := ctx.KVStore(pk.key)
	prefixStore := prefix.NewStore(store, addr) // prefix store for specific account
	endKey := storetypes.PrefixEndBytes(types.KeyPrefixPubKeyHistory)
	return prefixStore.ReverseIterator(types.KeyPrefixPubKeyHistory, endKey)
}
