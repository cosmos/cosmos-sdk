package mapping

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

type Value interface {
	Get(Context, interface{})
	GetIfExists(Context, interface{})
	GetSafe(Context, interface{}) error
	GetRaw(Context) []byte
	Set(Context, interface{})
	SetRaw(Context, []byte)
	Exists(Context) bool
	Delete(Context)
	Equal(Context, interface{}) bool
	Key() []byte

	// XXX: unsafe
	KVStore(Context) KVStore
}

var _ Value = value{}

type value struct {
	base Base
	key  []byte
}

func NewValue(base Base, key []byte) Value {
	return value{
		base: base,
		key:  key,
	}
}

func (v value) store(ctx Context) KVStore {
	return v.base.store(ctx)
}

func (v value) Cdc() *codec.Codec {
	return v.base.Cdc()
}

func (v value) Get(ctx Context, ptr interface{}) {
	v.base.cdc.MustUnmarshalBinaryBare(v.store(ctx).Get(v.key), ptr)
}

func (v value) GetIfExists(ctx Context, ptr interface{}) {
	bz := v.store(ctx).Get(v.key)
	if bz != nil {
		v.base.cdc.MustUnmarshalBinaryBare(bz, ptr)
	}
}

func (v value) GetSafe(ctx Context, ptr interface{}) error {
	bz := v.store(ctx).Get(v.key)
	if bz == nil {
		return ErrEmptyvalue()
	}
	err := v.base.cdc.UnmarshalBinaryBare(bz, ptr)
	if err != nil {
		return ErrUnmarshal(err)
	}
	return nil
}

func (v value) GetRaw(ctx Context) []byte {
	return v.store(ctx).Get(v.key)
}

func (v value) Set(ctx Context, o interface{}) {
	v.store(ctx).Set(v.key, v.base.cdc.MustMarshalBinaryBare(o))
}

func (v value) SetRaw(ctx Context, bz []byte) {
	v.store(ctx).Set(v.key, bz)
}

func (v value) Exists(ctx Context) bool {
	return v.store(ctx).Has(v.key)
}

func (v value) Delete(ctx Context) {
	v.store(ctx).Delete(v.key)
}

func (v value) Equal(ctx Context, o interface{}) bool {
	return bytes.Equal(v.GetRaw(ctx), v.base.cdc.MustMarshalBinaryBare(o))
}

func (v value) Key() []byte {
	return v.base.key(v.key)
}

func (v value) KVStore(ctx Context) KVStore {
	return v.store(ctx)
}

/*
func (v value) KeyPath() KeyPath {
	return v.base.KeyPath().AppendKey(v.key, KeyEncodingHex)
}
*/
type GetSafeErrorType byte

const (
	ErrTypeEmptyvalue GetSafeErrorType = iota
	ErrTypeUnmarshal
)

func (ty GetSafeErrorType) Format(msg string) (res string) {
	switch ty {
	case ErrTypeEmptyvalue:
		res = fmt.Sprintf("Empty value found")
	case ErrTypeUnmarshal:
		res = fmt.Sprintf("Error while unmarshal")
	default:
		panic("Unknown error type")
	}

	if msg != "" {
		res = fmt.Sprintf("%s: %s", res, msg)
	}

	return
}

type GetSafeError struct {
	ty    GetSafeErrorType
	inner error
}

var _ error = (*GetSafeError)(nil) // TODO: sdk.Error

func (err *GetSafeError) Error() string {
	if err.inner == nil {
		return err.ty.Format("")
	}
	return err.ty.Format(err.inner.Error())
}

func ErrEmptyvalue() *GetSafeError {
	return &GetSafeError{
		ty: ErrTypeEmptyvalue,
	}
}

func ErrUnmarshal(err error) *GetSafeError {
	return &GetSafeError{
		ty:    ErrTypeUnmarshal,
		inner: err,
	}
}
