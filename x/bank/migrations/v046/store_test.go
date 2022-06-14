package v046_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/depinject"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	v043 "github.com/cosmos/cosmos-sdk/x/bank/migrations/v043"
	v046 "github.com/cosmos/cosmos-sdk/x/bank/migrations/v046"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestMigrateStore(t *testing.T) {
	var encCfg codec.Codec
	depinject.Inject(banktestutil.AppConfig, &encCfg)
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
		bz, err := encCfg.Marshal(&b)
		require.NoError(t, err)

		prefixAccStore.Set([]byte(b.Denom), bz)
	}

	require.NoError(t, v046.MigrateStore(ctx, bankKey, encCfg))

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
	var encCfg codec.Codec
	depinject.Inject(banktestutil.AppConfig, &encCfg)
	bankKey := sdk.NewKVStoreKey("bank")
	ctx := testutil.DefaultContext(bankKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(bankKey)
	metaData := []types.Metadata{
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
	denomMetadataStore := prefix.NewStore(store, v043.DenomMetadataPrefix)

	for i := range []int{0, 1} {
		key := append(v043.DenomMetadataPrefix, []byte(metaData[i].Base)...)
		// keys before 0.45 had denom two times in the key
		key = append(key, []byte(metaData[i].Base)...)
		bz, err := encCfg.Marshal(&metaData[i])
		require.NoError(t, err)
		denomMetadataStore.Set(key, bz)
	}

	require.NoError(t, v046.MigrateStore(ctx, bankKey, encCfg))

	denomMetadataStore = prefix.NewStore(store, v043.DenomMetadataPrefix)
	denomMetadataIter := denomMetadataStore.Iterator(nil, nil)
	defer denomMetadataIter.Close()
	for i := 0; denomMetadataIter.Valid(); denomMetadataIter.Next() {
		var result types.Metadata
		newKey := denomMetadataIter.Key()

		// make sure old entry is deleted
		oldKey := append(newKey, newKey[1:]...)
		bz := denomMetadataStore.Get(oldKey)
		require.Nil(t, bz)

		require.Equal(t, string(newKey)[1:], metaData[i].Base, "idx: %d", i)
		bz = denomMetadataStore.Get(denomMetadataIter.Key())
		require.NotNil(t, bz)
		err := encCfg.Unmarshal(bz, &result)
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
