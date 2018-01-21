package auth

import (
	"fmt"
	"reflect"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements sdk.AccountStore.
// This AccountStore encodes/decodes accounts using the
// go-wire (binary) encoding/decoding library.
type accountStore struct {

	// The (unexposed) key used to access the store from the Context.
	key sdk.SubstoreKey

	// The prototypical sdk.Account concrete type.
	proto sdk.Account

	// The wire codec for binary encoding/decoding of accounts.
	cdc *wire.Codec
}

// NewAccountStore returns a new sdk.AccountStore that
// uses go-wire to (binary) encode and decode concrete sdk.Accounts.
func NewAccountStore(key sdk.SubstoreKey, proto sdk.Account) accountStore {
	return accountStore{
		key:   key,
		proto: proto,
		cdc:   wire.NewCodec(),
	}
}

// Returns the go-wire codec.  You may need to register interfaces
// and concrete types here, if your app's sdk.Account
// implementation includes interface fields.
// NOTE: It is not secure to expose the codec, so check out
// .Seal().
func (as accountStore) WireCodec() *wire.Codec {
	return as.cdc
}

// Returns a "sealed" accountStore.
// The codec is not accessible from a sealedAccountStore
func (as accountStore) Seal() sealedAccountStore {
	return sealedAccountStore{as}
}

// Implements sdk.AccountStore.
func (as accountStore) NewAccountWithAddress(ctx sdk.Context, addr crypto.Address) sdk.Account {
	acc := as.clonePrototype()
	acc.SetAddress(addr)
	return acc
}

// Implements sdk.AccountStore.
func (as accountStore) GetAccount(ctx sdk.Context, addr crypto.Address) sdk.Account {
	store := ctx.KVStore(as.key)
	bz := store.Get(addr)
	if bz == nil {
		return nil
	}
	acc := as.decodeAccount(bz)
	return acc
}

// Implements sdk.AccountStore.
func (as accountStore) SetAccount(ctx sdk.Context, acc sdk.Account) {
	addr := acc.GetAddress()
	store := ctx.KVStore(as.key)
	bz := as.encodeAccount(acc)
	store.Set(addr, bz)
}

//----------------------------------------
// misc.

// Creates a new struct (or pointer to struct) from as.proto.
func (as accountStore) clonePrototype() sdk.Account {
	protoRt := reflect.TypeOf(as.proto)
	if protoRt.Kind() == reflect.Ptr {
		protoCrt := protoRt.Elem()
		if protoCrt.Kind() != reflect.Struct {
			panic("AccountStore requires a struct proto sdk.Account, or a pointer to one")
		}
		protoRv := reflect.New(protoCrt)
		clone, ok := protoRv.Interface().(sdk.Account)
		if !ok {
			panic(fmt.Sprintf("AccountStore requires a proto sdk.Account, but %v doesn't implement sdk.Account", protoRt))
		}
		return clone
	} else {
		protoRv := reflect.New(protoRt).Elem()
		clone, ok := protoRv.Interface().(sdk.Account)
		if !ok {
			panic(fmt.Sprintf("AccountStore requires a proto sdk.Account, but %v doesn't implement sdk.Account", protoRt))
		}
		return clone
	}
}

func (as accountStore) encodeAccount(acc sdk.Account) []byte {
	bz, err := as.cdc.MarshalBinary(acc)
	if err != nil {
		panic(err)
	}
	return bz
}

func (as accountStore) decodeAccount(bz []byte) sdk.Account {
	acc := as.clonePrototype()
	err := as.cdc.UnmarshalBinary(bz, &acc)
	if err != nil {
		panic(err)
	}
	return acc
}

//----------------------------------------
// sealedAccountStore

type sealedAccountStore struct {
	accountStore
}

// There's no way for external modules to mutate the
// sas.accountStore.ctx from here, even with reflection.
func (sas sealedAccountStore) WireCodec() *wire.Codec {
	panic("accountStore is sealed")
}
