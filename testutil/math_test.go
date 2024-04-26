package testutil_test

import (
	"encoding/json"
	"fmt"
	"testing"

	math "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

type testLegDecCoin struct {
	Denom  string
	Amount int64
	Fee    sdk.DecCoin
}

type testDecCoin struct {
	Denom  string
	Amount int64
	Fee    math.Dec
}

func TestDiffDecimalsMigrationWithLDec(t *testing.T) {
	key := storetypes.NewKVStoreKey("test")
	ctx := testutil.DefaultContext(key, storetypes.NewTransientStoreKey("transient"))

	err := testutil.DiffCollectionsMigration(
		ctx,
		key,
		100,
		func(i int64) {
			legacyDec := testLegDecCoin{
				Denom:  "test",
				Amount: i,
				Fee:    sdk.NewDecCoinFromDec("test", math.LegacyNewDec(100)),
			}

			feeBytes, err := json.Marshal(legacyDec.Fee)
			if err != nil {
				t.Fatal(err)
			}

			ctx.KVStore(key).Set([]byte(fmt.Sprintf("%d", i)), feeBytes)
		},
		"somerandomhashtostartwith",
	)
	require.Error(t, err)

	ctx = testutil.DefaultContext(key, storetypes.NewTransientStoreKey("transient"))

	err = testutil.DiffCollectionsMigration(
		ctx,
		key,
		100,
		func(i int64) {
			legacyDec := testLegDecCoin{
				Denom:  "test",
				Amount: i,
				Fee:    sdk.NewDecCoinFromDec("test", math.LegacyNewDec(100)),
			}

			feeBytes, err := json.Marshal(legacyDec.Fee)
			if err != nil {
				t.Fatal(err)
			}

			ctx.KVStore(key).Set([]byte(fmt.Sprintf("%d", i)), feeBytes)
		},
		"4b782f32948a596f8507f09817eec2307ce2ffee1aba5a548004cccb062ccdbd",
	)
	require.NoError(t, err)


	ctx = testutil.DefaultContext(key, storetypes.NewTransientStoreKey("transient"))

	err = testutil.DiffCollectionsMigration(
		ctx,
		key,
		100,
		func(i int64) {
			Dec := testDecCoin{
				Denom:  "test",
				Amount: i,
				Fee:    math.NewDecFromInt64(100),
			}

			feeBytes, err := json.Marshal(Dec.Fee)
			if err != nil {
				t.Fatal(err)
			}

			ctx.KVStore(key).Set([]byte(fmt.Sprintf("%d", i)), feeBytes)
		},
		"4b782f32948a596f8507f09817eec2307ce2ffee1aba5a548004cccb062ccdbd",
	)
	require.NoError(t, err)
}
