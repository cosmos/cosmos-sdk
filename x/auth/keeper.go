package auth

import (
	"github.com/tendermint/tendermint/crypto"

	codec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

var (
	// Prefix for account-by-address store
	AddressStoreKeyPrefix = []byte{0x01}

	globalAccountNumberKey = []byte("globalAccountNumber")

	// StoreKey is string representation of the store key for auth
	StoreKey = "acc"

	// FeeStoreKey is a string representation of the store key for fees
	FeeStoreKey = "fee"
)

// This AccountKeeper encodes/decodes accounts using the go-amino (binary)
// encoding/decoding library.
type AccountKeeper struct {
	// The (unexposed) key used to access the store from the Context.
	key sdk.StoreKey

	// The prototypical Account constructor.
	proto func() Account

	// The codec codec for binary encoding/decoding of accounts.
	cdc *codec.Codec

	paramSubspace params.Subspace
}

// NewAccountKeeper returns a new sdk.AccountKeeper that uses go-amino to
// (binary) encode and decode concrete sdk.Accounts.
// nolint
func NewAccountKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramstore params.Subspace, proto func() Account,
) AccountKeeper {

	return AccountKeeper{
		key:           key,
		proto:         proto,
		cdc:           cdc,
		paramSubspace: paramstore.WithTypeTable(ParamTypeTable()),
	}
}

// Implaements sdk.AccountKeeper.
func (ak AccountKeeper) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) Account {
	acc := ak.proto()
	err := acc.SetAddress(addr)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	err = acc.SetAccountNumber(ak.GetNextAccountNumber(ctx))
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	return acc
}

// New Account
func (ak AccountKeeper) NewAccount(ctx sdk.Context, acc Account) Account {
	err := acc.SetAccountNumber(ak.GetNextAccountNumber(ctx))
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
func (ak AccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) Account {
	store := ctx.KVStore(ak.key)
	bz := store.Get(AddressStoreKey(addr))
	if bz == nil {
		return nil
	}
	acc := ak.decodeAccount(bz)
	return acc
}

// Implements sdk.AccountKeeper.
func (ak AccountKeeper) SetAccount(ctx sdk.Context, acc Account) {
	addr := acc.GetAddress()
	store := ctx.KVStore(ak.key)
	bz := ak.encodeAccount(acc)
	store.Set(AddressStoreKey(addr), bz)
}

// RemoveAccount removes an account for the account mapper store.
func (ak AccountKeeper) RemoveAccount(ctx sdk.Context, acc Account) {
	addr := acc.GetAddress()
	store := ctx.KVStore(ak.key)
	store.Delete(AddressStoreKey(addr))
}

// Implements sdk.AccountKeeper.
func (ak AccountKeeper) IterateAccounts(ctx sdk.Context, process func(Account) (stop bool)) {
	store := ctx.KVStore(ak.key)
	iter := sdk.KVStorePrefixIterator(store, AddressStoreKeyPrefix)
	defer iter.Close()
	for {
		if !iter.Valid() {
			return
		}
		val := iter.Value()
		acc := ak.decodeAccount(val)
		if process(acc) {
			return
		}
		iter.Next()
	}
}

// Returns the PubKey of the account at address
func (ak AccountKeeper) GetPubKey(ctx sdk.Context, addr sdk.AccAddress) (crypto.PubKey, sdk.Error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(addr.String())
	}
	return acc.GetPubKey(), nil
}

// Returns the Sequence of the account at address
func (ak AccountKeeper) GetSequence(ctx sdk.Context, addr sdk.AccAddress) (uint64, sdk.Error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return 0, sdk.ErrUnknownAddress(addr.String())
	}
	return acc.GetSequence(), nil
}

func (ak AccountKeeper) setSequence(ctx sdk.Context, addr sdk.AccAddress, newSequence uint64) sdk.Error {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return sdk.ErrUnknownAddress(addr.String())
	}
	err := acc.SetSequence(newSequence)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	ak.SetAccount(ctx, acc)
	return nil
}

// Returns and increments the global account number counter
func (ak AccountKeeper) GetNextAccountNumber(ctx sdk.Context) uint64 {
	var accNumber uint64
	store := ctx.KVStore(ak.key)
	bz := store.Get(globalAccountNumberKey)
	if bz == nil {
		accNumber = 0
	} else {
		err := ak.cdc.UnmarshalBinaryLengthPrefixed(bz, &accNumber)
		if err != nil {
			panic(err)
		}
	}

	bz = ak.cdc.MustMarshalBinaryLengthPrefixed(accNumber + 1)
	store.Set(globalAccountNumberKey, bz)

	return accNumber
}

//-----------------------------------------------------------------------------
// Params

// SetParams sets the auth module's parameters.
func (ak AccountKeeper) SetParams(ctx sdk.Context, params Params) {
	ak.paramSubspace.SetParamSet(ctx, &params)
}

// GetParams gets the auth module's parameters.
func (ak AccountKeeper) GetParams(ctx sdk.Context) (params Params) {
	ak.paramSubspace.GetParamSet(ctx, &params)
	return
}

//-----------------------------------------------------------------------------
// Misc.

func (ak AccountKeeper) encodeAccount(acc Account) []byte {
	bz, err := ak.cdc.MarshalBinaryBare(acc)
	if err != nil {
		panic(err)
	}
	return bz
}

func (ak AccountKeeper) decodeAccount(bz []byte) (acc Account) {
	err := ak.cdc.UnmarshalBinaryBare(bz, &acc)
	if err != nil {
		panic(err)
	}
	return
}
