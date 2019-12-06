package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

func TestRandSubsetCoins(t *testing.T) {
	tests := []struct {
		name  string
		r     *rand.Rand
		coins sdk.Coins
	}{
		{"seed=1", rand.New(rand.NewSource(1)), mustParseCoins("100stake,2testtoken")},
		{"seed=50", rand.New(rand.NewSource(50)), mustParseCoins("100stake,2testtoken")},
		{"seed=99", rand.New(rand.NewSource(99)), mustParseCoins("100stake,2testtoken")},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := simulation.RandSubsetCoins(tt.r, tt.coins)
			gotStringRep := got.String()
			sortedStringRep := got.Sort().String()
			require.Equal(t, gotStringRep, sortedStringRep)
		})
	}
}

func mustParseCoins(s string) sdk.Coins {
	coins, err := sdk.ParseCoins(s)
	if err != nil {
		panic(err)
	}
	return coins
}
