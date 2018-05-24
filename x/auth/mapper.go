package auth

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	crypto "github.com/tendermint/go-crypto"
)

// This AccountMapper encodes/decodes accounts using the
// go-amino (binary) encoding/decoding library.
type AccountMapper struct {

	// The (unexposed) key used to access the store from the Context.
	key sdk.StoreKey

	// The prototypical Account concrete type.
	proto Account

	// The wire codec for binary encoding/decoding of accounts.
	cdc *wire.Codec
}

// NewAccountMapper returns a new sdk.AccountMapper that
// uses go-amino to (binary) encode and decode concrete sdk.Accounts.
// nolint
func NewAccountMapper(cdc *wire.Codec, key sdk.StoreKey, proto Account) AccountMapper {
	return AccountMapper{
		key:   key,
		proto: proto,
		cdc:   cdc,
	}
}

// Implaements sdk.AccountMapper.
func (am AccountMapper) NewAccountWithAddress(ctx sdk.Context, addr sdk.Address) Account {
	acc := am.clonePrototype()
	acc.SetAddress(addr)
	return acc
}

// Implements sdk.AccountMapper.
func (am AccountMapper) GetAccount(ctx sdk.Context, addr sdk.Address) Account {
	store := ctx.KVStore(am.key)
	bz := store.Get(addr)
	if bz == nil {
		return nil
	}
	acc := am.decodeAccount(bz)
	return acc
}

// Implements sdk.AccountMapper.
func (am AccountMapper) SetAccount(ctx sdk.Context, acc Account) {
	addr := acc.GetAddress()
	store := ctx.KVStore(am.key)
	bz := am.encodeAccount(acc)
	store.Set(addr, bz)
}

// Implements sdk.AccountMapper.
func (am AccountMapper) IterateAccounts(ctx sdk.Context, process func(Account) (stop bool)) {
	store := ctx.KVStore(am.key)
	iter := store.Iterator(nil, nil)
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
func (am AccountMapper) GetPubKey(ctx sdk.Context, addr sdk.Address) (crypto.PubKey, sdk.Error) {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(addr.String())
	}
	return acc.GetPubKey(), nil
}

// Sets the PubKey of the account at address
func (am AccountMapper) SetPubKey(ctx sdk.Context, addr sdk.Address, newPubKey crypto.PubKey) sdk.Error {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		return sdk.ErrUnknownAddress(addr.String())
	}
	acc.SetPubKey(newPubKey)
	am.SetAccount(ctx, acc)
	return nil
}

// Returns the Sequence of the account at address
func (am AccountMapper) GetSequence(ctx sdk.Context, addr sdk.Address) (int64, sdk.Error) {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		return 0, sdk.ErrUnknownAddress(addr.String())
	}
	return acc.GetSequence(), nil
}

func (am AccountMapper) setSequence(ctx sdk.Context, addr sdk.Address, newSequence int64) sdk.Error {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		return sdk.ErrUnknownAddress(addr.String())
	}
	acc.SetSequence(newSequence)
	am.SetAccount(ctx, acc)
	return nil
}

//----------------------------------------
// misc.

// Creates a new struct (or pointer to struct) from am.proto.
func (am AccountMapper) clonePrototype() Account {
	protoRt := reflect.TypeOf(am.proto)
	if protoRt.Kind() == reflect.Ptr {
		protoCrt := protoRt.Elem()
		if protoCrt.Kind() != reflect.Struct {
			panic("accountMapper requires a struct proto sdk.Account, or a pointer to one")
		}
		protoRv := reflect.New(protoCrt)
		clone, ok := protoRv.Interface().(Account)
		if !ok {
			panic(fmt.Sprintf("accountMapper requires a proto sdk.Account, but %v doesn't implement sdk.Account", protoRt))
		}
		return clone
	}

	protoRv := reflect.New(protoRt).Elem()
	clone, ok := protoRv.Interface().(Account)
	if !ok {
		panic(fmt.Sprintf("accountMapper requires a proto sdk.Account, but %v doesn't implement sdk.Account", protoRt))
	}
	return clone
}

func (am AccountMapper) encodeAccount(acc Account) []byte {
	bz, err := am.cdc.MarshalBinaryBare(acc)
	if err != nil {
		panic(err)
	}
	return bz
}

func (am AccountMapper) decodeAccount(bz []byte) (acc Account) {
	err := am.cdc.UnmarshalBinaryBare(bz, &acc)
	if err != nil {
		panic(err)
	}
	return
}
