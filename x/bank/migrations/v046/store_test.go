package v046_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	v043 "github.com/cosmos/cosmos-sdk/x/bank/migrations/v043"
	v046 "github.com/cosmos/cosmos-sdk/x/bank/migrations/v046"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var (
	metaData = []types.Metadata{
		{
			Name:        "Cosmos Hub Atom",
			Symbol:      "ATOM",
			Description: "The native staking token of the Cosmos Hub.",
			DenomUnits: []*types.DenomUnit{
				{"uatom", uint32(0), []string{"microatom"}},
				{"matom", uint32(3), []string{"milliatom"}},
				{"atom", uint32(6), nil},
			},
			Base:    "uatom",
			Display: "atom",
		},
		{
			Name:        "Token",
			Symbol:      "TOKEN",
			Description: "The native staking token of the Token Hub.",
			DenomUnits: []*types.DenomUnit{
				{"1token", uint32(5), []string{"decitoken"}},
				{"2token", uint32(4), []string{"centitoken"}},
				{"3token", uint32(7), []string{"dekatoken"}},
			},
			Base:    "utoken",
			Display: "token",
		},
	}
)

func TestMigrateStore(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	bankKey := sdk.NewKVStoreKey("bank")
	ctx := testutil.DefaultContext(bankKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(bankKey)

	addr := sdk.AccAddress([]byte("addr________________"))
	prefixAccStore := prefix.NewStore(store, v043.CreateAccountBalancesPrefix(addr))

	balances := sdk.NewCoins(
		sdk.NewCoin("foo", sdk.NewInt(10000)),
		sdk.NewCoin("bar", sdk.NewInt(20000)),
	)

	for _, b := range balances {
		bz, err := encCfg.Codec.Marshal(&b)
		require.NoError(t, err)

		prefixAccStore.Set([]byte(b.Denom), bz)
	}

	require.NoError(t, v046.MigrateStore(ctx, bankKey, encCfg.Codec))

	for _, b := range balances {
		addrPrefixStore := prefix.NewStore(store, types.CreateAccountBalancesPrefix(addr))
		bz := addrPrefixStore.Get([]byte(b.Denom))
		var expected sdk.Int
		require.NoError(t, expected.Unmarshal(bz))
		require.Equal(t, expected, b.Amount)
	}

	for _, b := range balances {
		denomPrefixStore := prefix.NewStore(store, v046.CreateDenomAddressPrefix(b.Denom))
		bz := denomPrefixStore.Get(address.MustLengthPrefix(addr))
		require.NotNil(t, bz)
	}
}

func TestMigrateDenomMetaData(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	bankKey := sdk.NewKVStoreKey("bank")
	ctx := testutil.DefaultContext(bankKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(bankKey)
	denomMetadataStore := prefix.NewStore(store, v043.DenomMetadataPrefix)

	for i := range []int{0, 1} {
		// keys before 0.45 had denom two times in the key
		key := append([]byte{}, []byte(metaData[i].Base)...)
		key = append(key, []byte(metaData[i].Base)...)
		bz, err := encCfg.Codec.Marshal(&metaData[i])
		require.NoError(t, err)
		denomMetadataStore.Set(key, bz)
	}

	require.NoError(t, v046.MigrateStore(ctx, bankKey, encCfg.Codec))

	denomMetadataStore = prefix.NewStore(store, v043.DenomMetadataPrefix)
	assertCorrectDenomKeys(t, denomMetadataStore, encCfg.Codec)
}

// migrateDenomMetadataV0464 is the denom metadata migration function present
// in v0.46.4. It is buggy, as discovered in https://github.com/cosmos/cosmos-sdk/pull/13821.
// It is copied verbatim here to test the helper function Migrate_V046_4_To_V046_5
// which aims to fix the bug on chains already on v0.46.
//
// Copied from:
// https://github.com/cosmos/cosmos-sdk/blob/v0.46.4/x/bank/migrations/v046/store.go#L75-L94
func migrateDenomMetadataV0464(store sdk.KVStore) error {
	oldDenomMetaDataStore := prefix.NewStore(store, v043.DenomMetadataPrefix)

	oldDenomMetaDataIter := oldDenomMetaDataStore.Iterator(nil, nil)
	defer oldDenomMetaDataIter.Close()

	for ; oldDenomMetaDataIter.Valid(); oldDenomMetaDataIter.Next() {
		oldKey := oldDenomMetaDataIter.Key()
		l := len(oldKey)/2 + 1

		newKey := make([]byte, len(types.DenomMetadataPrefix)+l)
		// old key: prefix_bytes | denom_bytes | denom_bytes
		copy(newKey, types.DenomMetadataPrefix)
		copy(newKey[len(types.DenomMetadataPrefix):], oldKey[:l])
		store.Set(newKey, oldDenomMetaDataIter.Value())
		oldDenomMetaDataStore.Delete(oldKey)
	}

	return nil
}

func TestMigrate_V046_4_To_V046_5(t *testing.T) {
	// Step 1. Create a v0.43 state.
	encCfg := simapp.MakeTestEncodingConfig()
	bankKey := sdk.NewKVStoreKey("bank")
	ctx := testutil.DefaultContext(bankKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(bankKey)
	denomMetadataStore := prefix.NewStore(store, v046.DenomMetadataPrefix)

	for i := range []int{0, 1} {
		// keys before 0.45 had denom two times in the key
		key := append([]byte{}, []byte(metaData[i].Base)...)
		key = append(key, []byte(metaData[i].Base)...)
		bz, err := encCfg.Codec.Marshal(&metaData[i])
		require.NoError(t, err)
		denomMetadataStore.Set(key, bz)
	}

	// Step 2. Migrate to v0.46 using the BUGGED migration (present in<=v0.46.4).
	require.NoError(t, migrateDenomMetadataV0464(store))

	denomMetadataIter := denomMetadataStore.Iterator(nil, nil)
	defer denomMetadataIter.Close()
	for i := 0; denomMetadataIter.Valid(); denomMetadataIter.Next() {
		newKey := denomMetadataIter.Key()
		require.NotEqual(t, string(newKey), metaData[i].Base, "idx: %d", i) // not equal, because we had wrong keys
		i++
	}

	// Step 3. Use the helper function to migrate to a correct v0.46.5 state.
	require.NoError(t, v046.Migrate_V046_4_To_V046_5(store))

	assertCorrectDenomKeys(t, denomMetadataStore, encCfg.Codec)
}

// assertCorrectDenomKeys makes sure the denom keys present in state are
// correct and resolve to the correct metadata.
func assertCorrectDenomKeys(t *testing.T, denomMetadataStore prefix.Store, cdc codec.Codec) {
	denomMetadataIter := denomMetadataStore.Iterator(nil, nil)
	defer denomMetadataIter.Close()
	for i := 0; denomMetadataIter.Valid(); denomMetadataIter.Next() {
		var result types.Metadata
		newKey := denomMetadataIter.Key()

		// make sure old entry is deleted
		oldKey := append(newKey, newKey[0:]...)
		bz := denomMetadataStore.Get(oldKey)
		require.Nil(t, bz)

		require.Equal(t, string(newKey), metaData[i].Base, "idx: %d", i)
		bz = denomMetadataStore.Get(denomMetadataIter.Key())
		require.NotNil(t, bz)
		err := cdc.Unmarshal(bz, &result)
		require.NoError(t, err)
		assertMetaDataEqual(t, metaData[i], result)
		i++
	}
}

func assertMetaDataEqual(t *testing.T, expected, actual types.Metadata) {
	require.Equal(t, expected.GetBase(), actual.GetBase())
	require.Equal(t, expected.GetDisplay(), actual.GetDisplay())
	require.Equal(t, expected.GetDescription(), actual.GetDescription())
	require.Equal(t, expected.GetDenomUnits()[1].GetDenom(), actual.GetDenomUnits()[1].GetDenom())
	require.Equal(t, expected.GetDenomUnits()[1].GetExponent(), actual.GetDenomUnits()[1].GetExponent())
	require.Equal(t, expected.GetDenomUnits()[1].GetAliases(), actual.GetDenomUnits()[1].GetAliases())
}
