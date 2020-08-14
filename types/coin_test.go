package types

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	testDenom1 = "atom"
	testDenom2 = "muon"
)

// ----------------------------------------------------------------------------
// Coin tests

func TestCoin(t *testing.T) {
	require.Panics(t, func() { NewInt64Coin(testDenom1, -1) })
	require.Panics(t, func() { NewCoin(testDenom1, NewInt(-1)) })
	require.Equal(t, NewInt(10), NewInt64Coin(strings.ToUpper(testDenom1), 10).Amount)
	require.Equal(t, NewInt(10), NewCoin(strings.ToUpper(testDenom1), NewInt(10)).Amount)
	require.Equal(t, NewInt(5), NewInt64Coin(testDenom1, 5).Amount)
	require.Equal(t, NewInt(5), NewCoin(testDenom1, NewInt(5)).Amount)
}

func TestCoin_String(t *testing.T) {
	coin := NewCoin(testDenom1, NewInt(10))
	require.Equal(t, fmt.Sprintf("10%s", testDenom1), coin.String())
}

func TestIsEqualCoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
		panics   bool
	}{
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), true, false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom2, 1), false, true},
		{NewInt64Coin("stake", 1), NewInt64Coin("stake", 10), false, false},
	}

	for tcIndex, tc := range cases {
		tc := tc
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.IsEqual(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsEqual(tc.inputTwo)
			require.Equal(t, tc.expected, res, "coin equality relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestCoinIsValid(t *testing.T) {
	loremIpsum := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam viverra dui vel nulla aliquet, non dictum elit aliquam. Proin consequat leo in consectetur mattis. Phasellus eget odio luctus, rutrum dolor at, venenatis ante. Praesent metus erat, sodales vitae sagittis eget, commodo non ipsum. Duis eget urna quis erat mattis pulvinar. Vivamus egestas imperdiet sem, porttitor hendrerit lorem pulvinar in. Vivamus laoreet sapien eget libero euismod tristique. Suspendisse tincidunt nulla quis luctus mattis.
	Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Sed id turpis at erat placerat fermentum id sed sapien. Fusce mattis enim id nulla viverra, eget placerat eros aliquet. Nunc fringilla urna ac condimentum ultricies. Praesent in eros ac neque fringilla sodales. Donec ut venenatis eros. Quisque iaculis lectus neque, a varius sem ullamcorper nec. Cras tincidunt dignissim libero nec volutpat. Donec molestie enim sed metus venenatis, quis elementum sem varius. Curabitur eu venenatis nulla.
	Cras sit amet ligula vel turpis placerat sollicitudin. Nunc massa odio, eleifend id lacus nec, ultricies elementum arcu. Donec imperdiet nulla lacus, a venenatis lacus fermentum nec. Proin vestibulum dolor enim, vitae posuere velit aliquet non. Suspendisse pharetra condimentum nunc tincidunt viverra. Etiam posuere, ligula ut maximus congue, mauris orci consectetur velit, vel finibus eros metus non tellus. Nullam et dictum metus. Aliquam maximus fermentum mauris elementum aliquet. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Etiam dapibus lectus sed tellus rutrum tincidunt. Nulla at dolor sem. Ut non dictum arcu, eget congue sem.`

	loremIpsum = strings.ReplaceAll(loremIpsum, " ", "")
	loremIpsum = strings.ReplaceAll(loremIpsum, ".", "")
	loremIpsum = strings.ReplaceAll(loremIpsum, ",", "")

	cases := []struct {
		coin       Coin
		expectPass bool
	}{
		{Coin{testDenom1, NewInt(-1)}, false},
		{Coin{testDenom1, NewInt(0)}, true},
		{Coin{testDenom1, OneInt()}, true},
		{Coin{"Atom", OneInt()}, true},
		{Coin{"ATOM", OneInt()}, true},
		{Coin{"a", OneInt()}, false},
		{Coin{loremIpsum, OneInt()}, false},
		{Coin{"ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2", OneInt()}, true},
		{Coin{"atOm", OneInt()}, true},
		{Coin{"     ", OneInt()}, false},
	}

	for i, tc := range cases {
		require.Equal(t, tc.expectPass, tc.coin.IsValid(), "unexpected result for IsValid, tc #%d", i)
	}
}

func TestAddCoin(t *testing.T) {
	cases := []struct {
		inputOne    Coin
		inputTwo    Coin
		expected    Coin
		shouldPanic bool
	}{
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 2), false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom1, 1), false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom2, 1), NewInt64Coin(testDenom1, 1), true},
	}

	for tcIndex, tc := range cases {
		tc := tc
		if tc.shouldPanic {
			require.Panics(t, func() { tc.inputOne.Add(tc.inputTwo) })
		} else {
			res := tc.inputOne.Add(tc.inputTwo)
			require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestSubCoin(t *testing.T) {
	cases := []struct {
		inputOne    Coin
		inputTwo    Coin
		expected    Coin
		shouldPanic bool
	}{
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom2, 1), NewInt64Coin(testDenom1, 1), true},
		{NewInt64Coin(testDenom1, 10), NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 9), false},
		{NewInt64Coin(testDenom1, 5), NewInt64Coin(testDenom1, 3), NewInt64Coin(testDenom1, 2), false},
		{NewInt64Coin(testDenom1, 5), NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom1, 5), false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 5), Coin{}, true},
	}

	for tcIndex, tc := range cases {
		tc := tc
		if tc.shouldPanic {
			require.Panics(t, func() { tc.inputOne.Sub(tc.inputTwo) })
		} else {
			res := tc.inputOne.Sub(tc.inputTwo)
			require.Equal(t, tc.expected, res, "difference of coins is incorrect, tc #%d", tcIndex)
		}
	}

	tc := struct {
		inputOne Coin
		inputTwo Coin
		expected int64
	}{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), 0}
	res := tc.inputOne.Sub(tc.inputTwo)
	require.Equal(t, tc.expected, res.Amount.Int64())
}

func TestIsGTECoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
		panics   bool
	}{
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), true, false},
		{NewInt64Coin(testDenom1, 2), NewInt64Coin(testDenom1, 1), true, false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom2, 1), false, true},
	}

	for tcIndex, tc := range cases {
		tc := tc
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.IsGTE(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsGTE(tc.inputTwo)
			require.Equal(t, tc.expected, res, "coin GTE relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestIsLTCoin(t *testing.T) {
	cases := []struct {
		inputOne Coin
		inputTwo Coin
		expected bool
		panics   bool
	}{
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), false, false},
		{NewInt64Coin(testDenom1, 2), NewInt64Coin(testDenom1, 1), false, false},
		{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1), false, true},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom2, 1), false, true},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 1), false, false},
		{NewInt64Coin(testDenom1, 1), NewInt64Coin(testDenom1, 2), true, false},
	}

	for tcIndex, tc := range cases {
		tc := tc
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.IsLT(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsLT(tc.inputTwo)
			require.Equal(t, tc.expected, res, "coin LT relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func TestCoinIsZero(t *testing.T) {
	coin := NewInt64Coin(testDenom1, 0)
	res := coin.IsZero()
	require.True(t, res)

	coin = NewInt64Coin(testDenom1, 1)
	res = coin.IsZero()
	require.False(t, res)
}

func TestFilteredZeroCoins(t *testing.T) {
	cases := []struct {
		name     string
		input    Coins
		original string
		expected string
	}{
		{
			name: "all greater than zero",
			input: Coins{
				{"testa", OneInt()},
				{"testb", NewInt(2)},
				{"testc", NewInt(3)},
				{"testd", NewInt(4)},
				{"teste", NewInt(5)},
			},
			original: "1testa,2testb,3testc,4testd,5teste",
			expected: "1testa,2testb,3testc,4testd,5teste",
		},
		{
			name: "zero coin in middle",
			input: Coins{
				{"testa", OneInt()},
				{"testb", NewInt(2)},
				{"testc", NewInt(0)},
				{"testd", NewInt(4)},
				{"teste", NewInt(5)},
			},
			original: "1testa,2testb,0testc,4testd,5teste",
			expected: "1testa,2testb,4testd,5teste",
		},
		{
			name: "zero coin end (unordered)",
			input: Coins{
				{"teste", NewInt(5)},
				{"testc", NewInt(3)},
				{"testa", OneInt()},
				{"testd", NewInt(4)},
				{"testb", NewInt(0)},
			},
			original: "5teste,3testc,1testa,4testd,0testb",
			expected: "1testa,3testc,4testd,5teste",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			undertest := NewCoins(tt.input...)
			require.Equal(t, tt.expected, undertest.String(), "NewCoins must return expected results")
			require.Equal(t, tt.original, tt.input.String(), "input must be unmodified and match original")
		})
	}
}

// ----------------------------------------------------------------------------
// Coins tests

func TestCoins_String(t *testing.T) {
	cases := []struct {
		name     string
		input    Coins
		expected string
	}{
		{
			"empty coins",
			Coins{},
			"",
		},
		{
			"single coin",
			Coins{{"tree", OneInt()}},
			"1tree",
		},
		{
			"multiple coins",
			Coins{
				{"tree", OneInt()},
				{"gas", OneInt()},
				{"mineral", OneInt()},
			},
			"1tree,1gas,1mineral",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.input.String())
		})
	}
}

func TestIsZeroCoins(t *testing.T) {
	cases := []struct {
		inputOne Coins
		expected bool
	}{
		{Coins{}, true},
		{Coins{NewInt64Coin(testDenom1, 0)}, true},
		{Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 0)}, true},
		{Coins{NewInt64Coin(testDenom1, 1)}, false},
		{Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1)}, false},
	}

	for _, tc := range cases {
		res := tc.inputOne.IsZero()
		require.Equal(t, tc.expected, res)
	}
}

func TestEqualCoins(t *testing.T) {
	cases := []struct {
		inputOne Coins
		inputTwo Coins
		expected bool
		panics   bool
	}{
		{Coins{}, Coins{}, true, false},
		{Coins{NewInt64Coin(testDenom1, 0)}, Coins{NewInt64Coin(testDenom1, 0)}, true, false},
		{Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1)}, Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1)}, true, false},
		{Coins{NewInt64Coin(testDenom1, 0)}, Coins{NewInt64Coin(testDenom2, 0)}, false, true},
		{Coins{NewInt64Coin(testDenom1, 0)}, Coins{NewInt64Coin(testDenom1, 1)}, false, false},
		{Coins{NewInt64Coin(testDenom1, 0)}, Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1)}, false, false},
		{Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1)}, Coins{NewInt64Coin(testDenom1, 0), NewInt64Coin(testDenom2, 1)}, true, false},
	}

	for tcnum, tc := range cases {
		tc := tc
		if tc.panics {
			require.Panics(t, func() { tc.inputOne.IsEqual(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsEqual(tc.inputTwo)
			require.Equal(t, tc.expected, res, "Equality is differed from exported. tc #%d, expected %b, actual %b.", tcnum, tc.expected, res)
		}
	}
}

func TestAddCoins(t *testing.T) {
	zero := NewInt(0)
	one := OneInt()
	two := NewInt(2)

	cases := []struct {
		inputOne Coins
		inputTwo Coins
		expected Coins
	}{
		{Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, two}, {testDenom2, two}}},
		{Coins{{testDenom1, zero}, {testDenom2, one}}, Coins{{testDenom1, zero}, {testDenom2, zero}}, Coins{{testDenom2, one}}},
		{Coins{{testDenom1, two}}, Coins{{testDenom2, zero}}, Coins{{testDenom1, two}}},
		{Coins{{testDenom1, one}}, Coins{{testDenom1, one}, {testDenom2, two}}, Coins{{testDenom1, two}, {testDenom2, two}}},
		{Coins{{testDenom1, zero}, {testDenom2, zero}}, Coins{{testDenom1, zero}, {testDenom2, zero}}, Coins(nil)},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.Add(tc.inputTwo...)
		assert.True(t, res.IsValid())
		require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
	}
}

func TestSubCoins(t *testing.T) {
	zero := NewInt(0)
	one := OneInt()
	two := NewInt(2)

	testCases := []struct {
		inputOne    Coins
		inputTwo    Coins
		expected    Coins
		shouldPanic bool
	}{
		{Coins{{testDenom1, two}}, Coins{{testDenom1, one}, {testDenom2, two}}, Coins{{testDenom1, one}, {testDenom2, two}}, true},
		{Coins{{testDenom1, two}}, Coins{{testDenom2, zero}}, Coins{{testDenom1, two}}, false},
		{Coins{{testDenom1, one}}, Coins{{testDenom2, zero}}, Coins{{testDenom1, one}}, false},
		{Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, one}}, Coins{{testDenom2, one}}, false},
		{Coins{{testDenom1, one}, {testDenom2, one}}, Coins{{testDenom1, two}}, Coins{}, true},
	}

	for i, tc := range testCases {
		tc := tc
		if tc.shouldPanic {
			require.Panics(t, func() { tc.inputOne.Sub(tc.inputTwo) })
		} else {
			res := tc.inputOne.Sub(tc.inputTwo)
			assert.True(t, res.IsValid())
			require.Equal(t, tc.expected, res, "sum of coins is incorrect, tc #%d", i)
		}
	}
}

func TestCoins_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		coins   Coins
		expPass bool
	}{
		{
			"valid lowercase coins",
			Coins{
				{"gas", OneInt()},
				{"mineral", OneInt()},
				{"tree", OneInt()},
			},
			true,
		},
		{
			"valid uppercase coins",
			Coins{
				{"GAS", OneInt()},
				{"MINERAL", OneInt()},
				{"TREE", OneInt()},
			},
			true,
		},
		{
			"valid uppercase coin",
			Coins{
				{"ATOM", OneInt()},
			},
			true,
		},
		{
			"valid lower and uppercase coins (1)",
			Coins{
				{"GAS", OneInt()},
				{"gAs", OneInt()},
			},
			true,
		},
		{
			"valid lower and uppercase coins (2)",
			Coins{
				{"ATOM", OneInt()},
				{"Atom", OneInt()},
				{"atom", OneInt()},
			},
			true,
		},
		{
			"mixed case (1)",
			Coins{
				{"MineraL", OneInt()},
				{"TREE", OneInt()},
				{"gAs", OneInt()},
			},
			true,
		},
		{
			"mixed case (2)",
			Coins{
				{"gAs", OneInt()},
				{"mineral", OneInt()},
			},
			true,
		},
		{
			"mixed case (3)",
			Coins{
				{"gAs", OneInt()},
			},
			true,
		},
		{
			"unicode letters and numbers",
			Coins{
				{"𐀀𐀆𐀉Ⅲ", OneInt()},
			},
			false,
		},
		{
			"emojis",
			Coins{
				{"🤑😋🤔", OneInt()},
			},
			false,
		},
		{
			"IBC denominations (ADR 001)",
			Coins{
				{"ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2", OneInt()},
				{"ibc/876563AAAACF739EB061C67CDB5EDF2B7C9FD4AA9D876450CC21210807C2820A", NewInt(2)},
			},
			true,
		},
		{
			"empty (1)",
			NewCoins(),
			true,
		},
		{
			"empty (2)",
			Coins{},
			true,
		},
		{
			"invalid denomination (1)",
			Coins{
				{"MineraL", OneInt()},
				{"0TREE", OneInt()},
				{"gAs", OneInt()},
			},
			false,
		},
		{
			"invalid denomination (2)",
			Coins{
				{"-GAS", OneInt()},
				{"gAs", OneInt()},
			},
			false,
		},
		{
			"bad sort (1)",
			Coins{
				{"tree", OneInt()},
				{"gas", OneInt()},
				{"mineral", OneInt()},
			},
			false,
		},
		{
			"bad sort (2)",
			Coins{
				{"gas", OneInt()},
				{"tree", OneInt()},
				{"mineral", OneInt()},
			},
			false,
		},
		{
			"non-positive amount (1)",
			Coins{
				{"gas", OneInt()},
				{"tree", NewInt(0)},
				{"mineral", OneInt()},
			},
			false,
		},
		{
			"non-positive amount (2)",
			Coins{
				{"gas", NewInt(-1)},
				{"tree", OneInt()},
				{"mineral", OneInt()},
			},
			false,
		}, {
			"duplicate denomination",
			Coins{
				{"gas", OneInt()},
				{"gas", OneInt()},
				{"mineral", OneInt()},
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.coins.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}

func TestCoinsGT(t *testing.T) {
	one := OneInt()
	two := NewInt(2)

	assert.False(t, Coins{}.IsAllGT(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGT(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGT(Coins{{testDenom2, two}}))
}

func TestCoinsLT(t *testing.T) {
	one := OneInt()
	two := NewInt(2)

	assert.False(t, Coins{}.IsAllLT(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLT(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom2, two}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(Coins{{testDenom1, two}, {testDenom2, two}}))
	assert.True(t, Coins{}.IsAllLT(Coins{{testDenom1, one}}))
}

func TestCoinsLTE(t *testing.T) {
	one := OneInt()
	two := NewInt(2)

	assert.True(t, Coins{}.IsAllLTE(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllLTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllLTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(Coins{{testDenom2, two}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{}.IsAllLTE(Coins{{testDenom1, one}}))
}

func TestParse(t *testing.T) {
	one := OneInt()

	cases := []struct {
		input    string
		valid    bool  // if false, we expect an error on parse
		expected Coins // if valid is true, make sure this is returned
	}{
		{"", true, nil},
		{"1foo", true, Coins{{"foo", one}}},
		{"10bar", true, Coins{{"bar", NewInt(10)}}},
		{"99bar,1foo", true, Coins{{"bar", NewInt(99)}, {"foo", one}}},
		{"98 bar , 1 foo  ", true, Coins{{"bar", NewInt(98)}, {"foo", one}}},
		{"  55\t \t bling\n", true, Coins{{"bling", NewInt(55)}}},
		{"2foo, 97 bar", true, Coins{{"bar", NewInt(97)}, {"foo", NewInt(2)}}},
		{"5 mycoin,", false, nil},             // no empty coins in a list
		{"2 3foo, 97 bar", false, nil},        // 3foo is invalid coin name
		{"11me coin, 12you coin", false, nil}, // no spaces in coin names
		{"1.2btc", false, nil},                // amount must be integer
		{"5foo:bar", false, nil},              // invalid separator
		{"10atom10", true, Coins{{"atom10", NewInt(10)}}},
	}

	for tcIndex, tc := range cases {
		res, err := ParseCoins(tc.input)
		if !tc.valid {
			require.Error(t, err, "%s: %#v. tc #%d", tc.input, res, tcIndex)
		} else if assert.Nil(t, err, "%s: %+v", tc.input, err) {
			require.Equal(t, tc.expected, res, "coin parsing was incorrect, tc #%d", tcIndex)
		}
	}
}

func TestSortCoins(t *testing.T) {
	good := Coins{
		NewInt64Coin("gas", 1),
		NewInt64Coin("mineral", 1),
		NewInt64Coin("tree", 1),
	}
	empty := Coins{
		NewInt64Coin("gold", 0),
	}
	badSort1 := Coins{
		NewInt64Coin("tree", 1),
		NewInt64Coin("gas", 1),
		NewInt64Coin("mineral", 1),
	}
	badSort2 := Coins{ // both are after the first one, but the second and third are in the wrong order
		NewInt64Coin("gas", 1),
		NewInt64Coin("tree", 1),
		NewInt64Coin("mineral", 1),
	}
	badAmt := Coins{
		NewInt64Coin("gas", 1),
		NewInt64Coin("tree", 0),
		NewInt64Coin("mineral", 1),
	}
	dup := Coins{
		NewInt64Coin("gas", 1),
		NewInt64Coin("gas", 1),
		NewInt64Coin("mineral", 1),
	}

	cases := []struct {
		name  string
		coins Coins
		validBefore,
		validAfter bool
	}{
		{"valid coins", good, true, true},
		{"empty coins", empty, false, false},
		{"bad sort (1)", badSort1, false, true},
		{"bad sort (2)", badSort2, false, true},
		{"zero value coin", badAmt, false, false},
		{"duplicate coins", dup, false, false},
	}

	for _, tc := range cases {
		err := tc.coins.Validate()
		if tc.validBefore {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}

		tc.coins.Sort()

		err = tc.coins.Validate()
		if tc.validAfter {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}

func TestAmountOf(t *testing.T) {
	case0 := Coins{}
	case1 := Coins{
		NewInt64Coin("gold", 0),
	}
	case2 := Coins{
		NewInt64Coin("gas", 1),
		NewInt64Coin("mineral", 1),
		NewInt64Coin("tree", 1),
	}
	case3 := Coins{
		NewInt64Coin("mineral", 1),
		NewInt64Coin("tree", 1),
	}
	case4 := Coins{
		NewInt64Coin("gas", 8),
	}

	cases := []struct {
		coins           Coins
		amountOf        int64
		amountOfSpace   int64
		amountOfGAS     int64
		amountOfMINERAL int64
		amountOfTREE    int64
	}{
		{case0, 0, 0, 0, 0, 0},
		{case1, 0, 0, 0, 0, 0},
		{case2, 0, 0, 1, 1, 1},
		{case3, 0, 0, 0, 1, 1},
		{case4, 0, 0, 8, 0, 0},
	}

	for _, tc := range cases {
		assert.Equal(t, NewInt(tc.amountOfGAS), tc.coins.AmountOf("gas"))
		assert.Equal(t, NewInt(tc.amountOfMINERAL), tc.coins.AmountOf("mineral"))
		assert.Equal(t, NewInt(tc.amountOfTREE), tc.coins.AmountOf("tree"))
	}

	assert.Panics(t, func() { cases[0].coins.AmountOf("10Invalid") })
}

func TestCoinsIsAnyGTE(t *testing.T) {
	one := OneInt()
	two := NewInt(2)

	assert.False(t, Coins{}.IsAnyGTE(Coins{}))
	assert.False(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{}))
	assert.False(t, Coins{}.IsAnyGTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{{testDenom1, two}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{{testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAnyGTE(Coins{{testDenom1, two}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{{testDenom1, one}}))
	assert.True(t, Coins{{testDenom1, two}}.IsAnyGTE(Coins{{testDenom1, one}}))
	assert.True(t, Coins{{testDenom1, one}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{{testDenom2, two}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, one}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAnyGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{{"xxx", one}, {"yyy", one}}.IsAnyGTE(Coins{{testDenom2, one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}

func TestCoinsIsAllGT(t *testing.T) {
	one := OneInt()
	two := NewInt(2)

	assert.False(t, Coins{}.IsAllGT(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGT(Coins{}))
	assert.False(t, Coins{}.IsAllGT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, two}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGT(Coins{{testDenom1, two}, {testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, one}}))
	assert.True(t, Coins{{testDenom1, two}}.IsAllGT(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, two}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, one}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGT(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{"xxx", one}, {"yyy", one}}.IsAllGT(Coins{{testDenom2, one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}

func TestCoinsIsAllGTE(t *testing.T) {
	one := OneInt()
	two := NewInt(2)

	assert.True(t, Coins{}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom2, two}}))
	assert.False(t, Coins{}.IsAllGTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom1, two}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGTE(Coins{{testDenom1, two}, {testDenom2, one}}))
	assert.True(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom1, one}}))
	assert.True(t, Coins{{testDenom1, two}}.IsAllGTE(Coins{{testDenom1, one}}))
	assert.False(t, Coins{{testDenom1, one}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, two}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{testDenom2, one}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.True(t, Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, one}}))
	assert.False(t, Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(Coins{{testDenom1, one}, {testDenom2, two}}))
	assert.False(t, Coins{{"xxx", one}, {"yyy", one}}.IsAllGTE(Coins{{testDenom2, one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}

func TestNewCoins(t *testing.T) {
	tenatom := NewInt64Coin("atom", 10)
	tenbtc := NewInt64Coin("btc", 10)
	zeroeth := NewInt64Coin("eth", 0)
	invalidCoin := Coin{"0ETH", OneInt()}
	tests := []struct {
		name      string
		coins     Coins
		want      Coins
		wantPanic bool
	}{
		{"empty args", []Coin{}, Coins{}, false},
		{"one coin", []Coin{tenatom}, Coins{tenatom}, false},
		{"sort after create", []Coin{tenbtc, tenatom}, Coins{tenatom, tenbtc}, false},
		{"sort and remove zeroes", []Coin{zeroeth, tenbtc, tenatom}, Coins{tenatom, tenbtc}, false},
		{"panic on dups", []Coin{tenatom, tenatom}, Coins{}, true},
		{"panic on invalid coin", []Coin{invalidCoin, tenatom}, Coins{}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() { NewCoins(tt.coins...) })
				return
			}
			got := NewCoins(tt.coins...)
			require.True(t, got.IsEqual(tt.want))
		})
	}
}

func TestCoinsIsAnyGT(t *testing.T) {
	twoAtom := NewInt64Coin("atom", 2)
	fiveAtom := NewInt64Coin("atom", 5)
	threeEth := NewInt64Coin("eth", 3)
	sixEth := NewInt64Coin("eth", 6)
	twoBtc := NewInt64Coin("btc", 2)

	tests := []struct {
		name    string
		coinsA  Coins
		coinsB  Coins
		expPass bool
	}{
		{"{} ≤ {}", Coins{}, Coins{}, false},
		{"{} ≤ 5atom", Coins{}, Coins{fiveAtom}, false},
		{"5atom > 2atom", Coins{fiveAtom}, Coins{twoAtom}, true},
		{"2atom ≤ 5atom", Coins{twoAtom}, Coins{fiveAtom}, false},
		{"2atom,6eth > 2btc,5atom,3eth", Coins{twoAtom, sixEth}, Coins{twoBtc, fiveAtom, threeEth}, true},
		{"2btc,2atom,3eth ≤ 5atom,6eth", Coins{twoBtc, twoAtom, threeEth}, Coins{fiveAtom, sixEth}, false},
		{"2atom,6eth ≤ 2btc,5atom", Coins{twoAtom, sixEth}, Coins{twoBtc, fiveAtom}, false},
	}

	for _, tc := range tests {
		require.True(t, tc.expPass == tc.coinsA.IsAnyGT(tc.coinsB), tc.name)
	}
}

func TestMarshalJSONCoins(t *testing.T) {
	cdc := codec.New()
	RegisterCodec(cdc)

	testCases := []struct {
		name      string
		input     Coins
		strOutput string
	}{
		{"nil coins", nil, `[]`},
		{"empty coins", Coins{}, `[]`},
		{"non-empty coins", NewCoins(NewInt64Coin("foo", 50)), `[{"denom":"foo","amount":"50"}]`},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			bz, err := cdc.MarshalJSON(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.strOutput, string(bz))

			var newCoins Coins
			require.NoError(t, cdc.UnmarshalJSON(bz, &newCoins))

			if tc.input.Empty() {
				require.Nil(t, newCoins)
			} else {
				require.Equal(t, tc.input, newCoins)
			}
		})
	}
}

func TestCoinAminoEncoding(t *testing.T) {
	c := NewInt64Coin(testDenom1, 5)

	bz1, err := cdc.MarshalBinaryBare(c)
	require.NoError(t, err)

	bz2, err := cdc.MarshalBinaryLengthPrefixed(c)
	require.NoError(t, err)

	bz3, err := c.Marshal()
	require.NoError(t, err)
	require.Equal(t, bz1, bz3)
	require.Equal(t, bz2[1:], bz3)
}
