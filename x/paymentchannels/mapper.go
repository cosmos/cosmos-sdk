package paymentchannels

import (
	"fmt"
	"reflect"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// This PaymentChannelMapper encodes/decodes paymentchannels using the
// go-wire (binary) encoding/decoding library.
type paymentChannelMapper struct {

	// The (unexposed) key used to access the store from the Context.
	key sdk.StoreKey

	// The prototypical sdk.Account concrete type.
	proto PaymentChannel

	// The wire codec for binary encoding/decoding of accounts.
	cdc *wire.Codec
}

// NewPaymentChannelMapper returns a new PaymentChannelMapper that
// uses go-wire to (binary) encode and decode PaymentChannels.
func NewPaymentChannelMapper(key sdk.StoreKey, proto PaymentChannel) accountMapper {
	cdc := wire.NewCodec()
	return paymentChannelMapper{
		key:   key,
		proto: proto,
		cdc:   cdc,
	}
}

// Returns the go-wire codec.
func (pcm paymentChannelMapper) WireCodec() *wire.Codec {
	return pcm.cdc
}

// Returns a "sealed" accountMapper.
// The codec is not accessible from a sealedAccountMapper.
func (pcm paymentChannelMapper) Seal() sealedPaymentChannelMapper {
	return sealedPaymentChannelMapper{pcm}
}

// Implements sdk.AccountMapper.
func (pcm paymentChannelMapper) NewPaymentChannel(ctx sdk.Context, addr crypto.Address) sdk.Account {
	acc := am.clonePrototype()
	acc.SetAddress(addr)
	return acc
}

// Implements sdk.AccountMapper.
func (am accountMapper) GetAccount(ctx sdk.Context, addr crypto.Address) sdk.Account {
	store := ctx.KVStore(am.key)
	bz := store.Get(addr)
	if bz == nil {
		return nil
	}
	acc := am.decodeAccount(bz)
	return acc
}

// Implements sdk.AccountMapper.
func (am accountMapper) SetPaymentChannel(ctx sdk.Context, channel PaymentChannel) {
	store := ctx.KVStore(pcm.key)

	id := channel.GetAddress()
	bz := am.encodePaymentChannel(acc)
	store.Set(addr, bz)
}

//----------------------------------------
// misc.

func (am accountMapper) clonePrototypePtr() interface{} {
	protoRt := reflect.TypeOf(am.proto)
	if protoRt.Kind() == reflect.Ptr {
		protoErt := protoRt.Elem()
		if protoErt.Kind() != reflect.Struct {
			panic("accountMapper requires a struct proto sdk.Account, or a pointer to one")
		}
		protoRv := reflect.New(protoErt)
		return protoRv.Interface()
	}

	protoRv := reflect.New(protoRt)
	return protoRv.Interface()
}

// Creates a new struct (or pointer to struct) from pcm.proto.
func (pcm paymentChannelMapper) clonePrototype() PaymentChannel {
	protoRt := reflect.TypeOf(pcm.proto)
	if protoRt.Kind() == reflect.Ptr {
		protoCrt := protoRt.Elem()
		if protoCrt.Kind() != reflect.Struct {
			panic("paymentChannelMapper requires a struct proto PaymentChannel, or a pointer to one")
		}
		protoRv := reflect.New(protoCrt)
		clone, ok := protoRv.Interface().(PaymentChannel)
		if !ok {
			panic(fmt.Sprintf("paymentChannelMapper requires a proto PaymentChannel, but %v doesn't implement PaymentChannel", protoRt))
		}
		return clone
	}
	protoRv := reflect.New(protoRt).Elem()
	clone, ok := protoRv.Interface().(sdk.Account)
	if !ok {
		panic(fmt.Sprintf("paymentChannelMapper requires a proto PaymentChannel, but %v doesn't implement PaymentChannel", protoRt))
	}
	return clone
}

func (pcm paymentChannelMapper) encodePaymentChannel(channel PaymentChannel) []byte {
	bz, err := pcm.cdc.MarshalBinary(channel)
	if err != nil {
		panic(err)
	}
	return bz
}

func (pcm paymentChannelMapper) decodePaymentChannel(bz []byte) PaymentChannel {
	channelPtr := pcm.clonePrototypePtr()
	err := pcm.cdc.UnmarshalBinary(bz, channelPtr)
	if err != nil {
		panic(err)
	}
	if reflect.ValueOf(pcm.proto).Kind() == reflect.Ptr {
		return reflect.ValueOf(channelPtr).Interface().(PaymentChannel)
	}

	return reflect.ValueOf(channelPtr).Elem().Interface().(PaymentChannel)
}

//----------------------------------------
// sealedAccountMapper

type sealedPaymentChannelMapper struct {
	paymentChannelMapper
}

// There's no way for external modules to mutate the
// sam.accountMapper.ctx from here, even with reflection.
func (sam sealedPaymentChannelMapper) WireCodec() *wire.Codec {
	panic("paymentChannelMapper is sealed")
}
