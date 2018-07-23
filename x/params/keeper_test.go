package params

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

func defaultContext(key sdk.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	return ctx
}

func TestKeeper(t *testing.T) {
	kvs := []struct {
		key   string
		param int64
	}{
		{"key1", 10},
		{"key2", 55},
		{"key3", 182},
		{"key4", 17582},
		{"key5", 2768554},
	}

	skey := sdk.NewKVStoreKey("test")
	ctx := defaultContext(skey)
	setter := NewKeeper(wire.NewCodec(), skey).Setter()

	for _, kv := range kvs {
		err := setter.Set(ctx, kv.key, kv.param)
		assert.Nil(t, err)
	}

	for _, kv := range kvs {
		var param int64
		err := setter.Get(ctx, kv.key, &param)
		assert.Nil(t, err)
		assert.Equal(t, kv.param, param)
	}

	cdc := wire.NewCodec()
	for _, kv := range kvs {
		var param int64
		bz := setter.GetRaw(ctx, kv.key)
		err := cdc.UnmarshalBinary(bz, &param)
		assert.Nil(t, err)
		assert.Equal(t, kv.param, param)
	}

	for _, kv := range kvs {
		var param bool
		err := setter.Get(ctx, kv.key, &param)
		assert.NotNil(t, err)
	}

	for _, kv := range kvs {
		err := setter.Set(ctx, kv.key, true)
		assert.NotNil(t, err)
	}
}

func TestGetter(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx := defaultContext(key)
	keeper := NewKeeper(wire.NewCodec(), key)

	g := keeper.Getter()
	s := keeper.Setter()

	kvs := []struct {
		key   string
		param interface{}
	}{
		{"string", "test"},
		{"bool", true},
		{"int16", int16(1)},
		{"int32", int32(1)},
		{"int64", int64(1)},
		{"uint16", uint16(1)},
		{"uint32", uint32(1)},
		{"uint64", uint64(1)},
		{"int", sdk.NewInt(1)},
		{"uint", sdk.NewUint(1)},
		{"rat", sdk.NewRat(1)},
	}

	assert.NotPanics(t, func() { s.SetString(ctx, kvs[0].key, "test") })
	assert.NotPanics(t, func() { s.SetBool(ctx, kvs[1].key, true) })
	assert.NotPanics(t, func() { s.SetInt16(ctx, kvs[2].key, int16(1)) })
	assert.NotPanics(t, func() { s.SetInt32(ctx, kvs[3].key, int32(1)) })
	assert.NotPanics(t, func() { s.SetInt64(ctx, kvs[4].key, int64(1)) })
	assert.NotPanics(t, func() { s.SetUint16(ctx, kvs[5].key, uint16(1)) })
	assert.NotPanics(t, func() { s.SetUint32(ctx, kvs[6].key, uint32(1)) })
	assert.NotPanics(t, func() { s.SetUint64(ctx, kvs[7].key, uint64(1)) })
	assert.NotPanics(t, func() { s.SetInt(ctx, kvs[8].key, sdk.NewInt(1)) })
	assert.NotPanics(t, func() { s.SetUint(ctx, kvs[9].key, sdk.NewUint(1)) })
	assert.NotPanics(t, func() { s.SetRat(ctx, kvs[10].key, sdk.NewRat(1)) })

	var res interface{}
	var err error

	// String
	def0 := "default"
	res, err = g.GetString(ctx, kvs[0].key)
	assert.Nil(t, err)
	assert.Equal(t, kvs[0].param, res)

	_, err = g.GetString(ctx, "invalid")
	assert.NotNil(t, err)

	res = g.GetStringWithDefault(ctx, kvs[0].key, def0)
	assert.Equal(t, kvs[0].param, res)

	res = g.GetStringWithDefault(ctx, "invalid", def0)
	assert.Equal(t, def0, res)

	// Bool
	def1 := false
	res, err = g.GetBool(ctx, kvs[1].key)
	assert.Nil(t, err)
	assert.Equal(t, kvs[1].param, res)

	_, err = g.GetBool(ctx, "invalid")
	assert.NotNil(t, err)

	res = g.GetBoolWithDefault(ctx, kvs[1].key, def1)
	assert.Equal(t, kvs[1].param, res)

	res = g.GetBoolWithDefault(ctx, "invalid", def1)
	assert.Equal(t, def1, res)

	// Int16
	def2 := int16(0)
	res, err = g.GetInt16(ctx, kvs[2].key)
	assert.Nil(t, err)
	assert.Equal(t, kvs[2].param, res)

	_, err = g.GetInt16(ctx, "invalid")
	assert.NotNil(t, err)

	res = g.GetInt16WithDefault(ctx, kvs[2].key, def2)
	assert.Equal(t, kvs[2].param, res)

	res = g.GetInt16WithDefault(ctx, "invalid", def2)
	assert.Equal(t, def2, res)

	// Int32
	def3 := int32(0)
	res, err = g.GetInt32(ctx, kvs[3].key)
	assert.Nil(t, err)
	assert.Equal(t, kvs[3].param, res)

	_, err = g.GetInt32(ctx, "invalid")
	assert.NotNil(t, err)

	res = g.GetInt32WithDefault(ctx, kvs[3].key, def3)
	assert.Equal(t, kvs[3].param, res)

	res = g.GetInt32WithDefault(ctx, "invalid", def3)
	assert.Equal(t, def3, res)

	// Int64
	def4 := int64(0)
	res, err = g.GetInt64(ctx, kvs[4].key)
	assert.Nil(t, err)
	assert.Equal(t, kvs[4].param, res)

	_, err = g.GetInt64(ctx, "invalid")
	assert.NotNil(t, err)

	res = g.GetInt64WithDefault(ctx, kvs[4].key, def4)
	assert.Equal(t, kvs[4].param, res)

	res = g.GetInt64WithDefault(ctx, "invalid", def4)
	assert.Equal(t, def4, res)

	// Uint16
	def5 := uint16(0)
	res, err = g.GetUint16(ctx, kvs[5].key)
	assert.Nil(t, err)
	assert.Equal(t, kvs[5].param, res)

	_, err = g.GetUint16(ctx, "invalid")
	assert.NotNil(t, err)

	res = g.GetUint16WithDefault(ctx, kvs[5].key, def5)
	assert.Equal(t, kvs[5].param, res)

	res = g.GetUint16WithDefault(ctx, "invalid", def5)
	assert.Equal(t, def5, res)

	// Uint32
	def6 := uint32(0)
	res, err = g.GetUint32(ctx, kvs[6].key)
	assert.Nil(t, err)
	assert.Equal(t, kvs[6].param, res)

	_, err = g.GetUint32(ctx, "invalid")
	assert.NotNil(t, err)

	res = g.GetUint32WithDefault(ctx, kvs[6].key, def6)
	assert.Equal(t, kvs[6].param, res)

	res = g.GetUint32WithDefault(ctx, "invalid", def6)
	assert.Equal(t, def6, res)

	// Uint64
	def7 := uint64(0)
	res, err = g.GetUint64(ctx, kvs[7].key)
	assert.Nil(t, err)
	assert.Equal(t, kvs[7].param, res)

	_, err = g.GetUint64(ctx, "invalid")
	assert.NotNil(t, err)

	res = g.GetUint64WithDefault(ctx, kvs[7].key, def7)
	assert.Equal(t, kvs[7].param, res)

	res = g.GetUint64WithDefault(ctx, "invalid", def7)
	assert.Equal(t, def7, res)

	// Int
	def8 := sdk.NewInt(0)
	res, err = g.GetInt(ctx, kvs[8].key)
	assert.Nil(t, err)
	assert.Equal(t, kvs[8].param, res)

	_, err = g.GetInt(ctx, "invalid")
	assert.NotNil(t, err)

	res = g.GetIntWithDefault(ctx, kvs[8].key, def8)
	assert.Equal(t, kvs[8].param, res)

	res = g.GetIntWithDefault(ctx, "invalid", def8)
	assert.Equal(t, def8, res)

	// Uint
	def9 := sdk.NewUint(0)
	res, err = g.GetUint(ctx, kvs[9].key)
	assert.Nil(t, err)
	assert.Equal(t, kvs[9].param, res)

	_, err = g.GetUint(ctx, "invalid")
	assert.NotNil(t, err)

	res = g.GetUintWithDefault(ctx, kvs[9].key, def9)
	assert.Equal(t, kvs[9].param, res)

	res = g.GetUintWithDefault(ctx, "invalid", def9)
	assert.Equal(t, def9, res)

	// Rat
	def10 := sdk.NewRat(0)
	res, err = g.GetRat(ctx, kvs[10].key)
	assert.Nil(t, err)
	assert.Equal(t, kvs[10].param, res)

	_, err = g.GetRat(ctx, "invalid")
	assert.NotNil(t, err)

	res = g.GetRatWithDefault(ctx, kvs[10].key, def10)
	assert.Equal(t, kvs[10].param, res)

	res = g.GetRatWithDefault(ctx, "invalid", def10)
	assert.Equal(t, def10, res)

}
