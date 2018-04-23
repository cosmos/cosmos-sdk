package types

import (
	"fmt"
	"reflect"

	wire "github.com/cosmos/cosmos-sdk/wire"
)

var _ AccountMapper = (*accountMapper)(nil)

// Implements AccountMapper.
// This AccountMapper encodes/decodes accounts using the
// go-amino (binary) encoding/decoding library.
type accountMapper struct {

	// The (unexposed) key used to access the store from the Context.
	key StoreKey

	// The prototypical Account concrete type.
	proto Account

	// The wire codec for binary encoding/decoding of accounts.
	cdc *wire.Codec
}

// NewAccountMapper returns a new AccountMapper that
// uses go-amino to (binary) encode and decode concrete Accounts.
// nolint
func NewAccountMapper(cdc *wire.Codec, key StoreKey, proto Account) accountMapper {
	return accountMapper{
		key:   key,
		proto: proto,
		cdc:   cdc,
	}
}

// Implements AccountMapper.
func (am accountMapper) NewAccountWithAddress(ctx Context, addr Address) Account {
	acc := am.clonePrototype()
	acc.SetAddress(addr)
	return acc
}

// Implements AccountMapper.
func (am accountMapper) GetAccount(ctx Context, addr Address) Account {
	store := ctx.KVStore(am.key)
	bz := store.Get(addr)
	if bz == nil {
		return nil
	}
	acc := am.decodeAccount(bz)
	return acc
}

// Implements AccountMapper.
func (am accountMapper) SetAccount(ctx Context, acc Account) {
	addr := acc.GetAddress()
	store := ctx.KVStore(am.key)
	bz := am.encodeAccount(acc)
	store.Set(addr, bz)
}

//----------------------------------------
// misc.

// Creates a new struct (or pointer to struct) from am.proto.
func (am accountMapper) clonePrototype() Account {
	protoRt := reflect.TypeOf(am.proto)
	if protoRt.Kind() == reflect.Ptr {
		protoCrt := protoRt.Elem()
		if protoCrt.Kind() != reflect.Struct {
			panic("accountMapper requires a struct proto Account, or a pointer to one")
		}
		protoRv := reflect.New(protoCrt)
		clone, ok := protoRv.Interface().(Account)
		if !ok {
			panic(fmt.Sprintf("accountMapper requires a proto Account, but %v doesn't implement Account", protoRt))
		}
		return clone
	}

	protoRv := reflect.New(protoRt).Elem()
	clone, ok := protoRv.Interface().(Account)
	if !ok {
		panic(fmt.Sprintf("accountMapper requires a proto Account, but %v doesn't implement Account", protoRt))
	}
	return clone
}

func (am accountMapper) encodeAccount(acc Account) []byte {
	bz, err := am.cdc.MarshalBinaryBare(acc)
	if err != nil {
		panic(err)
	}
	return bz
}

func (am accountMapper) decodeAccount(bz []byte) (acc Account) {
	err := am.cdc.UnmarshalBinaryBare(bz, &acc)
	if err != nil {
		panic(err)
	}
	return
}
