package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

func TestRandomAccounts(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(time.Now().Unix()))
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"0-accounts", 0, 0},
		{"0-accounts", 1, 1},
		{"0-accounts", 1_000, 1_000},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := simulation.RandomAccounts(r, tt.n)
			require.Equal(t, tt.want, len(got))
			if tt.n == 0 {
				return
			}
			acc, i := simulation.RandomAcc(r, got)
			require.True(t, acc.Equals(got[i]))
			accFound, found := simulation.FindAccount(got, acc.Address)
			require.True(t, found)
			require.True(t, accFound.Equals(acc))
		})
	}
}

func TestFindAccountEmptySlice(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(time.Now().Unix()))
	accs := simulation.RandomAccounts(r, 1)
	require.Equal(t, 1, len(accs))
	acc, found := simulation.FindAccount(nil, accs[0].Address)
	require.False(t, found)
	require.Nil(t, acc.Address)
	require.Nil(t, acc.PrivKey)
	require.Nil(t, acc.PubKey)
}

func TestRandomFees(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(time.Now().Unix()))
	tests := []struct {
		name           string
		spendableCoins sdk.Coins
		wantEmpty      bool
		wantErr        bool
	}{
		{"0 coins", sdk.Coins{}, true, false},
		{"0 coins", sdk.NewCoins(sdk.NewInt64Coin("aaa", 10), sdk.NewInt64Coin("bbb", 5)), false, false},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := simulation.RandomFees(r, sdk.Context{}, tt.spendableCoins)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomFees() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.wantEmpty, got.Empty())
		})
	}
}
