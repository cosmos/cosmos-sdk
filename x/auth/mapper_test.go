package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
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
	cdc := wire.NewCodec()
	RegisterBaseAccount(cdc)

	// make context and mapper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	mapper := NewAccountMapper(cdc, capKey, &BaseAccount{})

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
