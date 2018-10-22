package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	codec "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	capKey := sdk.NewKVStoreKey("capkey")
	capKey2 := sdk.NewKVStoreKey("capkey2")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(capKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(capKey2, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, capKey, capKey2
}

func TestAccountMapperGetSet(t *testing.T) {
	ms, capKey, _ := setupMultiStore()
	cdc := codec.New()
	RegisterBaseAccount(cdc)

	// make context and mapper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	mapper := NewAccountKeeper(cdc, capKey, ProtoBaseAccount)

	addr := sdk.AccAddress([]byte("some-address"))

	// no account before its created
	acc := mapper.GetAccount(ctx, addr)
	require.Nil(t, acc)

	// create account and check default values
	acc = mapper.NewAccountWithAddress(ctx, addr)
	require.NotNil(t, acc)
	require.Equal(t, addr, acc.GetAddress())
	require.EqualValues(t, nil, acc.GetPubKey())
	require.EqualValues(t, 0, acc.GetSequence())

	// NewAccount doesn't call Set, so it's still nil
	require.Nil(t, mapper.GetAccount(ctx, addr))

	// set some values on the account and save it
	newSequence := int64(20)
	acc.SetSequence(newSequence)
	mapper.SetAccount(ctx, acc)

	// check the new values
	acc = mapper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	require.Equal(t, newSequence, acc.GetSequence())
}

func TestAccountMapperRemoveAccount(t *testing.T) {
	ms, capKey, _ := setupMultiStore()
	cdc := codec.New()
	RegisterBaseAccount(cdc)

	// make context and mapper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	mapper := NewAccountKeeper(cdc, capKey, ProtoBaseAccount)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	// create accounts
	acc1 := mapper.NewAccountWithAddress(ctx, addr1)
	acc2 := mapper.NewAccountWithAddress(ctx, addr2)

	accSeq1 := int64(20)
	accSeq2 := int64(40)

	acc1.SetSequence(accSeq1)
	acc2.SetSequence(accSeq2)
	mapper.SetAccount(ctx, acc1)
	mapper.SetAccount(ctx, acc2)

	acc1 = mapper.GetAccount(ctx, addr1)
	require.NotNil(t, acc1)
	require.Equal(t, accSeq1, acc1.GetSequence())

	// remove one account
	mapper.RemoveAccount(ctx, acc1)
	acc1 = mapper.GetAccount(ctx, addr1)
	require.Nil(t, acc1)

	acc2 = mapper.GetAccount(ctx, addr2)
	require.NotNil(t, acc2)
	require.Equal(t, accSeq2, acc2.GetSequence())
}

func BenchmarkAccountMapperGetAccountFound(b *testing.B) {
	ms, capKey, _ := setupMultiStore()
	cdc := codec.New()
	RegisterBaseAccount(cdc)

	// make context and mapper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	mapper := NewAccountKeeper(cdc, capKey, ProtoBaseAccount)

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := mapper.NewAccountWithAddress(ctx, addr)
		mapper.SetAccount(ctx, acc)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		mapper.GetAccount(ctx, sdk.AccAddress(arr))
	}
}
