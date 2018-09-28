package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlusDecCoin(t *testing.T) {
	decCoinA1 := DecCoin("A", sdk.NewDecWithPrec(11, 1))
	decCoinA2 := DecCoin("A", sdk.NewDecWithPrec(22, 1))
	decCoinB1 := DecCoin("B", sdk.NewDecWithPrec(11, 1))
	decCoinC1 := DecCoin("C", sdk.NewDecWithPrec(11, 1))
	decCoinCn4 := DecCoin("C", sdk.NewDecWithPrec(-44, 1))
	decCoinC5 := DecCoin("C", sdk.NewDecWithPrec(55, 1))

	cases := []struct {
		inputOne DecCoin
		inputTwo DecCoin
		expected DecCoin
	}{
		{decCoinA1, decCoinA1, decCoinA2},
		{decCoinA1, decCoinB1, decCoinA1},
		{decCoinCn4, decCoinC5, decCoinC1},
	}
	for tcIndex, tc := range cases {
		res := tc.inputOne.Plus(tc.inputTwo)
		require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
	}
}

func TestPlusCoins(t *testing.T) {
	one := sdk.NewDec(1)
	zero := sdk.NewDec(0)
	negone := sdk.NewDec(-1)
	two := sdk.NewDec(2)

	cases := []struct {
		inputOne DecCoins
		inputTwo DecCoins
		expected DecCoins
	}{
		{DecCoins{{"A", one}, {"B", one}}, DecCoins{{"A", one}, {"B", one}}, DecCoins{{"A", two}, {"B", two}}},
		{DecCoins{{"A", zero}, {"B", one}}, DecCoins{{"A", zero}, {"B", zero}}, DecCoins{{"B", one}}},
		{DecCoins{{"A", zero}, {"B", zero}}, DecCoins{{"A", zero}, {"B", zero}}, DecCoins(nil)},
		{DecCoins{{"A", one}, {"B", zero}}, DecCoins{{"A", negone}, {"B", zero}}, DecCoins(nil)},
		{DecCoins{{"A", negone}, {"B", zero}}, DecCoins{{"A", zero}, {"B", zero}}, DecCoins{{"A", negone}}},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.Plus(tc.inputTwo)
		assert.True(t, res.IsValid())
		require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
	}
}
