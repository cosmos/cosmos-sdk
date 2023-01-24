package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/errors"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	govKeeper sanction.GovKeeper

	authority string

	unsanctionableAddrs map[string]bool

	msgSanctionTypeURL          string
	msgUnsanctionTypeURL        string
	msgExecLegacyContentTypeURL string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	bankKeeper sanction.BankKeeper,
	govKeeper sanction.GovKeeper,
	authority string,
	unsanctionableAddrs []sdk.AccAddress,
) Keeper {
	rv := Keeper{
		cdc:                         cdc,
		storeKey:                    storeKey,
		govKeeper:                   govKeeper,
		authority:                   authority,
		unsanctionableAddrs:         make(map[string]bool),
		msgSanctionTypeURL:          sdk.MsgTypeURL(&sanction.MsgSanction{}),
		msgUnsanctionTypeURL:        sdk.MsgTypeURL(&sanction.MsgUnsanction{}),
		msgExecLegacyContentTypeURL: sdk.MsgTypeURL(&govv1.MsgExecLegacyContent{}),
	}
	for _, addr := range unsanctionableAddrs {
		// using string(addr) here instead of addr.String() to cut down on the need to bech32 encode things.
		rv.unsanctionableAddrs[string(addr)] = true
	}
	bankKeeper.SetSanctionKeeper(rv)
	return rv
}

// GetAuthority returns this module's authority string.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// IsSanctionedAddr returns true if the provided address is currently sanctioned (either permanently or temporarily).
func (k Keeper) IsSanctionedAddr(ctx sdk.Context, addr sdk.AccAddress) bool {
	if len(addr) == 0 || k.IsAddrThatCannotBeSanctioned(addr) {
		return false
	}
	store := ctx.KVStore(k.storeKey)
	tempEntry := k.getLatestTempEntry(store, addr)
	if IsSanctionBz(tempEntry) {
		return true
	}
	if IsUnsanctionBz(tempEntry) {
		return false
	}
	key := CreateSanctionedAddrKey(addr)
	return store.Has(key)
}

// SanctionAddresses creates permanent sanctioned address entries for each of the provided addresses.
// Also deletes any temporary entries for each address.
func (k Keeper) SanctionAddresses(ctx sdk.Context, addrs ...sdk.AccAddress) error {
	store := ctx.KVStore(k.storeKey)
	val := []byte{SanctionB}
	for _, addr := range addrs {
		if k.IsAddrThatCannotBeSanctioned(addr) {
			return errors.ErrUnsanctionableAddr.Wrap(addr.String())
		}
		key := CreateSanctionedAddrKey(addr)
		store.Set(key, val)
		if err := ctx.EventManager().EmitTypedEvent(sanction.NewEventAddressSanctioned(addr)); err != nil {
			return err
		}
	}
	k.DeleteAddrTempEntries(ctx, addrs...)
	return nil
}

// UnsanctionAddresses deletes any sanctioned address entries for each provided address.
// Also deletes any temporary entries for each address.
func (k Keeper) UnsanctionAddresses(ctx sdk.Context, addrs ...sdk.AccAddress) error {
	store := ctx.KVStore(k.storeKey)
	for _, addr := range addrs {
		key := CreateSanctionedAddrKey(addr)
		store.Delete(key)
		if err := ctx.EventManager().EmitTypedEvent(sanction.NewEventAddressUnsanctioned(addr)); err != nil {
			return err
		}
	}
	k.DeleteAddrTempEntries(ctx, addrs...)
	return nil
}

// AddTemporarySanction adds a temporary sanction with the given gov prop id for each of the provided addresses.
func (k Keeper) AddTemporarySanction(ctx sdk.Context, govPropID uint64, addrs ...sdk.AccAddress) error {
	return k.addTempEntries(ctx, SanctionB, govPropID, addrs)
}

// AddTemporaryUnsanction adds a temporary unsanction with the given gov prop id for each of the provided addresses.
func (k Keeper) AddTemporaryUnsanction(ctx sdk.Context, govPropID uint64, addrs ...sdk.AccAddress) error {
	return k.addTempEntries(ctx, UnsanctionB, govPropID, addrs)
}

// addTempEntries adds a temporary entry with the given value and gov prop id for each address given.
func (k Keeper) addTempEntries(ctx sdk.Context, value byte, govPropID uint64, addrs []sdk.AccAddress) error {
	store := ctx.KVStore(k.storeKey)
	val := []byte{value}
	for _, addr := range addrs {
		if value == SanctionB && k.IsAddrThatCannotBeSanctioned(addr) {
			return errors.ErrUnsanctionableAddr.Wrap(addr.String())
		}
		key := CreateTemporaryKey(addr, govPropID)
		store.Set(key, val)
		indKey := CreateProposalTempIndexKey(govPropID, addr)
		store.Set(indKey, val)
		if err := ctx.EventManager().EmitTypedEvent(NewTempEvent(value, addr)); err != nil {
			return err
		}
	}
	return nil
}

// getLatestTempEntry gets the most recent temporary entry for the given address.
func (k Keeper) getLatestTempEntry(store sdk.KVStore, addr sdk.AccAddress) []byte {
	if len(addr) == 0 {
		return nil
	}
	pre := CreateTemporaryAddrPrefix(addr)
	preStore := prefix.NewStore(store, pre)
	iter := preStore.ReverseIterator(nil, nil)
	defer iter.Close()
	if iter.Valid() {
		return iter.Value()
	}
	return nil
}

// DeleteGovPropTempEntries deletes the temporary entries for the given proposal id.
func (k Keeper) DeleteGovPropTempEntries(ctx sdk.Context, govPropID uint64) {
	var toRemove [][]byte
	k.IterateProposalIndexEntries(ctx, &govPropID, func(cbGovPropID uint64, cbAddr sdk.AccAddress) bool {
		toRemove = append(toRemove,
			CreateTemporaryKey(cbAddr, cbGovPropID),
			CreateProposalTempIndexKey(cbGovPropID, cbAddr),
		)
		return false
	})
	if len(toRemove) > 0 {
		store := ctx.KVStore(k.storeKey)
		for _, key := range toRemove {
			store.Delete(key)
		}
	}
}

// DeleteAddrTempEntries deletes all temporary entries for each given address.
func (k Keeper) DeleteAddrTempEntries(ctx sdk.Context, addrs ...sdk.AccAddress) {
	if len(addrs) == 0 {
		return
	}
	var toRemove [][]byte
	callback := func(cbAddr sdk.AccAddress, cbGovPropId uint64, _ bool) bool {
		toRemove = append(toRemove,
			CreateTemporaryKey(cbAddr, cbGovPropId),
			CreateProposalTempIndexKey(cbGovPropId, cbAddr),
		)
		return false
	}
	for _, addr := range addrs {
		if len(addr) > 0 {
			k.IterateTemporaryEntries(ctx, addr, callback)
		}
	}
	if len(toRemove) > 0 {
		store := ctx.KVStore(k.storeKey)
		for _, key := range toRemove {
			store.Delete(key)
		}
	}
}

// getSanctionedAddressPrefixStore returns a kv store prefixed for sanctioned addresses, and the prefix bytes.
func (k Keeper) getSanctionedAddressPrefixStore(ctx sdk.Context) (sdk.KVStore, []byte) {
	return prefix.NewStore(ctx.KVStore(k.storeKey), SanctionedPrefix), SanctionedPrefix
}

// IterateSanctionedAddresses iterates over all of the permanently sanctioned addresses.
// The callback takes in the sanctioned address and should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateSanctionedAddresses(ctx sdk.Context, cb func(addr sdk.AccAddress) (stop bool)) {
	store, _ := k.getSanctionedAddressPrefixStore(ctx)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		addr, _ := ParseLengthPrefixedBz(iter.Key())
		if cb(addr) {
			break
		}
	}
}

// getTemporaryEntryPrefixStore returns a kv store prefixed for temporary sanction/unsanction entries, and the prefix bytes used.
// If an addr is provided, the store is prefixed for just the given address.
// If addr is empty, it will be prefixed for all temporary entries.
func (k Keeper) getTemporaryEntryPrefixStore(ctx sdk.Context, addr sdk.AccAddress) (sdk.KVStore, []byte) {
	pre := CreateTemporaryAddrPrefix(addr)
	return prefix.NewStore(ctx.KVStore(k.storeKey), pre), pre
}

// IterateTemporaryEntries iterates over each of the temporary entries.
// If an address is provided, only the temporary entries for that address are iterated,
// otherwise all entries are iterated.
// The callback takes in the address in question, the governance proposal associated with it, and whether it's a sanction (true) or unsanction (false).
// The callback should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateTemporaryEntries(ctx sdk.Context, addr sdk.AccAddress, cb func(addr sdk.AccAddress, govPropID uint64, isSanction bool) (stop bool)) {
	store, pre := k.getTemporaryEntryPrefixStore(ctx, addr)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := ConcatBz(pre, iter.Key())
		kAddr, govPropID := ParseTemporaryKey(key)
		isSanction := IsSanctionBz(iter.Value())
		if cb(kAddr, govPropID, isSanction) {
			break
		}
	}
}

// getProposalIndexPrefixStore returns a kv store prefixed for the gov prop -> temporary sanction/unsanction index entries,
// and the prefix bytes used.
// If a gov prop id is provided, the store is prefixed for just that proposal.
// If not provided, it will be prefixed for all temp index entries.
func (k Keeper) getProposalIndexPrefixStore(ctx sdk.Context, govPropID *uint64) (sdk.KVStore, []byte) {
	pre := CreateProposalTempIndexPrefix(govPropID)
	return prefix.NewStore(ctx.KVStore(k.storeKey), pre), pre
}

// IterateProposalIndexEntries iterates over all of the index entries for temp entries.
// The callback takes in the gov prop id and address.
// The callback should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateProposalIndexEntries(ctx sdk.Context, govPropID *uint64, cb func(govPropID uint64, addr sdk.AccAddress) (stop bool)) {
	store, pre := k.getProposalIndexPrefixStore(ctx, govPropID)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := ConcatBz(pre, iter.Key())
		kPropID, kAddr := ParseProposalTempIndexKey(key)
		if cb(kPropID, kAddr) {
			break
		}
	}
}

// IsAddrThatCannotBeSanctioned returns true if the provided address is one of the ones that cannot be sanctioned.
// Returns false if the addr can be sanctioned.
func (k Keeper) IsAddrThatCannotBeSanctioned(addr sdk.AccAddress) bool {
	// Okay. I know this is a clunky name for this function.
	// IsUnsanctionableAddr would be a better name if it weren't WAY too close to IsSanctionedAddr.
	// The latter is the key function of this module, and I wanted to help prevent
	// confusion between this one and that one since they have vastly different purposes.
	return k.unsanctionableAddrs[string(addr)]
}

// GetParams gets the sanction module's params.
// If there isn't anything set in state, the defaults are returned.
func (k Keeper) GetParams(ctx sdk.Context) *sanction.Params {
	rv := sanction.DefaultParams()

	k.IterateParams(ctx, func(name, value string) bool {
		switch name {
		case ParamNameImmediateSanctionMinDeposit:
			rv.ImmediateSanctionMinDeposit = toCoinsOrDefault(value, rv.ImmediateSanctionMinDeposit)
		case ParamNameImmediateUnsanctionMinDeposit:
			rv.ImmediateUnsanctionMinDeposit = toCoinsOrDefault(value, rv.ImmediateUnsanctionMinDeposit)
		default:
			panic(fmt.Errorf("unknown param key: %q", name))
		}
		return false
	})

	return rv
}

// SetParams sets the sanction module's params.
// Providing a nil params will cause all params to be deleted (so that defaults are used).
func (k Keeper) SetParams(ctx sdk.Context, params *sanction.Params) error {
	store := ctx.KVStore(k.storeKey)
	if params == nil {
		k.deleteParam(store, ParamNameImmediateSanctionMinDeposit)
		k.deleteParam(store, ParamNameImmediateUnsanctionMinDeposit)
	} else {
		k.setParam(store, ParamNameImmediateSanctionMinDeposit, params.ImmediateSanctionMinDeposit.String())
		k.setParam(store, ParamNameImmediateUnsanctionMinDeposit, params.ImmediateUnsanctionMinDeposit.String())
	}
	return ctx.EventManager().EmitTypedEvent(&sanction.EventParamsUpdated{})
}

// IterateParams iterates over all params entries.
// The callback takes in the name and value, and should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateParams(ctx sdk.Context, cb func(name, value string) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), ParamsPrefix)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if cb(string(iter.Key()), string(iter.Value())) {
			break
		}
	}
}

// GetImmediateSanctionMinDeposit gets the minimum deposit for a sanction to happen immediately.
func (k Keeper) GetImmediateSanctionMinDeposit(ctx sdk.Context) sdk.Coins {
	return k.getParamAsCoinsOrDefault(
		ctx,
		ParamNameImmediateSanctionMinDeposit,
		sanction.DefaultImmediateSanctionMinDeposit,
	)
}

// GetImmediateUnsanctionMinDeposit gets the minimum deposit for an unsanction to happen immediately.
func (k Keeper) GetImmediateUnsanctionMinDeposit(ctx sdk.Context) sdk.Coins {
	return k.getParamAsCoinsOrDefault(
		ctx,
		ParamNameImmediateUnsanctionMinDeposit,
		sanction.DefaultImmediateUnsanctionMinDeposit,
	)
}

// getParam returns a param value and whether it existed.
func (k Keeper) getParam(store sdk.KVStore, name string) (string, bool) {
	key := CreateParamKey(name)
	if store.Has(key) {
		return string(store.Get(key)), true
	}
	return "", false
}

// setParam sets a param value.
func (k Keeper) setParam(store sdk.KVStore, name, value string) {
	key := CreateParamKey(name)
	val := []byte(value)
	store.Set(key, val)
}

// deleteParam deletes a param value.
func (k Keeper) deleteParam(store sdk.KVStore, name string) {
	key := CreateParamKey(name)
	store.Delete(key)
}

// getParamAsCoinsOrDefault gets a param value and converts it to a coins if possible.
// If the param doesn't exist, the default is returned.
// If the param's value cannot be converted to a Coins, the default is returned.
func (k Keeper) getParamAsCoinsOrDefault(ctx sdk.Context, name string, dflt sdk.Coins) sdk.Coins {
	coins, has := k.getParam(ctx.KVStore(k.storeKey), name)
	if !has {
		return dflt
	}
	return toCoinsOrDefault(coins, dflt)
}

// toCoinsOrDefault converts a string to coins if possible or else returns the provided default.
func toCoinsOrDefault(coins string, dflt sdk.Coins) sdk.Coins {
	rv, err := sdk.ParseCoinsNormalized(coins)
	if err != nil {
		return dflt
	}
	return rv
}

// toAccAddrs converts the provided strings into a slice of sdk.AccAddress.
// If any fail to convert, an error is returned.
func toAccAddrs(addrs []string) ([]sdk.AccAddress, error) {
	var err error
	rv := make([]sdk.AccAddress, len(addrs))
	for i, addr := range addrs {
		rv[i], err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid address[%d]: %w", i, err)
		}
	}
	return rv, nil
}
