package auth

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestAccountMapperGetSet(t *testing.T) {
	input := setupTestInput()
	addr := sdk.AccAddress([]byte("some-address"))

	// no account before its created
	acc := input.ak.GetAccount(input.ctx, addr)
	require.Nil(t, acc)

	// create account and check default values
	acc = input.ak.NewAccountWithAddress(input.ctx, addr)
	require.NotNil(t, acc)
	require.Equal(t, addr, acc.GetAddress())
	require.EqualValues(t, nil, acc.GetPubKey())
	require.EqualValues(t, 0, acc.GetSequence())

	// NewAccount doesn't call Set, so it's still nil
	require.Nil(t, input.ak.GetAccount(input.ctx, addr))

	// set some values on the account and save it
	newSequence := uint64(20)
	acc.SetSequence(newSequence)
	input.ak.SetAccount(input.ctx, acc)

	// check the new values
	acc = input.ak.GetAccount(input.ctx, addr)
	require.NotNil(t, acc)
	require.Equal(t, newSequence, acc.GetSequence())
}

func TestAccountMapperRemoveAccount(t *testing.T) {
	input := setupTestInput()
	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	// create accounts
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)
	acc2 := input.ak.NewAccountWithAddress(input.ctx, addr2)

	accSeq1 := uint64(20)
	accSeq2 := uint64(40)

	acc1.SetSequence(accSeq1)
	acc2.SetSequence(accSeq2)
	input.ak.SetAccount(input.ctx, acc1)
	input.ak.SetAccount(input.ctx, acc2)

	acc1 = input.ak.GetAccount(input.ctx, addr1)
	require.NotNil(t, acc1)
	require.Equal(t, accSeq1, acc1.GetSequence())

	// remove one account
	input.ak.RemoveAccount(input.ctx, acc1)
	acc1 = input.ak.GetAccount(input.ctx, addr1)
	require.Nil(t, acc1)

	acc2 = input.ak.GetAccount(input.ctx, addr2)
	require.NotNil(t, acc2)
	require.Equal(t, accSeq2, acc2.GetSequence())
}

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
		acc := input.ak.NewAccountWithAddress(input.ctx, addr)
		acc.SetCoins(coins)
		input.ak.SetAccount(input.ctx, acc)
	}
}
