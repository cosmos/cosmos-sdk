package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BenchmarkAccountMapperGetAccountFound(b *testing.B) {
	input := SetupTestInput()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := input.AccountKeeper.NewAccountWithAddress(input.Ctx, addr)
		input.AccountKeeper.SetAccount(input.Ctx, acc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		input.AccountKeeper.GetAccount(input.Ctx, sdk.AccAddress(arr))
	}
}

func BenchmarkAccountMapperGetAccountFoundWithCoins(b *testing.B) {
	input := SetupTestInput()
	coins := sdk.Coins{
		sdk.NewCoin("LTC", sdk.NewInt(1000)),
		sdk.NewCoin("BTC", sdk.NewInt(1000)),
		sdk.NewCoin("ETH", sdk.NewInt(1000)),
		sdk.NewCoin("XRP", sdk.NewInt(1000)),
		sdk.NewCoin("BCH", sdk.NewInt(1000)),
		sdk.NewCoin("EOS", sdk.NewInt(1000)),
	}

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := input.AccountKeeper.NewAccountWithAddress(input.Ctx, addr)
		acc.SetCoins(coins)
		input.AccountKeeper.SetAccount(input.Ctx, acc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		input.AccountKeeper.GetAccount(input.Ctx, sdk.AccAddress(arr))
	}
}

func BenchmarkAccountMapperSetAccount(b *testing.B) {
	input := SetupTestInput()

	b.ResetTimer()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := input.AccountKeeper.NewAccountWithAddress(input.Ctx, addr)
		input.AccountKeeper.SetAccount(input.Ctx, acc)
	}
}

func BenchmarkAccountMapperSetAccountWithCoins(b *testing.B) {
	input := SetupTestInput()
	coins := sdk.Coins{
		sdk.NewCoin("LTC", sdk.NewInt(1000)),
		sdk.NewCoin("BTC", sdk.NewInt(1000)),
		sdk.NewCoin("ETH", sdk.NewInt(1000)),
		sdk.NewCoin("XRP", sdk.NewInt(1000)),
		sdk.NewCoin("BCH", sdk.NewInt(1000)),
		sdk.NewCoin("EOS", sdk.NewInt(1000)),
	}

	b.ResetTimer()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := input.AccountKeeper.NewAccountWithAddress(input.Ctx, addr)
		acc.SetCoins(coins)
		input.AccountKeeper.SetAccount(input.Ctx, acc)
	}
}
