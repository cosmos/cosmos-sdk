package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BenchmarkAccountMapperGetAccountFound(b *testing.B) {
	app, ctx := createTestApp(false)

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		app.AccountKeeper.SetAccount(ctx, acc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		app.AccountKeeper.GetAccount(ctx, sdk.AccAddress(arr))
	}
}

func BenchmarkAccountMapperSetAccount(b *testing.B) {
	b.ReportAllocs()
	app, ctx := createTestApp(false)

	b.ResetTimer()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		app.AccountKeeper.SetAccount(ctx, acc)
	}
}
