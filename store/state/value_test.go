package state

import (
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var testcdc = codec.New()
var testkey = sdk.NewKVStoreKey("test")

func init() {
	// register
}

func key() (res []byte) {
	res = make([]byte, 64)
	rand.Read(res)
	return
}

type value interface {
	KeyBytes() []byte
	Get(Context, interface{})
	GetSafe(Context, interface{}) error
	GetRaw(Context) []byte
	Set(Context, interface{})
	SetRaw(Context, []byte)
	Exists(Context) bool
	Delete(Context)
	Query(ABCIQuerier, interface{}) (*Proof, error)
	Marshal(interface{}) []byte
	Unmarshal([]byte, interface{})
}

type typeValue interface {
	value
	Proto() interface{}
}

type valueT struct {
	Value
}

var _ value = valueT{}

func (v valueT) Marshal(o interface{}) []byte {
	return v.m.cdc.MustMarshalBinaryBare(o)
}

func (v valueT) Unmarshal(bz []byte, ptr interface{}) {
	v.m.cdc.MustUnmarshalBinaryBare(bz, ptr)
}

type booleanT struct {
	Boolean
}

var _ typeValue = booleanT{}

func newBoolean() booleanT {
	return booleanT{NewMapping(testkey, testcdc, nil).Value(key()).Boolean()}
}

func (booleanT) Proto() interface{} {
	return new(bool)
}

func (v booleanT) Get(ctx Context, ptr interface{}) {
	reflect.ValueOf(ptr).Elem().SetBool(v.Boolean.Get(ctx))
}

func (v booleanT) GetSafe(ctx Context, ptr interface{}) error {
	res, err := v.Boolean.GetSafe(ctx)
	if err != nil {
		return err
	}
	reflect.ValueOf(ptr).Elem().SetBool(res)
	return nil
}

func (v booleanT) Set(ctx Context, o interface{}) {
	v.Boolean.Set(ctx, o.(bool))
}

func (v booleanT) Marshal(o interface{}) []byte {
	switch o.(bool) {
	case false:
		return []byte{0x00}
	case true:
		return []byte{0x01}
	}
	panic("invalid boolean type")
}

func (v booleanT) Unmarshal(bz []byte, ptr interface{}) {
	switch bz[0] {
	case 0x00:
		reflect.ValueOf(ptr).Elem().SetBool(false)
	case 0x01:
		reflect.ValueOf(ptr).Elem().SetBool(true)
	}
}

func (v booleanT) Query(q ABCIQuerier, ptr interface{}) (proof *Proof, err error) {
	res, proof, err := v.Boolean.Query(q)
	if err != nil {
		return
	}
	reflect.ValueOf(ptr).Elem().SetBool(res)
	return
}

type integerT struct {
	Integer
}

var _ typeValue = integerT{}

func newInteger(enc IntEncoding) integerT {
	return integerT{NewMapping(testkey, testcdc, nil).Value(key()).Integer(enc)}
}

func (integerT) Proto() interface{} {
	return new(uint64)
}

func (v integerT) Get(ctx Context, ptr interface{}) {
	reflect.ValueOf(ptr).Elem().SetUint(v.Integer.Get(ctx))
}

func (v integerT) GetSafe(ctx Context, ptr interface{}) error {
	res, err := v.Integer.GetSafe(ctx)
	if err != nil {
		return err
	}
	reflect.ValueOf(ptr).Elem().SetUint(res)
	return nil
}

func (v integerT) Set(ctx Context, o interface{}) {
	v.Integer.Set(ctx, o.(uint64))
}

func (v integerT) Marshal(o interface{}) []byte {
	return EncodeInt(o.(uint64), v.enc)
}

func (v integerT) Unmarshal(bz []byte, ptr interface{}) {
	res, err := DecodeInt(bz, v.enc)
	if err != nil {
		panic(err)
	}
	reflect.ValueOf(ptr).Elem().SetUint(res)
}

func (v integerT) Query(q ABCIQuerier, ptr interface{}) (proof *Proof, err error) {
	res, proof, err := v.Integer.Query(q)
	if err != nil {
		return
	}
	reflect.ValueOf(ptr).Elem().SetUint(res)
	return
}

type enumT struct {
	Enum
}

var _ typeValue = enumT{}

func newEnum() enumT {
	return enumT{NewMapping(testkey, testcdc, nil).Value(key()).Enum()}
}

func (enumT) Proto() interface{} {
	return new(byte)
}

func (v enumT) Get(ctx Context, ptr interface{}) {
	reflect.ValueOf(ptr).Elem().Set(reflect.ValueOf(v.Enum.Get(ctx)))
}

func (v enumT) GetSafe(ctx Context, ptr interface{}) error {
	res, err := v.Enum.GetSafe(ctx)
	if err != nil {
		return err
	}
	reflect.ValueOf(ptr).Elem().Set(reflect.ValueOf(res))
	return nil
}

func (v enumT) Set(ctx Context, o interface{}) {
	v.Enum.Set(ctx, o.(byte))
}

func (v enumT) Marshal(o interface{}) []byte {
	return []byte{o.(byte)}
}

func (v enumT) Unmarshal(bz []byte, ptr interface{}) {
	reflect.ValueOf(ptr).Elem().Set(reflect.ValueOf(bz[0]))
}

func (v enumT) Query(q ABCIQuerier, ptr interface{}) (proof *Proof, err error) {
	res, proof, err := v.Enum.Query(q)
	if err != nil {
		return
	}
	reflect.ValueOf(ptr).Elem().Set(reflect.ValueOf(res))
	return
}

type stringT struct {
	String
}

var _ typeValue = stringT{}

func newString() stringT {
	return stringT{NewMapping(testkey, testcdc, nil).Value(key()).String()}
}

func (stringT) Proto() interface{} {
	return new(string)
}

func (v stringT) Get(ctx Context, ptr interface{}) {
	reflect.ValueOf(ptr).Elem().SetString(v.String.Get(ctx))
}

func (v stringT) GetSafe(ctx Context, ptr interface{}) error {
	res, err := v.String.GetSafe(ctx)
	if err != nil {
		return err
	}
	reflect.ValueOf(ptr).Elem().SetString(res)
	return nil
}

func (v stringT) Set(ctx Context, o interface{}) {
	v.String.Set(ctx, o.(string))
}

func (v stringT) Marshal(o interface{}) []byte {
	return []byte(o.(string))
}

func (v stringT) Unmarshal(bz []byte, ptr interface{}) {
	reflect.ValueOf(ptr).Elem().SetString(string(bz))
}

func (v stringT) Query(q ABCIQuerier, ptr interface{}) (proof *Proof, err error) {
	res, proof, err := v.String.Query(q)
	if err != nil {
		return
	}
	reflect.ValueOf(ptr).Elem().SetString(res)
	return
}

func defaultComponents() (sdk.Context, *rootmulti.Store) {
	db := dbm.NewMemDB()
	cms := rootmulti.NewStore(db)
	cms.MountStoreWithDB(testkey, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	return ctx, cms
}

func indirect(ptr interface{}) interface{} {
	return reflect.ValueOf(ptr).Elem().Interface()
}

func TestTypeValue(t *testing.T) {
	ctx, cms := defaultComponents()

	var table = []struct {
		ty   typeValue
		orig interface{}
	}{
		{newBoolean(), false},
		{newBoolean(), true},
		{newInteger(Dec), uint64(1024000)},
		{newInteger(Dec), uint64(2048000)},
		{newInteger(Bin), uint64(4096000)},
		{newInteger(Bin), uint64(8192000)},
		{newInteger(Hex), uint64(16384000)},
		{newInteger(Hex), uint64(32768000)},
		{newEnum(), byte(0x00)},
		{newEnum(), byte(0x78)},
		{newEnum(), byte(0xA0)},
		{newString(), "1234567890"},
		{newString(), "asdfghjkl"},
		{newString(), "qwertyuiop"},
	}

	for i, tc := range table {
		v := tc.ty
		// Exists expected false
		require.False(t, v.Exists(ctx))

		// Simple get-set
		v.Set(ctx, tc.orig)
		ptr := v.Proto()
		v.Get(ctx, ptr)
		require.Equal(t, tc.orig, indirect(ptr), "Expected equal on tc %d", i)
		ptr = v.Proto()
		err := v.GetSafe(ctx, ptr)
		require.NoError(t, err)
		require.Equal(t, tc.orig, indirect(ptr), "Expected equal on tc %d", i)

		// Raw get
		require.Equal(t, v.Marshal(tc.orig), v.GetRaw(ctx), "Expected equal on tc %d", i)

		// Exists expected true
		require.True(t, v.Exists(ctx))

		// After delete
		v.Delete(ctx)
		require.False(t, v.Exists(ctx))
		ptr = v.Proto()
		err = v.GetSafe(ctx, ptr)
		require.Error(t, err)
		require.Equal(t, reflect.Zero(reflect.TypeOf(ptr).Elem()).Interface(), indirect(ptr))
		require.Nil(t, v.GetRaw(ctx))

		// Set again and test abci query
		v.Set(ctx, tc.orig)
		cid := cms.Commit()
		ptr = v.Proto()
		q := NewStoreQuerier(cms)
		proof, err := v.Query(q, ptr)
		require.NoError(t, err)
		require.Equal(t, tc.orig, indirect(ptr), "Expected equal on tc %d", i)
		prt := rootmulti.DefaultProofRuntime()
		kp := merkle.KeyPath{}.
			AppendKey([]byte(testkey.Name()), merkle.KeyEncodingHex).
			AppendKey(v.KeyBytes(), merkle.KeyEncodingHex)
		require.NoError(t, prt.VerifyValue(proof, cid.Hash, kp.String(), v.GetRaw(ctx)))
	}
}
