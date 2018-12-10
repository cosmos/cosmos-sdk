package auth

import (
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

func BenchmarkAccountMapperGetAccountFoundWithCoins(b *testing.B) {
	ms, capKey, _ := setupMultiStore()
	cdc := codec.New()
	RegisterBaseAccount(cdc)

	// make context and mapper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	mapper := NewAccountKeeper(cdc, capKey, ProtoBaseAccount)

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
		acc := mapper.NewAccountWithAddress(ctx, addr)
		acc.SetCoins(coins)
		mapper.SetAccount(ctx, acc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		mapper.GetAccount(ctx, sdk.AccAddress(arr))
	}
}

func BenchmarkAccountMapperSetAccount(b *testing.B) {
	ms, capKey, _ := setupMultiStore()
	cdc := codec.New()
	RegisterBaseAccount(cdc)

	// make context and mapper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	mapper := NewAccountKeeper(cdc, capKey, ProtoBaseAccount)

	b.ResetTimer()
	// assumes b.N < 2**24
	for i := 0; i < b.N; i++ {
		arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
		addr := sdk.AccAddress(arr)
		acc := mapper.NewAccountWithAddress(ctx, addr)
		mapper.SetAccount(ctx, acc)
	}
}

func BenchmarkAccountMapperSetAccountWithCoins(b *testing.B) {
	ms, capKey, _ := setupMultiStore()
	cdc := codec.New()
	RegisterBaseAccount(cdc)

	// make context and mapper
	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	mapper := NewAccountKeeper(cdc, capKey, ProtoBaseAccount)

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
		acc := mapper.NewAccountWithAddress(ctx, addr)
		acc.SetCoins(coins)
		mapper.SetAccount(ctx, acc)
	}
}
