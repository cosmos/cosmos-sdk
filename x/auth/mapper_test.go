package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

func setupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	capKey := sdk.NewKVStoreKey("capkey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(capKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, capKey
}

func TestAccountMapperGetSet(t *testing.T) {
	ms, capKey := setupMultiStore()
	cdc := wire.NewCodec()
	RegisterBaseAccount(cdc)

	// make context and mapper
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	mapper := NewAccountMapper(cdc, capKey, &BaseAccount{})

	addr := sdk.Address([]byte("some-address"))

	// no account before its created
	acc := mapper.GetAccount(ctx, addr)
	assert.Nil(t, acc)

	// create account and check default values
	acc = mapper.NewAccountWithAddress(ctx, addr)
	assert.NotNil(t, acc)
	assert.Equal(t, addr, acc.GetAddress())
	assert.EqualValues(t, nil, acc.GetPubKey())
	assert.EqualValues(t, 0, acc.GetSequence())

	// NewAccount doesn't call Set, so it's still nil
	assert.Nil(t, mapper.GetAccount(ctx, addr))

	// set some values on the account and save it
	newSequence := int64(20)
	acc.SetSequence(newSequence)
	mapper.SetAccount(ctx, acc)

	// check the new values
	acc = mapper.GetAccount(ctx, addr)
	assert.NotNil(t, acc)
	assert.Equal(t, newSequence, acc.GetSequence())
}
