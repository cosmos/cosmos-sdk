package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNewDecCoin(t *testing.T) {
	require.NotPanics(t, func() {
		sdk.NewInt64DecCoin(testDenom1, 5)
	})
	require.NotPanics(t, func() {
		sdk.NewInt64DecCoin(testDenom1, 0)
	})
	require.NotPanics(t, func() {
		sdk.NewInt64DecCoin(strings.ToUpper(testDenom1), 5)
	})
	require.Panics(t, func() {
		sdk.NewInt64DecCoin(testDenom1, -5)
	})
}

func TestNewDecCoinFromDec(t *testing.T) {
	require.NotPanics(t, func() {
		sdk.NewDecCoinFromDec(testDenom1, sdk.NewDec(5))
	})
	require.NotPanics(t, func() {
		sdk.NewDecCoinFromDec(testDenom1, sdk.ZeroDec())
	})
	require.NotPanics(t, func() {
		sdk.NewDecCoinFromDec(strings.ToUpper(testDenom1), sdk.NewDec(5))
	})
	require.Panics(t, func() {
		sdk.NewDecCoinFromDec(testDenom1, sdk.NewDec(-5))
	})
}

func TestNewDecCoinFromCoin(t *testing.T) {
	require.NotPanics(t, func() {
		sdk.NewDecCoinFromCoin(sdk.Coin{testDenom1, sdk.NewInt(5)})
	})
	require.NotPanics(t, func() {
		sdk.NewDecCoinFromCoin(sdk.Coin{testDenom1, sdk.NewInt(0)})
	})
	require.NotPanics(t, func() {
		sdk.NewDecCoinFromCoin(sdk.Coin{strings.ToUpper(testDenom1), sdk.NewInt(5)})
	})
	require.Panics(t, func() {
		sdk.NewDecCoinFromCoin(sdk.Coin{testDenom1, sdk.NewInt(-5)})
	})
}

func TestDecCoinIsPositive(t *testing.T) {
	dc := sdk.NewInt64DecCoin(testDenom1, 5)
	require.True(t, dc.IsPositive())

	dc = sdk.NewInt64DecCoin(testDenom1, 0)
	require.False(t, dc.IsPositive())
}

func TestAddDecCoin(t *testing.T) {
	decCoinA1 := sdk.NewDecCoinFromDec(testDenom1, sdk.NewDecWithPrec(11, 1))
	decCoinA2 := sdk.NewDecCoinFromDec(testDenom1, sdk.NewDecWithPrec(22, 1))
	decCoinB1 := sdk.NewDecCoinFromDec(testDenom2, sdk.NewDecWithPrec(11, 1))

	// regular add
	res := decCoinA1.Add(decCoinA1)
	require.Equal(t, decCoinA2, res, "sum of coins is incorrect")

	// bad denom add
	require.Panics(t, func() {
		decCoinA1.Add(decCoinB1)
	}, "expected panic on sum of different denoms")
}

func TestAddDecCoins(t *testing.T) {
	one := sdk.NewDec(1)
	zero := sdk.NewDec(0)
	two := sdk.NewDec(2)

	cases := []struct {
		inputOne sdk.DecCoins
		inputTwo sdk.DecCoins
		expected sdk.DecCoins
	}{
		{sdk.DecCoins{{testDenom1, one}, {testDenom2, one}}, sdk.DecCoins{{testDenom1, one}, {testDenom2, one}}, sdk.DecCoins{{testDenom1, two}, {testDenom2, two}}},
		{sdk.DecCoins{{testDenom1, zero}, {testDenom2, one}}, sdk.DecCoins{{testDenom1, zero}, {testDenom2, zero}}, sdk.DecCoins{{testDenom2, one}}},
		{sdk.DecCoins{{testDenom1, zero}, {testDenom2, zero}}, sdk.DecCoins{{testDenom1, zero}, {testDenom2, zero}}, sdk.DecCoins(nil)},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.Add(tc.inputTwo...)
		require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
	}
}

func TestFilteredZeroDecCoins(t *testing.T) {
	cases := []struct {
		name     string
		input    sdk.DecCoins
		original string
		expected string
	}{
		{
			name: "all greater than zero",
			input: sdk.DecCoins{
				{"testa", sdk.NewDec(1)},
				{"testb", sdk.NewDec(2)},
				{"testc", sdk.NewDec(3)},
				{"testd", sdk.NewDec(4)},
				{"teste", sdk.NewDec(5)},
			},
			original: "1.000000000000000000testa,2.000000000000000000testb,3.000000000000000000testc,4.000000000000000000testd,5.000000000000000000teste",
			expected: "1.000000000000000000testa,2.000000000000000000testb,3.000000000000000000testc,4.000000000000000000testd,5.000000000000000000teste",
		},
		{
			name: "zero coin in middle",
			input: sdk.DecCoins{
				{"testa", sdk.NewDec(1)},
				{"testb", sdk.NewDec(2)},
				{"testc", sdk.NewDec(0)},
				{"testd", sdk.NewDec(4)},
				{"teste", sdk.NewDec(5)},
			},
			original: "1.000000000000000000testa,2.000000000000000000testb,0.000000000000000000testc,4.000000000000000000testd,5.000000000000000000teste",
			expected: "1.000000000000000000testa,2.000000000000000000testb,4.000000000000000000testd,5.000000000000000000teste",
		},
		{
			name: "zero coin end (unordered)",
			input: sdk.DecCoins{
				{"teste", sdk.NewDec(5)},
				{"testc", sdk.NewDec(3)},
				{"testa", sdk.NewDec(1)},
				{"testd", sdk.NewDec(4)},
				{"testb", sdk.NewDec(0)},
			},
			original: "5.000000000000000000teste,3.000000000000000000testc,1.000000000000000000testa,4.000000000000000000testd,0.000000000000000000testb",
			expected: "1.000000000000000000testa,3.000000000000000000testc,4.000000000000000000testd,5.000000000000000000teste",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			undertest := sdk.NewDecCoins(tt.input...)
			require.Equal(t, tt.expected, undertest.String(), "NewDecCoins must return expected results")
			require.Equal(t, tt.original, tt.input.String(), "input must be unmodified and match original")
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		coin       sdk.DecCoin
		expectPass bool
		msg        string
	}{
		{
			sdk.NewDecCoin("mytoken", sdk.NewInt(10)),
			true,
			"valid coins should have passed",
		},
		{
			sdk.DecCoin{Denom: "BTC", Amount: sdk.NewDec(10)},
			true,
			"valid uppercase denom",
		},
		{
			sdk.DecCoin{Denom: "Bitcoin", Amount: sdk.NewDec(10)},
			true,
			"valid mixed case denom",
		},
		{
			sdk.DecCoin{Denom: "btc", Amount: sdk.NewDec(-10)},
			false,
			"negative amount",
		},
	}

	for _, tc := range tests {
		tc := tc
		if tc.expectPass {
			require.True(t, tc.coin.IsValid(), tc.msg)
		} else {
			require.False(t, tc.coin.IsValid(), tc.msg)
		}
	}
}

func TestSubDecCoin(t *testing.T) {
	tests := []struct {
		coin       sdk.DecCoin
		expectPass bool
		msg        string
	}{
		{
			sdk.NewDecCoin("mytoken", sdk.NewInt(20)),
			true,
			"valid coins should have passed",
		},
		{
			sdk.NewDecCoin("othertoken", sdk.NewInt(20)),
			false,
			"denom mismatch",
		},
		{
			sdk.NewDecCoin("mytoken", sdk.NewInt(9)),
			false,
			"negative amount",
		},
	}

	decCoin := sdk.NewDecCoin("mytoken", sdk.NewInt(10))

	for _, tc := range tests {
		tc := tc
		if tc.expectPass {
			equal := tc.coin.Sub(decCoin)
			require.Equal(t, equal, decCoin, tc.msg)
		} else {
			require.Panics(t, func() { tc.coin.Sub(decCoin) }, tc.msg)
		}
	}
}

func TestSubDecCoins(t *testing.T) {
	tests := []struct {
		coins      sdk.DecCoins
		expectPass bool
		msg        string
	}{
		{
			sdk.NewDecCoinsFromCoins(sdk.NewCoin("mytoken", sdk.NewInt(10)), sdk.NewCoin("btc", sdk.NewInt(20)), sdk.NewCoin("eth", sdk.NewInt(30))),
			true,
			"sorted coins should have passed",
		},
		{
			sdk.DecCoins{sdk.NewDecCoin("mytoken", sdk.NewInt(10)), sdk.NewDecCoin("btc", sdk.NewInt(20)), sdk.NewDecCoin("eth", sdk.NewInt(30))},
			false,
			"unorted coins should panic",
		},
		{
			sdk.DecCoins{sdk.DecCoin{Denom: "BTC", Amount: sdk.NewDec(10)}, sdk.NewDecCoin("eth", sdk.NewInt(15)), sdk.NewDecCoin("mytoken", sdk.NewInt(5))},
			false,
			"invalid denoms",
		},
	}

	decCoins := sdk.NewDecCoinsFromCoins(sdk.NewCoin("btc", sdk.NewInt(10)), sdk.NewCoin("eth", sdk.NewInt(15)), sdk.NewCoin("mytoken", sdk.NewInt(5)))

	for _, tc := range tests {
		tc := tc
		if tc.expectPass {
			equal := tc.coins.Sub(decCoins)
			require.Equal(t, equal, decCoins, tc.msg)
		} else {
			require.Panics(t, func() { tc.coins.Sub(decCoins) }, tc.msg)
		}
	}
}

func TestSortDecCoins(t *testing.T) {
	good := sdk.DecCoins{
		sdk.NewInt64DecCoin("gas", 1),
		sdk.NewInt64DecCoin("mineral", 1),
		sdk.NewInt64DecCoin("tree", 1),
	}
	empty := sdk.DecCoins{
		sdk.NewInt64DecCoin("gold", 0),
	}
	badSort1 := sdk.DecCoins{
		sdk.NewInt64DecCoin("tree", 1),
		sdk.NewInt64DecCoin("gas", 1),
		sdk.NewInt64DecCoin("mineral", 1),
	}
	badSort2 := sdk.DecCoins{ // both are after the first one, but the second and third are in the wrong order
		sdk.NewInt64DecCoin("gas", 1),
		sdk.NewInt64DecCoin("tree", 1),
		sdk.NewInt64DecCoin("mineral", 1),
	}
	badAmt := sdk.DecCoins{
		sdk.NewInt64DecCoin("gas", 1),
		sdk.NewInt64DecCoin("tree", 0),
		sdk.NewInt64DecCoin("mineral", 1),
	}
	dup := sdk.DecCoins{
		sdk.NewInt64DecCoin("gas", 1),
		sdk.NewInt64DecCoin("gas", 1),
		sdk.NewInt64DecCoin("mineral", 1),
	}

	cases := []struct {
		name          string
		coins         sdk.DecCoins
		before, after bool // valid before/after sort
	}{
		{"valid coins", good, true, true},
		{"empty coins", empty, false, false},
		{"unsorted coins (1)", badSort1, false, true},
		{"unsorted coins (2)", badSort2, false, true},
		{"zero amount coins", badAmt, false, false},
		{"duplicate coins", dup, false, false},
	}

	for _, tc := range cases {
		require.Equal(t, tc.before, tc.coins.IsValid(), "coin validity is incorrect before sorting; %s", tc.name)
		tc.coins.Sort()
		require.Equal(t, tc.after, tc.coins.IsValid(), "coin validity is incorrect after sorting;  %s", tc.name)
	}
}

func TestDecCoinsValidate(t *testing.T) {
	testCases := []struct {
		input        sdk.DecCoins
		expectedPass bool
	}{
		{sdk.DecCoins{}, true},
		{sdk.DecCoins{sdk.DecCoin{testDenom1, sdk.NewDec(5)}}, true},
		{sdk.DecCoins{sdk.DecCoin{testDenom1, sdk.NewDec(5)}, sdk.DecCoin{testDenom2, sdk.NewDec(100000)}}, true},
		{sdk.DecCoins{sdk.DecCoin{testDenom1, sdk.NewDec(-5)}}, false},
		{sdk.DecCoins{sdk.DecCoin{"BTC", sdk.NewDec(5)}}, true},
		{sdk.DecCoins{sdk.DecCoin{"0BTC", sdk.NewDec(5)}}, false},
		{sdk.DecCoins{sdk.DecCoin{testDenom1, sdk.NewDec(5)}, sdk.DecCoin{"B", sdk.NewDec(100000)}}, false},
		{sdk.DecCoins{sdk.DecCoin{testDenom1, sdk.NewDec(5)}, sdk.DecCoin{testDenom2, sdk.NewDec(-100000)}}, false},
		{sdk.DecCoins{sdk.DecCoin{testDenom1, sdk.NewDec(-5)}, sdk.DecCoin{testDenom2, sdk.NewDec(100000)}}, false},
		{sdk.DecCoins{sdk.DecCoin{"BTC", sdk.NewDec(5)}, sdk.DecCoin{testDenom2, sdk.NewDec(100000)}}, true},
		{sdk.DecCoins{sdk.DecCoin{"0BTC", sdk.NewDec(5)}, sdk.DecCoin{testDenom2, sdk.NewDec(100000)}}, false},
	}

	for i, tc := range testCases {
		err := tc.input.Validate()
		if tc.expectedPass {
			require.NoError(t, err, "unexpected result for test case #%d, input: %v", i, tc.input)
		} else {
			require.Error(t, err, "unexpected result for test case #%d, input: %v", i, tc.input)
		}
	}
}

func TestParseDecCoins(t *testing.T) {
	testCases := []struct {
		input          string
		expectedResult sdk.DecCoins
		expectedErr    bool
	}{
		{"", nil, false},
		{"4stake", nil, true},
		{"5.5atom,4stake", nil, true},
		{"0.0stake", sdk.DecCoins{}, false}, // remove zero coins
		{"10.0btc,1.0atom,20.0btc", nil, true},
		{
			"0.004STAKE",
			sdk.DecCoins{sdk.NewDecCoinFromDec("STAKE", sdk.NewDecWithPrec(4000000000000000, sdk.Precision))},
			false,
		},
		{
			"0.004stake",
			sdk.DecCoins{sdk.NewDecCoinFromDec("stake", sdk.NewDecWithPrec(4000000000000000, sdk.Precision))},
			false,
		},
		{
			"5.04atom,0.004stake",
			sdk.DecCoins{
				sdk.NewDecCoinFromDec("atom", sdk.NewDecWithPrec(5040000000000000000, sdk.Precision)),
				sdk.NewDecCoinFromDec("stake", sdk.NewDecWithPrec(4000000000000000, sdk.Precision)),
			},
			false,
		},
		{"0.0stake,0.004stake,5.04atom", // remove zero coins
			sdk.DecCoins{
				sdk.NewDecCoinFromDec("atom", sdk.NewDecWithPrec(5040000000000000000, sdk.Precision)),
				sdk.NewDecCoinFromDec("stake", sdk.NewDecWithPrec(4000000000000000, sdk.Precision)),
			},
			false,
		},
	}

	for i, tc := range testCases {
		res, err := sdk.ParseDecCoins(tc.input)
		if tc.expectedErr {
			require.Error(t, err, "expected error for test case #%d, input: %v", i, tc.input)
		} else {
			require.NoError(t, err, "unexpected error for test case #%d, input: %v", i, tc.input)
			require.Equal(t, tc.expectedResult, res, "unexpected result for test case #%d, input: %v", i, tc.input)
		}
	}
}

func TestDecCoinsString(t *testing.T) {
	testCases := []struct {
		input    sdk.DecCoins
		expected string
	}{
		{sdk.DecCoins{}, ""},
		{
			sdk.DecCoins{
				sdk.NewDecCoinFromDec("atom", sdk.NewDecWithPrec(5040000000000000000, sdk.Precision)),
				sdk.NewDecCoinFromDec("stake", sdk.NewDecWithPrec(4000000000000000, sdk.Precision)),
			},
			"5.040000000000000000atom,0.004000000000000000stake",
		},
	}

	for i, tc := range testCases {
		out := tc.input.String()
		require.Equal(t, tc.expected, out, "unexpected result for test case #%d, input: %v", i, tc.input)
	}
}

func TestDecCoinsIntersect(t *testing.T) {
	testCases := []struct {
		input1         string
		input2         string
		expectedResult string
	}{
		{"", "", ""},
		{"1.0stake", "", ""},
		{"1.0stake", "1.0stake", "1.0stake"},
		{"", "1.0stake", ""},
		{"1.0stake", "", ""},
		{"2.0stake,1.0trope", "1.9stake", "1.9stake"},
		{"2.0stake,1.0trope", "2.1stake", "2.0stake"},
		{"2.0stake,1.0trope", "0.9trope", "0.9trope"},
		{"2.0stake,1.0trope", "1.9stake,0.9trope", "1.9stake,0.9trope"},
		{"2.0stake,1.0trope", "1.9stake,0.9trope,20.0other", "1.9stake,0.9trope"},
		{"2.0stake,1.0trope", "1.0other", ""},
	}

	for i, tc := range testCases {
		in1, err := sdk.ParseDecCoins(tc.input1)
		require.NoError(t, err, "unexpected parse error in %v", i)
		in2, err := sdk.ParseDecCoins(tc.input2)
		require.NoError(t, err, "unexpected parse error in %v", i)
		exr, err := sdk.ParseDecCoins(tc.expectedResult)
		require.NoError(t, err, "unexpected parse error in %v", i)

		require.True(t, in1.Intersect(in2).IsEqual(exr), "in1.cap(in2) != exr in %v", i)
	}
}

func TestDecCoinsTruncateDecimal(t *testing.T) {
	decCoinA := sdk.NewDecCoinFromDec("bar", sdk.MustNewDecFromStr("5.41"))
	decCoinB := sdk.NewDecCoinFromDec("foo", sdk.MustNewDecFromStr("6.00"))

	testCases := []struct {
		input          sdk.DecCoins
		truncatedCoins sdk.Coins
		changeCoins    sdk.DecCoins
	}{
		{sdk.DecCoins{}, sdk.Coins(nil), sdk.DecCoins(nil)},
		{
			sdk.DecCoins{decCoinA, decCoinB},
			sdk.Coins{sdk.NewInt64Coin(decCoinA.Denom, 5), sdk.NewInt64Coin(decCoinB.Denom, 6)},
			sdk.DecCoins{sdk.NewDecCoinFromDec(decCoinA.Denom, sdk.MustNewDecFromStr("0.41"))},
		},
		{
			sdk.DecCoins{decCoinB},
			sdk.Coins{sdk.NewInt64Coin(decCoinB.Denom, 6)},
			sdk.DecCoins(nil),
		},
	}

	for i, tc := range testCases {
		truncatedCoins, changeCoins := tc.input.TruncateDecimal()
		require.Equal(
			t, tc.truncatedCoins, truncatedCoins,
			"unexpected truncated coins; tc #%d, input: %s", i, tc.input,
		)
		require.Equal(
			t, tc.changeCoins, changeCoins,
			"unexpected change coins; tc #%d, input: %s", i, tc.input,
		)
	}
}

func TestDecCoinsQuoDecTruncate(t *testing.T) {
	x := sdk.MustNewDecFromStr("1.00")
	y := sdk.MustNewDecFromStr("10000000000000000000.00")

	testCases := []struct {
		coins  sdk.DecCoins
		input  sdk.Dec
		result sdk.DecCoins
		panics bool
	}{
		{sdk.DecCoins{}, sdk.ZeroDec(), sdk.DecCoins(nil), true},
		{sdk.DecCoins{sdk.NewDecCoinFromDec("foo", x)}, y, sdk.DecCoins(nil), false},
		{sdk.DecCoins{sdk.NewInt64DecCoin("foo", 5)}, sdk.NewDec(2), sdk.DecCoins{sdk.NewDecCoinFromDec("foo", sdk.MustNewDecFromStr("2.5"))}, false},
	}

	for i, tc := range testCases {
		tc := tc
		if tc.panics {
			require.Panics(t, func() { tc.coins.QuoDecTruncate(tc.input) })
		} else {
			res := tc.coins.QuoDecTruncate(tc.input)
			require.Equal(t, tc.result, res, "unexpected result; tc #%d, coins: %s, input: %s", i, tc.coins, tc.input)
		}
	}
}

func TestNewDecCoinsWithIsValid(t *testing.T) {
	fake1 := append(sdk.NewDecCoins(sdk.NewDecCoin("mytoken", sdk.NewInt(10))), sdk.DecCoin{Denom: "10BTC", Amount: sdk.NewDec(10)})
	fake2 := append(sdk.NewDecCoins(sdk.NewDecCoin("mytoken", sdk.NewInt(10))), sdk.DecCoin{Denom: "BTC", Amount: sdk.NewDec(-10)})

	tests := []struct {
		coin       sdk.DecCoins
		expectPass bool
		msg        string
	}{
		{
			sdk.NewDecCoins(sdk.NewDecCoin("mytoken", sdk.NewInt(10))),
			true,
			"valid coins should have passed",
		},
		{
			fake1,
			false,
			"invalid denoms",
		},
		{
			fake2,
			false,
			"negative amount",
		},
	}

	for _, tc := range tests {
		tc := tc
		if tc.expectPass {
			require.True(t, tc.coin.IsValid(), tc.msg)
		} else {
			require.False(t, tc.coin.IsValid(), tc.msg)
		}
	}
}

func TestDecCoins_AddDecCoinWithIsValid(t *testing.T) {
	lengthTestDecCoins := sdk.NewDecCoins().Add(sdk.NewDecCoin("mytoken", sdk.NewInt(10))).Add(sdk.DecCoin{Denom: "BTC", Amount: sdk.NewDec(10)})
	require.Equal(t, 2, len(lengthTestDecCoins), "should be 2")

	tests := []struct {
		coin       sdk.DecCoins
		expectPass bool
		msg        string
	}{
		{
			sdk.NewDecCoins().Add(sdk.NewDecCoin("mytoken", sdk.NewInt(10))),
			true,
			"valid coins should have passed",
		},
		{
			sdk.NewDecCoins().Add(sdk.NewDecCoin("mytoken", sdk.NewInt(10))).Add(sdk.DecCoin{Denom: "0BTC", Amount: sdk.NewDec(10)}),
			false,
			"invalid denoms",
		},
		{
			sdk.NewDecCoins().Add(sdk.NewDecCoin("mytoken", sdk.NewInt(10))).Add(sdk.DecCoin{Denom: "BTC", Amount: sdk.NewDec(-10)}),
			false,
			"negative amount",
		},
	}

	for _, tc := range tests {
		tc := tc
		if tc.expectPass {
			require.True(t, tc.coin.IsValid(), tc.msg)
		} else {
			require.False(t, tc.coin.IsValid(), tc.msg)
		}
	}
}
