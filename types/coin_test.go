package types_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	testDenom1 = "atom"
	testDenom2 = "muon"
)

type coinTestSuite struct {
	suite.Suite
}

func TestCoinTestSuite(t *testing.T) {
	suite.Run(t, new(coinTestSuite))
}

func (s *coinTestSuite) SetupSuite() {
	s.T().Parallel()
}

// ----------------------------------------------------------------------------
// Coin tests

func (s *coinTestSuite) TestCoin() {
	s.Require().Panics(func() { sdk.NewInt64Coin(testDenom1, -1) })
	s.Require().Panics(func() { sdk.NewCoin(testDenom1, sdk.NewInt(-1)) })
	s.Require().Equal(sdk.NewInt(10), sdk.NewInt64Coin(strings.ToUpper(testDenom1), 10).Amount)
	s.Require().Equal(sdk.NewInt(10), sdk.NewCoin(strings.ToUpper(testDenom1), sdk.NewInt(10)).Amount)
	s.Require().Equal(sdk.NewInt(5), sdk.NewInt64Coin(testDenom1, 5).Amount)
	s.Require().Equal(sdk.NewInt(5), sdk.NewCoin(testDenom1, sdk.NewInt(5)).Amount)
}

func (s *coinTestSuite) TestCoin_String() {
	coin := sdk.NewCoin(testDenom1, sdk.NewInt(10))
	s.Require().Equal(fmt.Sprintf("10%s", testDenom1), coin.String())
}

func (s *coinTestSuite) TestIsEqualCoin() {
	cases := []struct {
		inputOne sdk.Coin
		inputTwo sdk.Coin
		expected bool
		panics   bool
	}{
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom1, 1), true, false},
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom2, 1), false, true},
		{sdk.NewInt64Coin("stake", 1), sdk.NewInt64Coin("stake", 10), false, false},
	}

	for tcIndex, tc := range cases {
		tc := tc
		if tc.panics {
			s.Require().Panics(func() { tc.inputOne.IsEqual(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsEqual(tc.inputTwo)
			s.Require().Equal(tc.expected, res, "coin equality relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func (s *coinTestSuite) TestCoinIsValid() {
	loremIpsum := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam viverra dui vel nulla aliquet, non dictum elit aliquam. Proin consequat leo in consectetur mattis. Phasellus eget odio luctus, rutrum dolor at, venenatis ante. Praesent metus erat, sodales vitae sagittis eget, commodo non ipsum. Duis eget urna quis erat mattis pulvinar. Vivamus egestas imperdiet sem, porttitor hendrerit lorem pulvinar in. Vivamus laoreet sapien eget libero euismod tristique. Suspendisse tincidunt nulla quis luctus mattis.
	Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Sed id turpis at erat placerat fermentum id sed sapien. Fusce mattis enim id nulla viverra, eget placerat eros aliquet. Nunc fringilla urna ac condimentum ultricies. Praesent in eros ac neque fringilla sodales. Donec ut venenatis eros. Quisque iaculis lectus neque, a varius sem ullamcorper nec. Cras tincidunt dignissim libero nec volutpat. Donec molestie enim sed metus venenatis, quis elementum sem varius. Curabitur eu venenatis nulla.
	Cras sit amet ligula vel turpis placerat sollicitudin. Nunc massa odio, eleifend id lacus nec, ultricies elementum arcu. Donec imperdiet nulla lacus, a venenatis lacus fermentum nec. Proin vestibulum dolor enim, vitae posuere velit aliquet non. Suspendisse pharetra condimentum nunc tincidunt viverra. Etiam posuere, ligula ut maximus congue, mauris orci consectetur velit, vel finibus eros metus non tellus. Nullam et dictum metus. Aliquam maximus fermentum mauris elementum aliquet. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Etiam dapibus lectus sed tellus rutrum tincidunt. Nulla at dolor sem. Ut non dictum arcu, eget congue sem.`

	loremIpsum = strings.ReplaceAll(loremIpsum, " ", "")
	loremIpsum = strings.ReplaceAll(loremIpsum, ".", "")
	loremIpsum = strings.ReplaceAll(loremIpsum, ",", "")

	cases := []struct {
		coin       sdk.Coin
		expectPass bool
	}{
		{sdk.Coin{testDenom1, sdk.NewInt(-1)}, false},
		{sdk.Coin{testDenom1, sdk.NewInt(0)}, true},
		{sdk.Coin{testDenom1, sdk.OneInt()}, true},
		{sdk.Coin{"Atom", sdk.OneInt()}, true},
		{sdk.Coin{"ATOM", sdk.OneInt()}, true},
		{sdk.Coin{"a", sdk.OneInt()}, false},
		{sdk.Coin{loremIpsum, sdk.OneInt()}, false},
		{sdk.Coin{"ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2", sdk.OneInt()}, true},
		{sdk.Coin{"atOm", sdk.OneInt()}, true},
		{sdk.Coin{"     ", sdk.OneInt()}, false},
	}

	for i, tc := range cases {
		s.Require().Equal(tc.expectPass, tc.coin.IsValid(), "unexpected result for IsValid, tc #%d", i)
	}
}

func (s *coinTestSuite) TestCustomValidation() {

	newDnmRegex := `[\x{1F600}-\x{1F6FF}]`
	sdk.SetCoinDenomRegex(func() string {
		return newDnmRegex
	})

	cases := []struct {
		coin       sdk.Coin
		expectPass bool
	}{
		{sdk.Coin{"🙂", sdk.NewInt(1)}, true},
		{sdk.Coin{"🙁", sdk.NewInt(1)}, true},
		{sdk.Coin{"🌶", sdk.NewInt(1)}, false}, // outside the unicode range listed above
		{sdk.Coin{"asdf", sdk.NewInt(1)}, false},
		{sdk.Coin{"", sdk.NewInt(1)}, false},
	}

	for i, tc := range cases {
		s.Require().Equal(tc.expectPass, tc.coin.IsValid(), "unexpected result for IsValid, tc #%d", i)
	}
	sdk.SetCoinDenomRegex(sdk.DefaultCoinDenomRegex)
}

func (s *coinTestSuite) TestAddCoin() {
	cases := []struct {
		inputOne    sdk.Coin
		inputTwo    sdk.Coin
		expected    sdk.Coin
		shouldPanic bool
	}{
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom1, 2), false},
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom1, 0), sdk.NewInt64Coin(testDenom1, 1), false},
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom2, 1), sdk.NewInt64Coin(testDenom1, 1), true},
	}

	for tcIndex, tc := range cases {
		tc := tc
		if tc.shouldPanic {
			s.Require().Panics(func() { tc.inputOne.Add(tc.inputTwo) })
		} else {
			res := tc.inputOne.Add(tc.inputTwo)
			s.Require().Equal(tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
		}
	}
}

func (s *coinTestSuite) TestSubCoin() {
	cases := []struct {
		inputOne    sdk.Coin
		inputTwo    sdk.Coin
		expected    sdk.Coin
		shouldPanic bool
	}{
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom2, 1), sdk.NewInt64Coin(testDenom1, 1), true},
		{sdk.NewInt64Coin(testDenom1, 10), sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom1, 9), false},
		{sdk.NewInt64Coin(testDenom1, 5), sdk.NewInt64Coin(testDenom1, 3), sdk.NewInt64Coin(testDenom1, 2), false},
		{sdk.NewInt64Coin(testDenom1, 5), sdk.NewInt64Coin(testDenom1, 0), sdk.NewInt64Coin(testDenom1, 5), false},
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom1, 5), sdk.Coin{}, true},
	}

	for tcIndex, tc := range cases {
		tc := tc
		if tc.shouldPanic {
			s.Require().Panics(func() { tc.inputOne.Sub(tc.inputTwo) })
		} else {
			res := tc.inputOne.Sub(tc.inputTwo)
			s.Require().Equal(tc.expected, res, "difference of coins is incorrect, tc #%d", tcIndex)
		}
	}

	tc := struct {
		inputOne sdk.Coin
		inputTwo sdk.Coin
		expected int64
	}{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom1, 1), 0}
	res := tc.inputOne.Sub(tc.inputTwo)
	s.Require().Equal(tc.expected, res.Amount.Int64())
}

func (s *coinTestSuite) TestIsGTECoin() {
	cases := []struct {
		inputOne sdk.Coin
		inputTwo sdk.Coin
		expected bool
		panics   bool
	}{
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom1, 1), true, false},
		{sdk.NewInt64Coin(testDenom1, 2), sdk.NewInt64Coin(testDenom1, 1), true, false},
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom2, 1), false, true},
	}

	for tcIndex, tc := range cases {
		tc := tc
		if tc.panics {
			s.Require().Panics(func() { tc.inputOne.IsGTE(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsGTE(tc.inputTwo)
			s.Require().Equal(tc.expected, res, "coin GTE relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func (s *coinTestSuite) TestIsLTCoin() {
	cases := []struct {
		inputOne sdk.Coin
		inputTwo sdk.Coin
		expected bool
		panics   bool
	}{
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom1, 1), false, false},
		{sdk.NewInt64Coin(testDenom1, 2), sdk.NewInt64Coin(testDenom1, 1), false, false},
		{sdk.NewInt64Coin(testDenom1, 0), sdk.NewInt64Coin(testDenom2, 1), false, true},
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom2, 1), false, true},
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom1, 1), false, false},
		{sdk.NewInt64Coin(testDenom1, 1), sdk.NewInt64Coin(testDenom1, 2), true, false},
	}

	for tcIndex, tc := range cases {
		tc := tc
		if tc.panics {
			s.Require().Panics(func() { tc.inputOne.IsLT(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsLT(tc.inputTwo)
			s.Require().Equal(tc.expected, res, "coin LT relation is incorrect, tc #%d", tcIndex)
		}
	}
}

func (s *coinTestSuite) TestCoinIsZero() {
	coin := sdk.NewInt64Coin(testDenom1, 0)
	res := coin.IsZero()
	s.Require().True(res)

	coin = sdk.NewInt64Coin(testDenom1, 1)
	res = coin.IsZero()
	s.Require().False(res)
}

func (s *coinTestSuite) TestCoinIsNil() {
	coin := sdk.Coin{}
	res := coin.IsNil()
	s.Require().True(res)

	coin = sdk.Coin{Denom: "uatom"}
	res = coin.IsNil()
	s.Require().True(res)

	coin = sdk.NewInt64Coin(testDenom1, 1)
	res = coin.IsNil()
	s.Require().False(res)
}

func (s *coinTestSuite) TestFilteredZeroCoins() {
	cases := []struct {
		name     string
		input    sdk.Coins
		original string
		expected string
	}{
		{
			name: "all greater than zero",
			input: sdk.Coins{
				{"testa", sdk.OneInt()},
				{"testb", sdk.NewInt(2)},
				{"testc", sdk.NewInt(3)},
				{"testd", sdk.NewInt(4)},
				{"teste", sdk.NewInt(5)},
			},
			original: "1testa,2testb,3testc,4testd,5teste",
			expected: "1testa,2testb,3testc,4testd,5teste",
		},
		{
			name: "zero coin in middle",
			input: sdk.Coins{
				{"testa", sdk.OneInt()},
				{"testb", sdk.NewInt(2)},
				{"testc", sdk.NewInt(0)},
				{"testd", sdk.NewInt(4)},
				{"teste", sdk.NewInt(5)},
			},
			original: "1testa,2testb,0testc,4testd,5teste",
			expected: "1testa,2testb,4testd,5teste",
		},
		{
			name: "zero coin end (unordered)",
			input: sdk.Coins{
				{"teste", sdk.NewInt(5)},
				{"testc", sdk.NewInt(3)},
				{"testa", sdk.OneInt()},
				{"testd", sdk.NewInt(4)},
				{"testb", sdk.NewInt(0)},
			},
			original: "5teste,3testc,1testa,4testd,0testb",
			expected: "1testa,3testc,4testd,5teste",
		},
	}

	for _, tt := range cases {
		undertest := sdk.NewCoins(tt.input...)
		s.Require().Equal(tt.expected, undertest.String(), "NewCoins must return expected results")
		s.Require().Equal(tt.original, tt.input.String(), "input must be unmodified and match original")
	}
}

// ----------------------------------------------------------------------------
// Coins tests

func (s *coinTestSuite) TestCoins_String() {
	cases := []struct {
		name     string
		input    sdk.Coins
		expected string
	}{
		{
			"empty coins",
			sdk.Coins{},
			"",
		},
		{
			"single coin",
			sdk.Coins{{"tree", sdk.OneInt()}},
			"1tree",
		},
		{
			"multiple coins",
			sdk.Coins{
				{"tree", sdk.OneInt()},
				{"gas", sdk.OneInt()},
				{"mineral", sdk.OneInt()},
			},
			"1tree,1gas,1mineral",
		},
	}

	for _, tt := range cases {
		s.Require().Equal(tt.expected, tt.input.String())
	}
}

func (s *coinTestSuite) TestIsZeroCoins() {
	cases := []struct {
		inputOne sdk.Coins
		expected bool
	}{
		{sdk.Coins{}, true},
		{sdk.Coins{sdk.NewInt64Coin(testDenom1, 0)}, true},
		{sdk.Coins{sdk.NewInt64Coin(testDenom1, 0), sdk.NewInt64Coin(testDenom2, 0)}, true},
		{sdk.Coins{sdk.NewInt64Coin(testDenom1, 1)}, false},
		{sdk.Coins{sdk.NewInt64Coin(testDenom1, 0), sdk.NewInt64Coin(testDenom2, 1)}, false},
	}

	for _, tc := range cases {
		res := tc.inputOne.IsZero()
		s.Require().Equal(tc.expected, res)
	}
}

func (s *coinTestSuite) TestEqualCoins() {
	cases := []struct {
		inputOne sdk.Coins
		inputTwo sdk.Coins
		expected bool
		panics   bool
	}{
		{sdk.Coins{}, sdk.Coins{}, true, false},
		{sdk.Coins{sdk.NewInt64Coin(testDenom1, 0)}, sdk.Coins{sdk.NewInt64Coin(testDenom1, 0)}, true, false},
		{sdk.Coins{sdk.NewInt64Coin(testDenom1, 0), sdk.NewInt64Coin(testDenom2, 1)}, sdk.Coins{sdk.NewInt64Coin(testDenom1, 0), sdk.NewInt64Coin(testDenom2, 1)}, true, false},
		{sdk.Coins{sdk.NewInt64Coin(testDenom1, 0)}, sdk.Coins{sdk.NewInt64Coin(testDenom2, 0)}, false, true},
		{sdk.Coins{sdk.NewInt64Coin(testDenom1, 0)}, sdk.Coins{sdk.NewInt64Coin(testDenom1, 1)}, false, false},
		{sdk.Coins{sdk.NewInt64Coin(testDenom1, 0)}, sdk.Coins{sdk.NewInt64Coin(testDenom1, 0), sdk.NewInt64Coin(testDenom2, 1)}, false, false},
		{sdk.Coins{sdk.NewInt64Coin(testDenom1, 0), sdk.NewInt64Coin(testDenom2, 1)}, sdk.Coins{sdk.NewInt64Coin(testDenom1, 0), sdk.NewInt64Coin(testDenom2, 1)}, true, false},
	}

	for tcnum, tc := range cases {
		tc := tc
		if tc.panics {
			s.Require().Panics(func() { tc.inputOne.IsEqual(tc.inputTwo) })
		} else {
			res := tc.inputOne.IsEqual(tc.inputTwo)
			s.Require().Equal(tc.expected, res, "Equality is differed from exported. tc #%d, expected %b, actual %b.", tcnum, tc.expected, res)
		}
	}
}

func (s *coinTestSuite) TestAddCoins() {
	zero := sdk.NewInt(0)
	one := sdk.OneInt()
	two := sdk.NewInt(2)

	cases := []struct {
		inputOne sdk.Coins
		inputTwo sdk.Coins
		expected sdk.Coins
	}{
		{sdk.Coins{{testDenom1, one}, {testDenom2, one}}, sdk.Coins{{testDenom1, one}, {testDenom2, one}}, sdk.Coins{{testDenom1, two}, {testDenom2, two}}},
		{sdk.Coins{{testDenom1, zero}, {testDenom2, one}}, sdk.Coins{{testDenom1, zero}, {testDenom2, zero}}, sdk.Coins{{testDenom2, one}}},
		{sdk.Coins{{testDenom1, two}}, sdk.Coins{{testDenom2, zero}}, sdk.Coins{{testDenom1, two}}},
		{sdk.Coins{{testDenom1, one}}, sdk.Coins{{testDenom1, one}, {testDenom2, two}}, sdk.Coins{{testDenom1, two}, {testDenom2, two}}},
		{sdk.Coins{{testDenom1, zero}, {testDenom2, zero}}, sdk.Coins{{testDenom1, zero}, {testDenom2, zero}}, sdk.Coins(nil)},
	}

	for tcIndex, tc := range cases {
		res := tc.inputOne.Add(tc.inputTwo...)
		s.Require().True(res.IsValid())
		s.Require().Equal(tc.expected, res, "sum of coins is incorrect, tc #%d", tcIndex)
	}
}

func (s *coinTestSuite) TestSubCoins() {
	zero := sdk.NewInt(0)
	one := sdk.OneInt()
	two := sdk.NewInt(2)

	testCases := []struct {
		inputOne    sdk.Coins
		inputTwo    sdk.Coins
		expected    sdk.Coins
		shouldPanic bool
	}{
		{sdk.Coins{{testDenom1, two}}, sdk.Coins{{testDenom1, one}, {testDenom2, two}}, sdk.Coins{{testDenom1, one}, {testDenom2, two}}, true},
		{sdk.Coins{{testDenom1, two}}, sdk.Coins{{testDenom2, zero}}, sdk.Coins{{testDenom1, two}}, false},
		{sdk.Coins{{testDenom1, one}}, sdk.Coins{{testDenom2, zero}}, sdk.Coins{{testDenom1, one}}, false},
		{sdk.Coins{{testDenom1, one}, {testDenom2, one}}, sdk.Coins{{testDenom1, one}}, sdk.Coins{{testDenom2, one}}, false},
		{sdk.Coins{{testDenom1, one}, {testDenom2, one}}, sdk.Coins{{testDenom1, two}}, sdk.Coins{}, true},
	}

	for i, tc := range testCases {
		tc := tc
		if tc.shouldPanic {
			s.Require().Panics(func() { tc.inputOne.Sub(tc.inputTwo) })
		} else {
			res := tc.inputOne.Sub(tc.inputTwo)
			s.Require().True(res.IsValid())
			s.Require().Equal(tc.expected, res, "sum of coins is incorrect, tc #%d", i)
		}
	}
}

func (s *coinTestSuite) TestCoins_Validate() {
	testCases := []struct {
		name    string
		coins   sdk.Coins
		expPass bool
	}{
		{
			"valid lowercase coins",
			sdk.Coins{
				{"gas", sdk.OneInt()},
				{"mineral", sdk.OneInt()},
				{"tree", sdk.OneInt()},
			},
			true,
		},
		{
			"valid uppercase coins",
			sdk.Coins{
				{"GAS", sdk.OneInt()},
				{"MINERAL", sdk.OneInt()},
				{"TREE", sdk.OneInt()},
			},
			true,
		},
		{
			"valid uppercase coin",
			sdk.Coins{
				{"ATOM", sdk.OneInt()},
			},
			true,
		},
		{
			"valid lower and uppercase coins (1)",
			sdk.Coins{
				{"GAS", sdk.OneInt()},
				{"gAs", sdk.OneInt()},
			},
			true,
		},
		{
			"valid lower and uppercase coins (2)",
			sdk.Coins{
				{"ATOM", sdk.OneInt()},
				{"Atom", sdk.OneInt()},
				{"atom", sdk.OneInt()},
			},
			true,
		},
		{
			"mixed case (1)",
			sdk.Coins{
				{"MineraL", sdk.OneInt()},
				{"TREE", sdk.OneInt()},
				{"gAs", sdk.OneInt()},
			},
			true,
		},
		{
			"mixed case (2)",
			sdk.Coins{
				{"gAs", sdk.OneInt()},
				{"mineral", sdk.OneInt()},
			},
			true,
		},
		{
			"mixed case (3)",
			sdk.Coins{
				{"gAs", sdk.OneInt()},
			},
			true,
		},
		{
			"unicode letters and numbers",
			sdk.Coins{
				{"𐀀𐀆𐀉Ⅲ", sdk.OneInt()},
			},
			false,
		},
		{
			"emojis",
			sdk.Coins{
				{"🤑😋🤔", sdk.OneInt()},
			},
			false,
		},
		{
			"IBC denominations (ADR 001)",
			sdk.Coins{
				{"ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2", sdk.OneInt()},
				{"ibc/876563AAAACF739EB061C67CDB5EDF2B7C9FD4AA9D876450CC21210807C2820A", sdk.NewInt(2)},
			},
			true,
		},
		{
			"empty (1)",
			sdk.NewCoins(),
			true,
		},
		{
			"empty (2)",
			sdk.Coins{},
			true,
		},
		{
			"invalid denomination (1)",
			sdk.Coins{
				{"MineraL", sdk.OneInt()},
				{"0TREE", sdk.OneInt()},
				{"gAs", sdk.OneInt()},
			},
			false,
		},
		{
			"invalid denomination (2)",
			sdk.Coins{
				{"-GAS", sdk.OneInt()},
				{"gAs", sdk.OneInt()},
			},
			false,
		},
		{
			"bad sort (1)",
			sdk.Coins{
				{"tree", sdk.OneInt()},
				{"gas", sdk.OneInt()},
				{"mineral", sdk.OneInt()},
			},
			false,
		},
		{
			"bad sort (2)",
			sdk.Coins{
				{"gas", sdk.OneInt()},
				{"tree", sdk.OneInt()},
				{"mineral", sdk.OneInt()},
			},
			false,
		},
		{
			"non-positive amount (1)",
			sdk.Coins{
				{"gas", sdk.OneInt()},
				{"tree", sdk.NewInt(0)},
				{"mineral", sdk.OneInt()},
			},
			false,
		},
		{
			"non-positive amount (2)",
			sdk.Coins{
				{"gas", sdk.NewInt(-1)},
				{"tree", sdk.OneInt()},
				{"mineral", sdk.OneInt()},
			},
			false,
		}, {
			"duplicate denomination",
			sdk.Coins{
				{"gas", sdk.OneInt()},
				{"gas", sdk.OneInt()},
				{"mineral", sdk.OneInt()},
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.coins.Validate()
		if tc.expPass {
			s.Require().NoError(err, tc.name)
		} else {
			s.Require().Error(err, tc.name)
		}
	}
}

func (s *coinTestSuite) TestCoinsGT() {
	one := sdk.OneInt()
	two := sdk.NewInt(2)

	s.Require().False(sdk.Coins{}.IsAllGT(sdk.Coins{}))
	s.Require().True(sdk.Coins{{testDenom1, one}}.IsAllGT(sdk.Coins{}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllGT(sdk.Coins{{testDenom1, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllGT(sdk.Coins{{testDenom2, one}}))
	s.Require().True(sdk.Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGT(sdk.Coins{{testDenom2, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGT(sdk.Coins{{testDenom2, two}}))
}

func (s *coinTestSuite) TestCoinsLT() {
	one := sdk.OneInt()
	two := sdk.NewInt(2)

	s.Require().False(sdk.Coins{}.IsAllLT(sdk.Coins{}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllLT(sdk.Coins{}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllLT(sdk.Coins{{testDenom1, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllLT(sdk.Coins{{testDenom2, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(sdk.Coins{{testDenom2, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(sdk.Coins{{testDenom2, two}}))
	s.Require().False(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(sdk.Coins{{testDenom1, one}, {testDenom2, one}}))
	s.Require().True(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLT(sdk.Coins{{testDenom1, two}, {testDenom2, two}}))
	s.Require().True(sdk.Coins{}.IsAllLT(sdk.Coins{{testDenom1, one}}))
}

func (s *coinTestSuite) TestCoinsLTE() {
	one := sdk.OneInt()
	two := sdk.NewInt(2)

	s.Require().True(sdk.Coins{}.IsAllLTE(sdk.Coins{}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllLTE(sdk.Coins{}))
	s.Require().True(sdk.Coins{{testDenom1, one}}.IsAllLTE(sdk.Coins{{testDenom1, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllLTE(sdk.Coins{{testDenom2, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(sdk.Coins{{testDenom2, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(sdk.Coins{{testDenom2, two}}))
	s.Require().True(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(sdk.Coins{{testDenom1, one}, {testDenom2, one}}))
	s.Require().True(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllLTE(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().True(sdk.Coins{}.IsAllLTE(sdk.Coins{{testDenom1, one}}))
}

func (s *coinTestSuite) TestParseCoins() {
	one := sdk.OneInt()

	cases := []struct {
		input    string
		valid    bool      // if false, we expect an error on parse
		expected sdk.Coins // if valid is true, make sure this is returned
	}{
		{"", true, nil},
		{"0stake", true, sdk.Coins{}}, // remove zero coins
		{"0stake,1foo,99bar", true, sdk.Coins{{"bar", sdk.NewInt(99)}, {"foo", one}}}, // remove zero coins
		{"1foo", true, sdk.Coins{{"foo", one}}},
		{"10btc,1atom,20btc", false, nil},
		{"10bar", true, sdk.Coins{{"bar", sdk.NewInt(10)}}},
		{"99bar,1foo", true, sdk.Coins{{"bar", sdk.NewInt(99)}, {"foo", one}}},
		{"98 bar , 1 foo  ", true, sdk.Coins{{"bar", sdk.NewInt(98)}, {"foo", one}}},
		{"  55\t \t bling\n", true, sdk.Coins{{"bling", sdk.NewInt(55)}}},
		{"2foo, 97 bar", true, sdk.Coins{{"bar", sdk.NewInt(97)}, {"foo", sdk.NewInt(2)}}},
		{"5 mycoin,", false, nil},                           // no empty coins in a list
		{"2 3foo, 97 bar", false, nil},                      // 3foo is invalid coin name
		{"11me coin, 12you coin", false, nil},               // no spaces in coin names
		{"1.2btc", true, sdk.Coins{{"btc", sdk.NewInt(1)}}}, // amount can be decimal, will get truncated
		{"5foo:bar", false, nil},                            // invalid separator
		{"10atom10", true, sdk.Coins{{"atom10", sdk.NewInt(10)}}},
		{"200transfer/channelToA/uatom", true, sdk.Coins{{"transfer/channelToA/uatom", sdk.NewInt(200)}}},
		{"50ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2", true, sdk.Coins{{"ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2", sdk.NewInt(50)}}},
	}

	for tcIndex, tc := range cases {
		res, err := sdk.ParseCoinsNormalized(tc.input)
		if !tc.valid {
			s.Require().Error(err, "%s: %#v. tc #%d", tc.input, res, tcIndex)
		} else if s.Assert().Nil(err, "%s: %+v", tc.input, err) {
			s.Require().Equal(tc.expected, res, "coin parsing was incorrect, tc #%d", tcIndex)
		}
	}
}

func (s *coinTestSuite) TestSortCoins() {
	good := sdk.Coins{
		sdk.NewInt64Coin("gas", 1),
		sdk.NewInt64Coin("mineral", 1),
		sdk.NewInt64Coin("tree", 1),
	}
	empty := sdk.Coins{
		sdk.NewInt64Coin("gold", 0),
	}
	badSort1 := sdk.Coins{
		sdk.NewInt64Coin("tree", 1),
		sdk.NewInt64Coin("gas", 1),
		sdk.NewInt64Coin("mineral", 1),
	}
	badSort2 := sdk.Coins{ // both are after the first one, but the second and third are in the wrong order
		sdk.NewInt64Coin("gas", 1),
		sdk.NewInt64Coin("tree", 1),
		sdk.NewInt64Coin("mineral", 1),
	}
	badAmt := sdk.Coins{
		sdk.NewInt64Coin("gas", 1),
		sdk.NewInt64Coin("tree", 0),
		sdk.NewInt64Coin("mineral", 1),
	}
	dup := sdk.Coins{
		sdk.NewInt64Coin("gas", 1),
		sdk.NewInt64Coin("gas", 1),
		sdk.NewInt64Coin("mineral", 1),
	}

	cases := []struct {
		name  string
		coins sdk.Coins
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
			s.Require().NoError(err, tc.name)
		} else {
			s.Require().Error(err, tc.name)
		}

		tc.coins.Sort()

		err = tc.coins.Validate()
		if tc.validAfter {
			s.Require().NoError(err, tc.name)
		} else {
			s.Require().Error(err, tc.name)
		}
	}
}

func (s *coinTestSuite) TestAmountOf() {
	case0 := sdk.Coins{}
	case1 := sdk.Coins{
		sdk.NewInt64Coin("gold", 0),
	}
	case2 := sdk.Coins{
		sdk.NewInt64Coin("gas", 1),
		sdk.NewInt64Coin("mineral", 1),
		sdk.NewInt64Coin("tree", 1),
	}
	case3 := sdk.Coins{
		sdk.NewInt64Coin("mineral", 1),
		sdk.NewInt64Coin("tree", 1),
	}
	case4 := sdk.Coins{
		sdk.NewInt64Coin("gas", 8),
	}

	cases := []struct {
		coins           sdk.Coins
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
		s.Require().Equal(sdk.NewInt(tc.amountOfGAS), tc.coins.AmountOf("gas"))
		s.Require().Equal(sdk.NewInt(tc.amountOfMINERAL), tc.coins.AmountOf("mineral"))
		s.Require().Equal(sdk.NewInt(tc.amountOfTREE), tc.coins.AmountOf("tree"))
	}

	s.Require().Panics(func() { cases[0].coins.AmountOf("10Invalid") })
}

func (s *coinTestSuite) TestCoinsIsAnyGTE() {
	one := sdk.OneInt()
	two := sdk.NewInt(2)

	s.Require().False(sdk.Coins{}.IsAnyGTE(sdk.Coins{}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAnyGTE(sdk.Coins{}))
	s.Require().False(sdk.Coins{}.IsAnyGTE(sdk.Coins{{testDenom1, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAnyGTE(sdk.Coins{{testDenom1, two}}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAnyGTE(sdk.Coins{{testDenom2, one}}))
	s.Require().True(sdk.Coins{{testDenom1, one}, {testDenom2, two}}.IsAnyGTE(sdk.Coins{{testDenom1, two}, {testDenom2, one}}))
	s.Require().True(sdk.Coins{{testDenom1, one}}.IsAnyGTE(sdk.Coins{{testDenom1, one}}))
	s.Require().True(sdk.Coins{{testDenom1, two}}.IsAnyGTE(sdk.Coins{{testDenom1, one}}))
	s.Require().True(sdk.Coins{{testDenom1, one}}.IsAnyGTE(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().True(sdk.Coins{{testDenom2, two}}.IsAnyGTE(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().False(sdk.Coins{{testDenom2, one}}.IsAnyGTE(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().True(sdk.Coins{{testDenom1, one}, {testDenom2, two}}.IsAnyGTE(sdk.Coins{{testDenom1, one}, {testDenom2, one}}))
	s.Require().True(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAnyGTE(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().True(sdk.Coins{{"xxx", one}, {"yyy", one}}.IsAnyGTE(sdk.Coins{{testDenom2, one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}

func (s *coinTestSuite) TestCoinsIsAllGT() {
	one := sdk.OneInt()
	two := sdk.NewInt(2)

	s.Require().False(sdk.Coins{}.IsAllGT(sdk.Coins{}))
	s.Require().True(sdk.Coins{{testDenom1, one}}.IsAllGT(sdk.Coins{}))
	s.Require().False(sdk.Coins{}.IsAllGT(sdk.Coins{{testDenom1, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllGT(sdk.Coins{{testDenom1, two}}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllGT(sdk.Coins{{testDenom2, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGT(sdk.Coins{{testDenom1, two}, {testDenom2, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllGT(sdk.Coins{{testDenom1, one}}))
	s.Require().True(sdk.Coins{{testDenom1, two}}.IsAllGT(sdk.Coins{{testDenom1, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllGT(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().False(sdk.Coins{{testDenom2, two}}.IsAllGT(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().False(sdk.Coins{{testDenom2, one}}.IsAllGT(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().False(sdk.Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGT(sdk.Coins{{testDenom1, one}, {testDenom2, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGT(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().False(sdk.Coins{{"xxx", one}, {"yyy", one}}.IsAllGT(sdk.Coins{{testDenom2, one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}

func (s *coinTestSuite) TestCoinsIsAllGTE() {
	one := sdk.OneInt()
	two := sdk.NewInt(2)

	s.Require().True(sdk.Coins{}.IsAllGTE(sdk.Coins{}))
	s.Require().True(sdk.Coins{{testDenom1, one}}.IsAllGTE(sdk.Coins{}))
	s.Require().True(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(sdk.Coins{{testDenom2, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(sdk.Coins{{testDenom2, two}}))
	s.Require().False(sdk.Coins{}.IsAllGTE(sdk.Coins{{testDenom1, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllGTE(sdk.Coins{{testDenom1, two}}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllGTE(sdk.Coins{{testDenom2, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGTE(sdk.Coins{{testDenom1, two}, {testDenom2, one}}))
	s.Require().True(sdk.Coins{{testDenom1, one}}.IsAllGTE(sdk.Coins{{testDenom1, one}}))
	s.Require().True(sdk.Coins{{testDenom1, two}}.IsAllGTE(sdk.Coins{{testDenom1, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}}.IsAllGTE(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().False(sdk.Coins{{testDenom2, two}}.IsAllGTE(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().False(sdk.Coins{{testDenom2, one}}.IsAllGTE(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().True(sdk.Coins{{testDenom1, one}, {testDenom2, two}}.IsAllGTE(sdk.Coins{{testDenom1, one}, {testDenom2, one}}))
	s.Require().False(sdk.Coins{{testDenom1, one}, {testDenom2, one}}.IsAllGTE(sdk.Coins{{testDenom1, one}, {testDenom2, two}}))
	s.Require().False(sdk.Coins{{"xxx", one}, {"yyy", one}}.IsAllGTE(sdk.Coins{{testDenom2, one}, {"ccc", one}, {"yyy", one}, {"zzz", one}}))
}

func (s *coinTestSuite) TestNewCoins() {
	tenatom := sdk.NewInt64Coin("atom", 10)
	tenbtc := sdk.NewInt64Coin("btc", 10)
	zeroeth := sdk.NewInt64Coin("eth", 0)
	invalidCoin := sdk.Coin{"0ETH", sdk.OneInt()}
	tests := []struct {
		name      string
		coins     sdk.Coins
		want      sdk.Coins
		wantPanic bool
	}{
		{"empty args", []sdk.Coin{}, sdk.Coins{}, false},
		{"one coin", []sdk.Coin{tenatom}, sdk.Coins{tenatom}, false},
		{"sort after create", []sdk.Coin{tenbtc, tenatom}, sdk.Coins{tenatom, tenbtc}, false},
		{"sort and remove zeroes", []sdk.Coin{zeroeth, tenbtc, tenatom}, sdk.Coins{tenatom, tenbtc}, false},
		{"panic on dups", []sdk.Coin{tenatom, tenatom}, sdk.Coins{}, true},
		{"panic on invalid coin", []sdk.Coin{invalidCoin, tenatom}, sdk.Coins{}, true},
	}
	for _, tt := range tests {
		if tt.wantPanic {
			s.Require().Panics(func() { sdk.NewCoins(tt.coins...) })
			continue
		}
		got := sdk.NewCoins(tt.coins...)
		s.Require().True(got.IsEqual(tt.want))
	}
}

func (s *coinTestSuite) TestCoinsIsAnyGT() {
	twoAtom := sdk.NewInt64Coin("atom", 2)
	fiveAtom := sdk.NewInt64Coin("atom", 5)
	threeEth := sdk.NewInt64Coin("eth", 3)
	sixEth := sdk.NewInt64Coin("eth", 6)
	twoBtc := sdk.NewInt64Coin("btc", 2)

	tests := []struct {
		name    string
		coinsA  sdk.Coins
		coinsB  sdk.Coins
		expPass bool
	}{
		{"{} ≤ {}", sdk.Coins{}, sdk.Coins{}, false},
		{"{} ≤ 5atom", sdk.Coins{}, sdk.Coins{fiveAtom}, false},
		{"5atom > 2atom", sdk.Coins{fiveAtom}, sdk.Coins{twoAtom}, true},
		{"2atom ≤ 5atom", sdk.Coins{twoAtom}, sdk.Coins{fiveAtom}, false},
		{"2atom,6eth > 2btc,5atom,3eth", sdk.Coins{twoAtom, sixEth}, sdk.Coins{twoBtc, fiveAtom, threeEth}, true},
		{"2btc,2atom,3eth ≤ 5atom,6eth", sdk.Coins{twoBtc, twoAtom, threeEth}, sdk.Coins{fiveAtom, sixEth}, false},
		{"2atom,6eth ≤ 2btc,5atom", sdk.Coins{twoAtom, sixEth}, sdk.Coins{twoBtc, fiveAtom}, false},
	}

	for _, tc := range tests {
		s.Require().True(tc.expPass == tc.coinsA.IsAnyGT(tc.coinsB), tc.name)
	}
}

func (s *coinTestSuite) TestCoinsIsAnyNil() {
	twoAtom := sdk.NewInt64Coin("atom", 2)
	fiveAtom := sdk.NewInt64Coin("atom", 5)
	threeEth := sdk.NewInt64Coin("eth", 3)
	nilAtom := sdk.Coin{Denom: "atom"}

	s.Require().True(sdk.Coins{twoAtom, fiveAtom, threeEth, nilAtom}.IsAnyNil())
	s.Require().True(sdk.Coins{twoAtom, nilAtom, fiveAtom, threeEth}.IsAnyNil())
	s.Require().True(sdk.Coins{nilAtom, twoAtom, fiveAtom, threeEth}.IsAnyNil())
	s.Require().False(sdk.Coins{twoAtom, fiveAtom, threeEth}.IsAnyNil())

}

func (s *coinTestSuite) TestMarshalJSONCoins() {
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)

	testCases := []struct {
		name      string
		input     sdk.Coins
		strOutput string
	}{
		{"nil coins", nil, `[]`},
		{"empty coins", sdk.Coins{}, `[]`},
		{"non-empty coins", sdk.NewCoins(sdk.NewInt64Coin("foo", 50)), `[{"denom":"foo","amount":"50"}]`},
	}

	for _, tc := range testCases {
		bz, err := cdc.MarshalJSON(tc.input)
		s.Require().NoError(err)
		s.Require().Equal(tc.strOutput, string(bz))

		var newCoins sdk.Coins
		s.Require().NoError(cdc.UnmarshalJSON(bz, &newCoins))

		if tc.input.Empty() {
			s.Require().Nil(newCoins)
		} else {
			s.Require().Equal(tc.input, newCoins)
		}
	}
}

func (s *coinTestSuite) TestCoinAminoEncoding() {
	cdc := codec.NewLegacyAmino()
	c := sdk.NewInt64Coin(testDenom1, 5)

	bz1, err := cdc.MarshalBinaryBare(c)
	s.Require().NoError(err)

	bz2, err := cdc.MarshalBinaryLengthPrefixed(c)
	s.Require().NoError(err)

	bz3, err := c.Marshal()
	s.Require().NoError(err)
	s.Require().Equal(bz1, bz3)
	s.Require().Equal(bz2[1:], bz3)
}
