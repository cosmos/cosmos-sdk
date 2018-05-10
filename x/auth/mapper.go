package auth

import (
	"fmt"
	"reflect"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

// Implements bam.AccountMapper.
// This AccountMapper encodes/decodes accounts using the
// go-amino (binary) encoding/decoding library.
type AccountMapper struct {

	// The (unexposed) key used to access the store from the Context.
	key sdk.StoreKey

	// The prototypical bam.Account concrete type.
	proto Account

	// The wire codec for binary encoding/decoding of accounts.
	cdc *wire.Codec
}

// NewAccountMapper returns a new AccountMapper that
// uses go-amino to (binary) encode and decode concrete bam.Accounts.
// nolint
func NewAccountMapper(cdc *wire.Codec, key sdk.StoreKey, proto Account) AccountMapper {
	return AccountMapper{
		key:   key,
		proto: proto,
		cdc:   cdc,
	}
}

// Creates a new account using the address
func (am AccountMapper) NewAccountWithAddress(ctx sdk.Context, addr bam.Address) Account {
	acc := am.clonePrototype()
	acc.SetAddress(addr)
	return acc
}

// Gets an Account by Address
func (am AccountMapper) GetAccount(ctx sdk.Context, addr bam.Address) Account {
	store := ctx.KVStore(am.key)
	bz := store.Get(addr)
	if bz == nil {
		return nil
	}
	acc := am.decodeAccount(bz)
	return acc
}

// Sets an Account
func (am AccountMapper) SetAccount(ctx sdk.Context, acc Account) {
	addr := acc.GetAddress()
	store := ctx.KVStore(am.key)
	bz := am.encodeAccount(acc)
	store.Set(addr, bz)
}

// Iterates over all accounts and filters by process
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

//----------------------------------------
// misc.

// Creates a new struct (or pointer to struct) from am.proto.
func (am AccountMapper) clonePrototype() Account {
	protoRt := reflect.TypeOf(am.proto)
	if protoRt.Kind() == reflect.Ptr {
		protoCrt := protoRt.Elem()
		if protoCrt.Kind() != reflect.Struct {
			panic("accountMapper requires a struct proto bam.Account, or a pointer to one")
		}
		protoRv := reflect.New(protoCrt)
		clone, ok := protoRv.Interface().(Account)
		if !ok {
			panic(fmt.Sprintf("accountMapper requires a proto bam.Account, but %v doesn't implement bam.Account", protoRt))
		}
		return clone
	}

	protoRv := reflect.New(protoRt).Elem()
	clone, ok := protoRv.Interface().(Account)
	if !ok {
		panic(fmt.Sprintf("accountMapper requires a proto bam.Account, but %v doesn't implement bam.Account", protoRt))
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
