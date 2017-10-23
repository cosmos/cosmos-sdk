package eyes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wire "github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/cosmos/cosmos-sdk/util"
)

func TestHandler(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	key := []byte("one")
	val := []byte("foo")
	var height uint64 = 123

	h := NewHandler()
	ctx := util.MockContext("role-chain", height)
	store := state.NewMemKVStore()

	set := sdk.WrapTx(NewSetTx(key, val))
	remove := sdk.WrapTx(NewRemoveTx(key))
	invalid := sdk.WrapTx(NewSetTx(nil, nil))

	// make sure pricing makes sense
	cres, err := h.CheckTx(ctx, store, set)
	require.Nil(err, "%+v", err)
	require.True(cres.GasAllocated > 5, "%#v", cres)

	// set the value, no error
	dres, err := h.DeliverTx(ctx, store, set)
	require.Nil(err, "%+v", err)

	// get the data
	var data Data
	bs := store.Get(key)
	require.NotEmpty(bs)
	err = wire.ReadBinaryBytes(bs, &data)
	require.Nil(err, "%+v", err)
	assert.Equal(height, data.SetAt)
	assert.EqualValues(val, data.Value)

	// make sure pricing makes sense
	cres, err = h.CheckTx(ctx, store, remove)
	require.Nil(err, "%+v", err)
	require.True(cres.GasAllocated > 5, "%#v", cres)

	// remove the data returns the same as the above query
	dres, err = h.DeliverTx(ctx, store, remove)
	require.Nil(err, "%+v", err)
	require.EqualValues(bs, dres.Data)

	// make sure invalid fails both ways
	_, err = h.CheckTx(ctx, store, invalid)
	require.NotNil(err)
	_, err = h.DeliverTx(ctx, store, invalid)
	require.NotNil(err)
}
