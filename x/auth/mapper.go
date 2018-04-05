package auth

import (
	"bytes"
	"fmt"
	"reflect"

	oldwire "github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

var _ sdk.AccountMapper = (*accountMapper)(nil)
var _ sdk.AccountMapper = (*sealedAccountMapper)(nil)

// Implements sdk.AccountMapper.
// This AccountMapper encodes/decodes accounts using the
// go-wire (binary) encoding/decoding library.
type accountMapper struct {

	// The (unexposed) key used to access the store from the Context.
	key sdk.StoreKey

	// The prototypical sdk.Account concrete type.
	proto sdk.Account

	// The wire codec for binary encoding/decoding of accounts.
	cdc *wire.Codec
}

// NewAccountMapper returns a new sdk.AccountMapper that
// uses go-wire to (binary) encode and decode concrete sdk.Accounts.
func NewAccountMapper(key sdk.StoreKey, proto sdk.Account) accountMapper {
	cdc := wire.NewCodec()
	return accountMapper{
		key:   key,
		proto: proto,
		cdc:   cdc,
	}
}

// Create and return a sealed account mapper
func NewAccountMapperSealed(key sdk.StoreKey, proto sdk.Account) sealedAccountMapper {
	cdc := wire.NewCodec()
	am := accountMapper{
		key:   key,
		proto: proto,
		cdc:   cdc,
	}
	RegisterWireBaseAccount(cdc)

	// make accountMapper's WireCodec() inaccessible, return
	return am.Seal()
}

// Returns the go-wire codec.  You may need to register interfaces
// and concrete types here, if your app's sdk.Account
// implementation includes interface fields.
// NOTE: It is not secure to expose the codec, so check out
// .Seal().
func (am accountMapper) WireCodec() *wire.Codec {
	return am.cdc
}

// Returns a "sealed" accountMapper.
// The codec is not accessible from a sealedAccountMapper.
func (am accountMapper) Seal() sealedAccountMapper {
	return sealedAccountMapper{am}
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

//----------------------------------------
// sealedAccountMapper

type sealedAccountMapper struct {
	accountMapper
}

// There's no way for external modules to mutate the
// sam.accountMapper.ctx from here, even with reflection.
func (sam sealedAccountMapper) WireCodec() *wire.Codec {
	panic("accountMapper is sealed")
}

//----------------------------------------
// misc.

// NOTE: currently unused
func (am accountMapper) clonePrototypePtr() interface{} {
	protoRt := reflect.TypeOf(am.proto)
	if protoRt.Kind() == reflect.Ptr {
		protoErt := protoRt.Elem()
		if protoErt.Kind() != reflect.Struct {
			panic("accountMapper requires a struct proto sdk.Account, or a pointer to one")
		}
		protoRv := reflect.New(protoErt)
		return protoRv.Interface()
	} else {
		protoRv := reflect.New(protoRt)
		return protoRv.Interface()
	}
}

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
	} else {
		protoRv := reflect.New(protoRt).Elem()
		clone, ok := protoRv.Interface().(sdk.Account)
		if !ok {
			panic(fmt.Sprintf("accountMapper requires a proto sdk.Account, but %v doesn't implement sdk.Account", protoRt))
		}
		return clone
	}
}

func (am accountMapper) encodeAccount(acc sdk.Account) []byte {
	bz, err := am.cdc.MarshalBinary(acc)
	if err != nil {
		panic(err)
	}
	return bz
}

func (am accountMapper) decodeAccount(bz []byte) sdk.Account {
	// ... old go-wire ...
	r, n, err := bytes.NewBuffer(bz), new(int), new(error)
	accI := oldwire.ReadBinary(struct{ sdk.Account }{}, r, len(bz), n, err)
	if *err != nil {
		panic(*err)
	}

	acc := accI.(struct{ sdk.Account }).Account
	return acc

	/*
		accPtr := am.clonePrototypePtr()
			err := am.cdc.UnmarshalBinary(bz, accPtr)
			if err != nil {
				panic(err)
			}
			if reflect.ValueOf(am.proto).Kind() == reflect.Ptr {
				return reflect.ValueOf(accPtr).Interface().(sdk.Account)
			} else {
				return reflect.ValueOf(accPtr).Elem().Interface().(sdk.Account)
			}
	*/
}
