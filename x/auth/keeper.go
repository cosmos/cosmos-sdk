package auth

import (
	"sort"
	"sync"

	"github.com/hashicorp/golang-lru"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// Prefix for account-by-address store
	AddressStoreKeyPrefix = []byte{0x01}

	globalAccountNumberKey = []byte("globalAccountNumber")
)

// This AccountKeeper encodes/decodes accounts using the
// go-amino (binary) encoding/decoding library.
type AccountKeeper struct {

	// The (unexposed) key used to access the store from the Context.
	key sdk.StoreKey

	// The prototypical Account constructor.
	proto func() sdk.Account

	// The codec codec for binary encoding/decoding of accounts.
	cdc *codec.Codec
}

// NewAccountKeeper returns a new sdk.AccountKeeper that
// uses go-amino to (binary) encode and decode concrete sdk.Accounts.
// nolint
func NewAccountKeeper(cdc *codec.Codec, key sdk.StoreKey, proto func() sdk.Account) AccountKeeper {
	return AccountKeeper{
		key:   key,
		proto: proto,
		cdc:   cdc,
	}
}

// Implaements sdk.AccountKeeper.
func (am AccountKeeper) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.Account {
	acc := am.proto()
	err := acc.SetAddress(addr)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	err = acc.SetAccountNumber(am.GetNextAccountNumber(ctx))
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	return acc
}

// New Account
func (am AccountKeeper) NewAccount(ctx sdk.Context, acc sdk.Account) sdk.Account {
	err := acc.SetAccountNumber(am.GetNextAccountNumber(ctx))
	if err != nil {
		// TODO: Handle with #870
		panic(err)
	}
	return acc
}

// Turn an address to key used to get it from the account store
func AddressStoreKey(addr sdk.AccAddress) []byte {
	return append(AddressStoreKeyPrefix, addr.Bytes()...)
}

// Implements sdk.AccountKeeper.
func (am AccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.Account {
	cache := ctx.AccountCache()

	cVal := cache.GetAccount(addr)
	if acc, ok := cVal.(sdk.Account); ok {
		return acc
	}
	return nil
}

// Implements sdk.AccountKeeper.
func (am AccountKeeper) SetAccount(ctx sdk.Context, acc sdk.Account) {
	addr := acc.GetAddress()
	cache := ctx.AccountCache()
	cache.SetAccount(addr, acc)
}

// RemoveAccount removes an account for the account mapper store.
func (am AccountKeeper) RemoveAccount(ctx sdk.Context, acc sdk.Account) {
	addr := acc.GetAddress()
	cache := ctx.AccountCache()
	cache.Delete(addr)
}

// Implements sdk.AccountKeeper.
func (am AccountKeeper) IterateAccounts(ctx sdk.Context, process func(sdk.Account) (stop bool)) {
	store := ctx.KVStore(am.key)
	iter := sdk.KVStorePrefixIterator(store, AddressStoreKeyPrefix)
	defer iter.Close()
	for {
		if !iter.Valid() {
			return
		}
		val := iter.Value()
		acc := am.decodeAccount(val)
		if process(acc) {
			return
		}
		iter.Next()
	}
}

// Returns the PubKey of the account at address
func (am AccountKeeper) GetPubKey(ctx sdk.Context, addr sdk.AccAddress) (crypto.PubKey, sdk.Error) {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(addr.String())
	}
	return acc.GetPubKey(), nil
}

// Returns the Sequence of the account at address
func (am AccountKeeper) GetSequence(ctx sdk.Context, addr sdk.AccAddress) (uint64, sdk.Error) {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		return 0, sdk.ErrUnknownAddress(addr.String())
	}
	return acc.GetSequence(), nil
}

func (am AccountKeeper) setSequence(ctx sdk.Context, addr sdk.AccAddress, newSequence uint64) sdk.Error {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		return sdk.ErrUnknownAddress(addr.String())
	}
	err := acc.SetSequence(newSequence)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	am.SetAccount(ctx, acc)
	return nil
}

// Returns and increments the global account number counter
func (am AccountKeeper) GetNextAccountNumber(ctx sdk.Context) uint64 {
	var accNumber uint64
	store := ctx.KVStore(am.key)
	bz := store.Get(globalAccountNumberKey)
	if bz == nil {
		accNumber = 0
	} else {
		err := am.cdc.UnmarshalBinaryLengthPrefixed(bz, &accNumber)
		if err != nil {
			panic(err)
		}
	}

	bz = am.cdc.MustMarshalBinaryLengthPrefixed(accNumber + 1)
	store.Set(globalAccountNumberKey, bz)

	return accNumber
}

//----------------------------------------
// misc.

func (am AccountKeeper) encodeAccount(acc sdk.Account) []byte {
	bz, err := am.cdc.MarshalBinaryBare(acc)
	if err != nil {
		panic(err)
	}
	return bz
}

func (am AccountKeeper) decodeAccount(bz []byte) (acc sdk.Account) {
	err := am.cdc.UnmarshalBinaryBare(bz, &acc)
	if err != nil {
		panic(err)
	}
	return
}

func NewAccountStoreCache(cdc *codec.Codec, store sdk.KVStore, cap int) sdk.AccountStoreCache {
	cache, err := lru.New(cap)
	if err != nil {
		panic(err)
	}

	return &accountStoreCache{
		cdc:   cdc,
		cache: cache,
		store: store,
	}
}

type accountStoreCache struct {
	cdc   *codec.Codec
	cache *lru.Cache
	store sdk.KVStore
}

func (ac *accountStoreCache) getAccountFromCache(addr sdk.AccAddress) (acc sdk.Account, ok bool) {
	cacc, ok := ac.cache.Get(string(addr))
	if !ok {
		return nil, ok
	}
	if acc, ok := cacc.(sdk.Account); ok {
		return acc.Clone(), ok
	}
	return nil, false
}

func (ac *accountStoreCache) setAccountToCache(addr sdk.AccAddress, acc sdk.Account) {
	ac.cache.Add(string(addr), acc.Clone())
}

func (ac *accountStoreCache) GetAccount(addr sdk.AccAddress) sdk.Account {
	if acc, ok := ac.getAccountFromCache(addr); ok {
		return acc
	}

	bz := ac.store.Get(AddressStoreKey(addr))
	if bz == nil {
		return nil
	}
	acc := ac.decodeAccount(bz)
	ac.setAccountToCache(addr, acc)
	return acc
}

func (ac *accountStoreCache) SetAccount(addr sdk.AccAddress, acc sdk.Account) {
	cacc, ok := acc.(sdk.Account)
	if !ok {
		return
	}

	bz := ac.encodeAccount(cacc)
	ac.setAccountToCache(addr, cacc)
	ac.store.Set(AddressStoreKey(addr), bz)
}

func (ac *accountStoreCache) Delete(addr sdk.AccAddress) {
	ac.setAccountToCache(addr, nil)
	ac.store.Delete(AddressStoreKey(addr))
}

func (ac *accountStoreCache) encodeAccount(acc sdk.Account) []byte {
	bz, err := ac.cdc.MarshalBinaryBare(acc)
	if err != nil {
		panic(err)
	}
	return bz
}

func (ac *accountStoreCache) decodeAccount(bz []byte) (acc sdk.Account) {
	err := ac.cdc.UnmarshalBinaryBare(bz, &acc)
	if err != nil {
		panic(err)
	}
	return
}

type cValue struct {
	acc     sdk.Account
	deleted bool
	dirty   bool
}

func NewAccountCache(parent sdk.AccountStoreCache) sdk.AccountCache {
	return &accountCache{
		parent: parent,
	}
}

type accountCache struct {
	cache  sync.Map
	parent sdk.AccountStoreCache
}

func (ac *accountCache) GetAccount(addr sdk.AccAddress) sdk.Account {
	return ac.getAccountFromCache(addr)
}

func (ac *accountCache) SetAccount(addr sdk.AccAddress, acc sdk.Account) {
	ac.setAccountToCache(addr, acc, false, true)
}

func (ac *accountCache) Delete(addr sdk.AccAddress) {
	ac.setAccountToCache(addr, nil, true, true)
}

func (ac *accountCache) Cache() sdk.AccountCache {
	return &accountCache{
		parent: ac,
	}
}

func (ac *accountCache) Write() {
	// We need a copy of all of the keys.
	// Not the best, but probably not a bottleneck depending.
	// And there is a new problem, we can not get length of map,
	// so we can not prepare enough space for keys
	keys := make([]string, 0)
	ac.cache.Range(func(key, value interface{}) bool {
		dbValue := value.(cValue)
		if dbValue.dirty {
			keys = append(keys, key.(string))
		}
		return true
	})

	sort.Strings(keys)

	// TODO: Consider allowing usage of Batch, which would allow the write to
	// at least happen atomically.
	for _, key := range keys {
		// value should exist here, so does not check ok
		value, _ := ac.cache.Load(key)
		cacheValue := value.(cValue)

		if cacheValue.deleted {
			ac.parent.Delete(sdk.AccAddress(key))
		} else if cacheValue.acc == nil {
			// Skip, it already doesn't exist in parent.
		} else {
			ac.parent.SetAccount(sdk.AccAddress(key), cacheValue.acc)
		}
	}

	// clear the cache
	ac.cache = sync.Map{}
}

func (ac *accountCache) getAccountFromCache(addr sdk.AccAddress) (acc sdk.Account) {
	cacheVal, ok := ac.cache.Load(string(addr))
	if !ok {
		acc = ac.parent.GetAccount(addr)
		ac.setAccountToCache(addr, acc, false, false)
	} else {
		acc = cacheVal.(cValue).acc
	}

	if acc == nil {
		return nil
	}
	return acc.Clone()
}

func (ac *accountCache) setAccountToCache(addr sdk.AccAddress, acc sdk.Account, deleted bool, dirty bool) {
	if acc != nil {
		acc = acc.Clone()
	}

	ac.cache.Store(string(addr), cValue{
		acc:     acc,
		deleted: deleted,
		dirty:   dirty,
	})
}
