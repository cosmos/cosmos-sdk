package assoc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/examples/democoin/mock"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

func defaultContext(key sdk.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, nil)
	return ctx
}

func TestValidatorSet(t *testing.T) {
	key := sdk.NewKVStoreKey("test")
	ctx := defaultContext(key)

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")

	base := &mock.ValidatorSet{[]mock.Validator{
		{addr1, sdk.NewRat(1)},
		{addr2, sdk.NewRat(2)},
	}}

	valset := NewValidatorSet(wire.NewCodec(), sdk.NewPrefixStoreGetter(key, []byte("assoc")), base, 1, 5)

	require.Equal(t, base.Validator(ctx, addr1), valset.Validator(ctx, addr1))
	require.Equal(t, base.Validator(ctx, addr2), valset.Validator(ctx, addr2))

	assoc1 := []byte("asso1")
	assoc2 := []byte("asso2")

	require.True(t, valset.Associate(ctx, addr1, assoc1))
	require.True(t, valset.Associate(ctx, addr2, assoc2))

	require.Equal(t, base.Validator(ctx, addr1), valset.Validator(ctx, assoc1))
	require.Equal(t, base.Validator(ctx, addr2), valset.Validator(ctx, assoc2))

	require.Equal(t, base.Validator(ctx, addr1), valset.Validator(ctx, addr1))
	require.Equal(t, base.Validator(ctx, addr2), valset.Validator(ctx, addr2))

	assocs := valset.Associations(ctx, addr1)
	require.Equal(t, 1, len(assocs))
	require.True(t, bytes.Equal(assoc1, assocs[0]))

	require.False(t, valset.Associate(ctx, addr1, assoc2))
	require.False(t, valset.Associate(ctx, addr2, assoc1))

	valset.Dissociate(ctx, addr1, assoc1)
	valset.Dissociate(ctx, addr2, assoc2)

	require.Equal(t, base.Validator(ctx, addr1), valset.Validator(ctx, addr1))
	require.Equal(t, base.Validator(ctx, addr2), valset.Validator(ctx, addr2))

	require.Nil(t, valset.Validator(ctx, assoc1))
	require.Nil(t, valset.Validator(ctx, assoc2))
}
