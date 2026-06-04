package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testapp "github.com/cosmos/cosmos-sdk/testutil/testapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BenchmarkAccountMapperGetAccountFound(b *testing.B) {
	b.ReportAllocs()
	ta := testapp.Setup(b)
	require.NotNil(b, ta)

	ctx := testapp.NewContext(ta)
	accountKeeper := ta.AccountKeeper

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := accountKeeper.NewAccountWithAddress(ctx, addr)
		accountKeeper.SetAccount(ctx, acc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		accountKeeper.GetAccount(ctx, sdk.AccAddress(arr))
	}
}

func BenchmarkAccountMapperSetAccount(b *testing.B) {
	b.ReportAllocs()
	ta := testapp.Setup(b)
	require.NotNil(b, ta)

	ctx := testapp.NewContext(ta)
	accountKeeper := ta.AccountKeeper

	// assumes b.N < 2**24
	for i := 0; b.Loop(); i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := accountKeeper.NewAccountWithAddress(ctx, addr)
		accountKeeper.SetAccount(ctx, acc)
	}
}
