package auth

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BenchmarkAccountMapperGetAccountFound(b *testing.B) {
	input := setupTestInput()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := input.ak.NewAccountWithAddress(input.ctx, addr)
		input.ak.SetAccount(input.ctx, acc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		input.ak.GetAccount(input.ctx, sdk.AccAddress(arr))
	}
}

func BenchmarkAccountMapperGetAccountFoundWithCoins(b *testing.B) {
	input := setupTestInput()
	coins := sdk.Coins{
		sdk.NewCoin("ltc", sdk.NewInt(1000)),
		sdk.NewCoin("btc", sdk.NewInt(1000)),
		sdk.NewCoin("eth", sdk.NewInt(1000)),
		sdk.NewCoin("xrp", sdk.NewInt(1000)),
		sdk.NewCoin("bch", sdk.NewInt(1000)),
		sdk.NewCoin("eos", sdk.NewInt(1000)),
	}

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := input.ak.NewAccountWithAddress(input.ctx, addr)
		acc.SetCoins(coins)
		input.ak.SetAccount(input.ctx, acc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		input.ak.GetAccount(input.ctx, sdk.AccAddress(arr))
	}
}

func BenchmarkAccountMapperSetAccount(b *testing.B) {
	input := setupTestInput()

	b.ResetTimer()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := input.ak.NewAccountWithAddress(input.ctx, addr)
		input.ak.SetAccount(input.ctx, acc)
	}
}

func BenchmarkAccountMapperSetAccountWithCoins(b *testing.B) {
	input := setupTestInput()
	coins := sdk.Coins{
		sdk.NewCoin("ltc", sdk.NewInt(1000)),
		sdk.NewCoin("btc", sdk.NewInt(1000)),
		sdk.NewCoin("eth", sdk.NewInt(1000)),
		sdk.NewCoin("xrp", sdk.NewInt(1000)),
		sdk.NewCoin("bch", sdk.NewInt(1000)),
		sdk.NewCoin("eos", sdk.NewInt(1000)),
	}

	b.ResetTimer()

	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := input.ak.NewAccountWithAddress(input.ctx, addr)
		acc.SetCoins(coins)
		input.ak.SetAccount(input.ctx, acc)
	}
}
