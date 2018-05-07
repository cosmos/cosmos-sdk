package auth

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

var _ sdk.AccountMapper = (*accountMapper)(nil)

// Implements sdk.AccountMapper.
// This AccountMapper encodes/decodes accounts using the
// go-amino (binary) encoding/decoding library.
type accountMapper struct {

	// The (unexposed) key used to access the store from the Context.
	key sdk.StoreKey

	// The prototypical sdk.Account concrete type.
	proto sdk.Account

	// The wire codec for binary encoding/decoding of accounts.
	cdc *wire.Codec
}

// NewAccountMapper returns a new sdk.AccountMapper that
// uses go-amino to (binary) encode and decode concrete sdk.Accounts.
// nolint
func NewAccountMapper(cdc *wire.Codec, key sdk.StoreKey, proto sdk.Account) accountMapper {
	return accountMapper{
		key:   key,
		proto: proto,
		cdc:   cdc,
	}
}

// Implements sdk.AccountMapper.
func (am accountMapper) NewAccountWithAddress(ctx sdk.Context, addr sdk.Address) sdk.Account {
	acc := am.clonePrototype()
	acc.SetAddress(addr)
	return acc
}

// Implements sdk.AccountMapper.
func (am accountMapper) GetAccount(ctx sdk.Context, addr sdk.Address) sdk.Account {
	store := ctx.KVStore(am.key)
	bz := store.Get(addr)
	if bz == nil {
		return nil
	}
	acc := am.decodeAccount(bz)
	return acc
}

// Implements sdk.AccountMapper.
func (am accountMapper) SetAccount(ctx sdk.Context, acc sdk.Account) {
	addr := acc.GetAddress()
	store := ctx.KVStore(am.key)
	bz := am.encodeAccount(acc)
	store.Set(addr, bz)
}

// Implements sdk.AccountMapper.
func (am accountMapper) IterateAccounts(ctx sdk.Context, process func(sdk.Account) (stop bool)) {
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

//----------------------------------------
// misc.

// Creates a new struct (or pointer to struct) from am.proto.
func (am accountMapper) clonePrototype() sdk.Account {
	protoRt := reflect.TypeOf(am.proto)
	if protoRt.Kind() == reflect.Ptr {
		protoCrt := protoRt.Elem()
		if protoCrt.Kind() != reflect.Struct {
			panic("accountMapper requires a struct proto sdk.Account, or a pointer to one")
		}
		protoRv := reflect.New(protoCrt)
		clone, ok := protoRv.Interface().(sdk.Account)
		if !ok {
			panic(fmt.Sprintf("accountMapper requires a proto sdk.Account, but %v doesn't implement sdk.Account", protoRt))
		}
		return clone
	}

	protoRv := reflect.New(protoRt).Elem()
	clone, ok := protoRv.Interface().(sdk.Account)
	if !ok {
		panic(fmt.Sprintf("accountMapper requires a proto sdk.Account, but %v doesn't implement sdk.Account", protoRt))
	}
	return clone
}

func (am accountMapper) encodeAccount(acc sdk.Account) []byte {
	bz, err := am.cdc.MarshalBinaryBare(acc)
	if err != nil {
		panic(err)
	}
	return bz
}

func (am accountMapper) decodeAccount(bz []byte) (acc sdk.Account) {
	err := am.cdc.UnmarshalBinaryBare(bz, &acc)
	if err != nil {
		panic(err)
	}
	return
}
