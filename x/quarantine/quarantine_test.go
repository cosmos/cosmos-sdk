package quarantine_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	. "github.com/cosmos/cosmos-sdk/x/quarantine"
	. "github.com/cosmos/cosmos-sdk/x/quarantine/testutil"
)

type coinMaker func() sdk.Coins

var (
	coinMakerOK    coinMaker = func() sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin("okcoin", 100)) }
	coinMakerMulti coinMaker = func() sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin("multicoina", 33), sdk.NewInt64Coin("multicoinb", 67))
	}
	coinMakerEmpty coinMaker = func() sdk.Coins { return sdk.Coins{} }
	coinMakerNil   coinMaker = func() sdk.Coins { return nil }
	coinMakerBad   coinMaker = func() sdk.Coins { return sdk.Coins{sdk.Coin{Denom: "badcoin", Amount: sdk.NewInt(-1)}} }
)

func TestContainsAddress(t *testing.T) {
	// Technically, if containsAddress breaks, a lot of other tests should also break,
	// but I figure it's better safe than sorry.
	addrShort0 := MakeTestAddr("cs", 0)
	addrShort1 := MakeTestAddr("cs", 1)
	addrLong2 := MakeLongAddr("cs", 2)
	addrLong3 := MakeLongAddr("cs", 3)
	addrEmpty := make(sdk.AccAddress, 0)
	addrShort0Almost := MakeCopyOfAccAddress(addrShort0)
	addrShort0Almost[len(addrShort0Almost)-1]++
	addr2Almost := MakeCopyOfAccAddress(addrLong2)
	addr2Almost[len(addr2Almost)-1]++

	tests := []struct {
		name       string
		addrs      []sdk.AccAddress
		addrToFind sdk.AccAddress
		expected   bool
	}{
		{
			name:       "nil | nil",
			addrs:      nil,
			addrToFind: nil,
			expected:   false,
		},
		{
			name:       "nil | empty addr",
			addrs:      nil,
			addrToFind: addrEmpty,
			expected:   false,
		},
		{
			name:       "nil | short",
			addrs:      nil,
			addrToFind: addrShort0,
			expected:   false,
		},
		{
			name:       "nil | long",
			addrs:      nil,
			addrToFind: addrLong2,
			expected:   false,
		},
		{
			name:       "empty addr | empty addr",
			addrs:      []sdk.AccAddress{addrEmpty},
			addrToFind: addrEmpty,
			expected:   true,
		},
		{
			name:       "empty | nil",
			addrs:      []sdk.AccAddress{},
			addrToFind: nil,
			expected:   false,
		},
		{
			name:       "empty | empty addr",
			addrs:      []sdk.AccAddress{},
			addrToFind: addrEmpty,
			expected:   false,
		},
		{
			name:       "empty | short",
			addrs:      []sdk.AccAddress{},
			addrToFind: addrShort0,
			expected:   false,
		},
		{
			name:       "empty | long",
			addrs:      []sdk.AccAddress{},
			addrToFind: addrLong2,
			expected:   false,
		},
		{
			name:       "short0 | nil",
			addrs:      []sdk.AccAddress{addrShort0},
			addrToFind: nil,
			expected:   false,
		},
		{
			name:       "short0 | empty addr",
			addrs:      []sdk.AccAddress{addrShort0},
			addrToFind: addrEmpty,
			expected:   false,
		},
		{
			name:       "short0 | short0",
			addrs:      []sdk.AccAddress{addrShort0},
			addrToFind: addrShort0,
			expected:   true,
		},
		{
			name:       "short0 | short0 almost",
			addrs:      []sdk.AccAddress{addrShort0},
			addrToFind: addrShort0Almost,
			expected:   false,
		},
		{
			name:       "short0 | short1",
			addrs:      []sdk.AccAddress{addrShort0},
			addrToFind: addrShort1,
			expected:   false,
		},
		{
			name:       "short0 | long",
			addrs:      []sdk.AccAddress{addrShort0},
			addrToFind: addrLong2,
			expected:   false,
		},
		{
			name:       "long2 | nil",
			addrs:      []sdk.AccAddress{addrLong2},
			addrToFind: nil,
			expected:   false,
		},
		{
			name:       "long2 | empty addr",
			addrs:      []sdk.AccAddress{addrLong2},
			addrToFind: addrEmpty,
			expected:   false,
		},
		{
			name:       "long2 | long2",
			addrs:      []sdk.AccAddress{addrLong2},
			addrToFind: addrLong2,
			expected:   true,
		},
		{
			name:       "long2 | long2 almost",
			addrs:      []sdk.AccAddress{addrLong2},
			addrToFind: addr2Almost,
			expected:   false,
		},
		{
			name:       "long2 | long3",
			addrs:      []sdk.AccAddress{addrLong2},
			addrToFind: addrLong3,
			expected:   false,
		},
		{
			name:       "long2 | short",
			addrs:      []sdk.AccAddress{addrLong2},
			addrToFind: addrShort0,
			expected:   false,
		},
		{
			name:       "short0 long3 short1 long2 | empty",
			addrs:      []sdk.AccAddress{addrShort0, addrLong3, addrShort1, addrLong2},
			addrToFind: addrEmpty,
			expected:   false,
		},
		{
			name:       "short0 long3 short1 long2 | short0",
			addrs:      []sdk.AccAddress{addrShort0, addrLong3, addrShort1, addrLong2},
			addrToFind: addrShort0,
			expected:   true,
		},
		{
			name:       "short0 long3 short1 long2 | short1",
			addrs:      []sdk.AccAddress{addrShort0, addrLong3, addrShort1, addrLong2},
			addrToFind: addrShort1,
			expected:   true,
		},
		{
			name:       "short0 long3 short1 long2 | long2",
			addrs:      []sdk.AccAddress{addrShort0, addrLong3, addrShort1, addrLong2},
			addrToFind: addrLong2,
			expected:   true,
		},
		{
			name:       "short0 long3 short1 long2 | long3",
			addrs:      []sdk.AccAddress{addrShort0, addrLong3, addrShort1, addrLong2},
			addrToFind: addrLong3,
			expected:   true,
		},
		{
			name:       "short0 long3 short1 long2 | short0 almost",
			addrs:      []sdk.AccAddress{addrShort0, addrLong3, addrShort1, addrLong2},
			addrToFind: addrShort0Almost,
			expected:   false,
		},
		{
			name:       "short0 long3 short1 long2 | long2 almost",
			addrs:      []sdk.AccAddress{addrShort0, addrLong3, addrShort1, addrLong2},
			addrToFind: addr2Almost,
			expected:   false,
		},
		{
			name:       "long3 empty long3 | short1",
			addrs:      []sdk.AccAddress{addrLong3, addrEmpty, addrLong3},
			addrToFind: addrShort1,
			expected:   false,
		},
		{
			name:       "long3 empty long3 | long2",
			addrs:      []sdk.AccAddress{addrLong3, addrEmpty, addrLong3},
			addrToFind: addrLong2,
			expected:   false,
		},
		{
			name:       "long3 empty long3 | long3",
			addrs:      []sdk.AccAddress{addrLong3, addrEmpty, addrLong3},
			addrToFind: addrLong3,
			expected:   true,
		},
		{
			name:       "long3 empty long3 | empty",
			addrs:      []sdk.AccAddress{addrLong3, addrEmpty, addrLong3},
			addrToFind: addrEmpty,
			expected:   true,
		},
		{
			name:       "long3 empty long3 | nil",
			addrs:      []sdk.AccAddress{addrLong3, addrEmpty, addrLong3},
			addrToFind: nil,
			expected:   true,
		},
		{
			name:       "short0 almost short0 almost long2 almost | short0",
			addrs:      []sdk.AccAddress{addrShort0Almost, addrShort0Almost, addr2Almost},
			addrToFind: addrShort0,
			expected:   false,
		},
		{
			name:       "short0 almost short0 almost long2 almost | short0 almost",
			addrs:      []sdk.AccAddress{addrShort0Almost, addrShort0Almost, addr2Almost},
			addrToFind: addrShort0Almost,
			expected:   true,
		},
		{
			name:       "short0 almost short0 almost long2 almost | long2",
			addrs:      []sdk.AccAddress{addrShort0Almost, addrShort0Almost, addr2Almost},
			addrToFind: addrLong2,
			expected:   false,
		},
		{
			name:       "short0 almost short0 almost long2 almost | long2 almost",
			addrs:      []sdk.AccAddress{addrShort0Almost, addrShort0Almost, addr2Almost},
			addrToFind: addr2Almost,
			expected:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origSuffixes := MakeCopyOfAccAddresses(tc.addrs)
			origSuffixToFind := MakeCopyOfAccAddress(tc.addrToFind)

			actual := ContainsAddress(tc.addrs, tc.addrToFind)
			assert.Equal(t, tc.expected, actual, "containsSuffix result")
			assert.Equal(t, origSuffixes, tc.addrs, "addrs before and after containsSuffix")
			assert.Equal(t, origSuffixToFind, tc.addrToFind, "addrToFind before and after containsSuffix")
		})
	}
}

func TestFindAddresses(t *testing.T) {
	addr0 := MakeTestAddr("fa", 0)
	addr1 := MakeTestAddr("fa", 1)
	addr2 := MakeTestAddr("fa", 2)
	addr3 := MakeTestAddr("fa", 3)
	addr4 := MakeTestAddr("fa", 4)
	addr5 := MakeTestAddr("fa", 5)

	tests := []struct {
		name        string
		allAddrs    []sdk.AccAddress
		addrsToFind []sdk.AccAddress
		found       []sdk.AccAddress
		leftover    []sdk.AccAddress
	}{
		{
			name:        "nil nil",
			allAddrs:    nil,
			addrsToFind: nil,
			found:       nil,
			leftover:    nil,
		},
		{
			name:        "nil empty",
			allAddrs:    nil,
			addrsToFind: []sdk.AccAddress{},
			found:       nil,
			leftover:    nil,
		},
		{
			name:        "empty nil",
			allAddrs:    []sdk.AccAddress{},
			addrsToFind: nil,
			found:       nil,
			leftover:    nil,
		},
		{
			name:        "empty empty",
			allAddrs:    []sdk.AccAddress{},
			addrsToFind: []sdk.AccAddress{},
			found:       nil,
			leftover:    nil,
		},
		{
			name:        "two nil",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: nil,
			found:       nil,
			leftover:    []sdk.AccAddress{addr0, addr1},
		},
		{
			name:        "two empty",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{},
			found:       nil,
			leftover:    []sdk.AccAddress{addr0, addr1},
		},
		{
			name:        "two first",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{addr0},
			found:       []sdk.AccAddress{addr0},
			leftover:    []sdk.AccAddress{addr1},
		},
		{
			name:        "two second",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{addr1},
			found:       []sdk.AccAddress{addr1},
			leftover:    []sdk.AccAddress{addr0},
		},
		{
			name:        "two other",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{addr2},
			found:       nil,
			leftover:    []sdk.AccAddress{addr0, addr1},
		},
		{
			name:        "two first second",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{addr0, addr1},
			found:       []sdk.AccAddress{addr0, addr1},
			leftover:    nil,
		},
		{
			name:        "two second first",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{addr1, addr0},
			found:       []sdk.AccAddress{addr0, addr1},
			leftover:    nil,
		},
		{
			name:        "two first second first",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{addr0, addr1, addr0},
			found:       []sdk.AccAddress{addr0, addr1},
			leftover:    nil,
		},
		{
			name:        "two first first",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{addr0, addr0},
			found:       []sdk.AccAddress{addr0},
			leftover:    []sdk.AccAddress{addr1},
		},
		{
			name:        "two second second",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{addr1, addr1},
			found:       []sdk.AccAddress{addr1},
			leftover:    []sdk.AccAddress{addr0},
		},
		{
			name:        "two first other",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{addr0, addr2},
			found:       []sdk.AccAddress{addr0},
			leftover:    []sdk.AccAddress{addr1},
		},
		{
			name:        "two second other",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{addr1, addr2},
			found:       []sdk.AccAddress{addr1},
			leftover:    []sdk.AccAddress{addr0},
		},
		{
			name:        "two other first",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{addr2, addr0},
			found:       []sdk.AccAddress{addr0},
			leftover:    []sdk.AccAddress{addr1},
		},
		{
			name:        "two other second",
			allAddrs:    []sdk.AccAddress{addr0, addr1},
			addrsToFind: []sdk.AccAddress{addr2, addr1},
			found:       []sdk.AccAddress{addr1},
			leftover:    []sdk.AccAddress{addr0},
		},
		{
			name:        "four other third other",
			allAddrs:    []sdk.AccAddress{addr0, addr1, addr2, addr3},
			addrsToFind: []sdk.AccAddress{addr4, addr2, addr5},
			found:       []sdk.AccAddress{addr2},
			leftover:    []sdk.AccAddress{addr0, addr1, addr3},
		},
		{
			name:        "dups in allAddrs",
			allAddrs:    []sdk.AccAddress{addr0, addr1, addr0, addr1, addr2},
			addrsToFind: []sdk.AccAddress{addr0, addr1},
			found:       []sdk.AccAddress{addr0, addr1, addr0, addr1},
			leftover:    []sdk.AccAddress{addr2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			allAddrsOrig := MakeCopyOfAccAddresses(tc.allAddrs)
			addrsToFindOrig := MakeCopyOfAccAddresses(tc.addrsToFind)
			found, leftover := FindAddresses(tc.allAddrs, tc.addrsToFind)
			assert.Equal(t, tc.found, found, "found")
			assert.Equal(t, tc.leftover, leftover, "leftover")
			assert.Equal(t, allAddrsOrig, tc.allAddrs, "allAddrs before and after findAddresses")
			assert.Equal(t, addrsToFindOrig, tc.addrsToFind, "addrsToFindOrig before and after findAddresses")
		})
	}
}

func TestContainsSuffix(t *testing.T) {
	// Technically, if containsSuffix breaks, a lot of other tests should also break,
	// but I figure it's better safe than sorry.
	suffixShort0 := []byte(MakeTestAddr("cs", 0))
	suffixShort1 := []byte(MakeTestAddr("cs", 1))
	suffixLong2 := []byte(MakeLongAddr("cs", 2))
	suffixLong3 := []byte(MakeLongAddr("cs", 3))
	suffixEmpty := make([]byte, 0)
	suffixShort0Almost := MakeCopyOfByteSlice(suffixShort0)
	suffixShort0Almost[len(suffixShort0Almost)-1]++
	suffixLong2Almost := MakeCopyOfByteSlice(suffixLong2)
	suffixLong2Almost[len(suffixLong2Almost)-1]++

	tests := []struct {
		name         string
		suffixes     [][]byte
		suffixToFind []byte
		expected     bool
	}{
		{
			name:         "nil | nil",
			suffixes:     nil,
			suffixToFind: nil,
			expected:     false,
		},
		{
			name:         "nil | empty suffix",
			suffixes:     nil,
			suffixToFind: suffixEmpty,
			expected:     false,
		},
		{
			name:         "nil | short",
			suffixes:     nil,
			suffixToFind: suffixShort0,
			expected:     false,
		},
		{
			name:         "nil | long",
			suffixes:     nil,
			suffixToFind: suffixLong2,
			expected:     false,
		},
		{
			name:         "empty suffix | empty suffix",
			suffixes:     [][]byte{suffixEmpty},
			suffixToFind: suffixEmpty,
			expected:     true,
		},
		{
			name:         "empty | nil",
			suffixes:     [][]byte{},
			suffixToFind: nil,
			expected:     false,
		},
		{
			name:         "empty | empty suffix",
			suffixes:     [][]byte{},
			suffixToFind: suffixEmpty,
			expected:     false,
		},
		{
			name:         "empty | short",
			suffixes:     [][]byte{},
			suffixToFind: suffixShort0,
			expected:     false,
		},
		{
			name:         "empty | long",
			suffixes:     [][]byte{},
			suffixToFind: suffixLong2,
			expected:     false,
		},
		{
			name:         "short0 | nil",
			suffixes:     [][]byte{suffixShort0},
			suffixToFind: nil,
			expected:     false,
		},
		{
			name:         "short0 | empty suffix",
			suffixes:     [][]byte{suffixShort0},
			suffixToFind: suffixEmpty,
			expected:     false,
		},
		{
			name:         "short0 | short0",
			suffixes:     [][]byte{suffixShort0},
			suffixToFind: suffixShort0,
			expected:     true,
		},
		{
			name:         "short0 | short0 almost",
			suffixes:     [][]byte{suffixShort0},
			suffixToFind: suffixShort0Almost,
			expected:     false,
		},
		{
			name:         "short0 | short1",
			suffixes:     [][]byte{suffixShort0},
			suffixToFind: suffixShort1,
			expected:     false,
		},
		{
			name:         "short0 | long",
			suffixes:     [][]byte{suffixShort0},
			suffixToFind: suffixLong2,
			expected:     false,
		},
		{
			name:         "long2 | nil",
			suffixes:     [][]byte{suffixLong2},
			suffixToFind: nil,
			expected:     false,
		},
		{
			name:         "long2 | empty suffix",
			suffixes:     [][]byte{suffixLong2},
			suffixToFind: suffixEmpty,
			expected:     false,
		},
		{
			name:         "long2 | long2",
			suffixes:     [][]byte{suffixLong2},
			suffixToFind: suffixLong2,
			expected:     true,
		},
		{
			name:         "long2 | long2 almost",
			suffixes:     [][]byte{suffixLong2},
			suffixToFind: suffixLong2Almost,
			expected:     false,
		},
		{
			name:         "long2 | long3",
			suffixes:     [][]byte{suffixLong2},
			suffixToFind: suffixLong3,
			expected:     false,
		},
		{
			name:         "long2 | short",
			suffixes:     [][]byte{suffixLong2},
			suffixToFind: suffixShort0,
			expected:     false,
		},
		{
			name:         "short0 long3 short1 long2 | empty",
			suffixes:     [][]byte{suffixShort0, suffixLong3, suffixShort1, suffixLong2},
			suffixToFind: suffixEmpty,
			expected:     false,
		},
		{
			name:         "short0 long3 short1 long2 | short0",
			suffixes:     [][]byte{suffixShort0, suffixLong3, suffixShort1, suffixLong2},
			suffixToFind: suffixShort0,
			expected:     true,
		},
		{
			name:         "short0 long3 short1 long2 | short1",
			suffixes:     [][]byte{suffixShort0, suffixLong3, suffixShort1, suffixLong2},
			suffixToFind: suffixShort1,
			expected:     true,
		},
		{
			name:         "short0 long3 short1 long2 | long2",
			suffixes:     [][]byte{suffixShort0, suffixLong3, suffixShort1, suffixLong2},
			suffixToFind: suffixLong2,
			expected:     true,
		},
		{
			name:         "short0 long3 short1 long2 | long3",
			suffixes:     [][]byte{suffixShort0, suffixLong3, suffixShort1, suffixLong2},
			suffixToFind: suffixLong3,
			expected:     true,
		},
		{
			name:         "short0 long3 short1 long2 | short0 almost",
			suffixes:     [][]byte{suffixShort0, suffixLong3, suffixShort1, suffixLong2},
			suffixToFind: suffixShort0Almost,
			expected:     false,
		},
		{
			name:         "short0 long3 short1 long2 | long2 almost",
			suffixes:     [][]byte{suffixShort0, suffixLong3, suffixShort1, suffixLong2},
			suffixToFind: suffixLong2Almost,
			expected:     false,
		},
		{
			name:         "long3 empty long3 | short1",
			suffixes:     [][]byte{suffixLong3, suffixEmpty, suffixLong3},
			suffixToFind: suffixShort1,
			expected:     false,
		},
		{
			name:         "long3 empty long3 | long2",
			suffixes:     [][]byte{suffixLong3, suffixEmpty, suffixLong3},
			suffixToFind: suffixLong2,
			expected:     false,
		},
		{
			name:         "long3 empty long3 | long3",
			suffixes:     [][]byte{suffixLong3, suffixEmpty, suffixLong3},
			suffixToFind: suffixLong3,
			expected:     true,
		},
		{
			name:         "long3 empty long3 | empty",
			suffixes:     [][]byte{suffixLong3, suffixEmpty, suffixLong3},
			suffixToFind: suffixEmpty,
			expected:     true,
		},
		{
			name:         "long3 empty long3 | nil",
			suffixes:     [][]byte{suffixLong3, suffixEmpty, suffixLong3},
			suffixToFind: nil,
			expected:     true,
		},
		{
			name:         "short0 almost short0 almost long2 almost | short0",
			suffixes:     [][]byte{suffixShort0Almost, suffixShort0Almost, suffixLong2Almost},
			suffixToFind: suffixShort0,
			expected:     false,
		},
		{
			name:         "short0 almost short0 almost long2 almost | short0 almost",
			suffixes:     [][]byte{suffixShort0Almost, suffixShort0Almost, suffixLong2Almost},
			suffixToFind: suffixShort0Almost,
			expected:     true,
		},
		{
			name:         "short0 almost short0 almost long2 almost | long2",
			suffixes:     [][]byte{suffixShort0Almost, suffixShort0Almost, suffixLong2Almost},
			suffixToFind: suffixLong2,
			expected:     false,
		},
		{
			name:         "short0 almost short0 almost long2 almost | long2 almost",
			suffixes:     [][]byte{suffixShort0Almost, suffixShort0Almost, suffixLong2Almost},
			suffixToFind: suffixLong2Almost,
			expected:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origSuffixes := MakeCopyOfByteSliceSlice(tc.suffixes)
			origSuffixToFind := MakeCopyOfByteSlice(tc.suffixToFind)

			actual := ContainsSuffix(tc.suffixes, tc.suffixToFind)
			assert.Equal(t, tc.expected, actual, "containsSuffix result")
			assert.Equal(t, origSuffixes, tc.suffixes, "suffixes before and after containsSuffix")
			assert.Equal(t, origSuffixToFind, tc.suffixToFind, "suffixToFind before and after containsSuffix")
		})
	}
}

func TestNewQuarantinedFunds(t *testing.T) {
	testAddr0 := MakeTestAddr("nqf", 0)
	testAddr1 := MakeTestAddr("nqf", 1)

	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []sdk.AccAddress
		Coins     sdk.Coins
		declined  bool
		expected  *QuarantinedFunds
	}{
		{
			name:      "control",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{testAddr1},
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 88)),
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 88)),
				Declined:                false,
			},
		},
		{
			name:      "declined true",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{testAddr1},
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 87)),
			declined:  true,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 87)),
				Declined:                true,
			},
		},
		{
			name:      "nil toAddr",
			toAddr:    nil,
			fromAddrs: []sdk.AccAddress{testAddr1},
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 86)),
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               "",
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 86)),
				Declined:                false,
			},
		},
		{
			name:      "nil fromAddrs",
			toAddr:    testAddr0,
			fromAddrs: nil,
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
				Declined:                false,
			},
		},
		{
			name:      "empty fromAddrs",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{},
			Coins:     sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("rando", 85)),
				Declined:                false,
			},
		},
		{
			name:      "empty coins",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{testAddr1},
			Coins:     sdk.Coins{},
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   sdk.Coins{},
				Declined:                false,
			},
		},
		{
			name:      "nil coins",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{testAddr1},
			Coins:     nil,
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   nil,
				Declined:                false,
			},
		},
		{
			name:      "invalid coins",
			toAddr:    testAddr0,
			fromAddrs: []sdk.AccAddress{testAddr1},
			Coins:     sdk.Coins{sdk.Coin{Denom: "bad", Amount: sdk.NewInt(-1)}},
			declined:  false,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   sdk.Coins{sdk.Coin{Denom: "bad", Amount: sdk.NewInt(-1)}},
				Declined:                false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewQuarantinedFunds(tc.toAddr, tc.fromAddrs, tc.Coins, tc.declined)
			assert.Equal(t, tc.expected, actual, "NewQuarantinedFunds")
		})
	}
}

func TestQuarantinedFunds_Validate(t *testing.T) {
	testAddr0 := MakeTestAddr("qfv", 0).String()
	testAddr1 := MakeTestAddr("qfv", 1).String()
	testAddr2 := MakeTestAddr("qfv", 2).String()

	tests := []struct {
		name          string
		qf            *QuarantinedFunds
		expectedInErr []string
	}{
		{
			name: "control",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "declined true",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			expectedInErr: nil,
		},
		{
			name: "bad to address",
			qf: &QuarantinedFunds{
				ToAddress:               "notgonnawork",
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"invalid to address"},
		},
		{
			name: "empty to address",
			qf: &QuarantinedFunds{
				ToAddress:               "",
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"invalid to address"},
		},
		{
			name: "bad from address",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{"alsonotgood"},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"invalid unaccepted from address[0]"},
		},
		{
			name: "empty from address",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{""},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"invalid unaccepted from address[0]"},
		},
		{
			name: "nil from addresses",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"at least one unaccepted from address is required", "invalid value"},
		},
		{
			name: "empty from addresses",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"at least one unaccepted from address is required", "invalid value"},
		},
		{
			name: "two from addresses both good",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1, testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "two same from addresses",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr2, testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"duplicate unaccepted from address", testAddr2},
		},
		{
			name: "three from addresses same first last",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1, testAddr2, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"duplicate unaccepted from address", testAddr1},
		},
		{
			name: "two from addresses first bad",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{"this is not an address", testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"invalid unaccepted from address[0]"},
		},
		{
			name: "two from addresses last bad",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1, "this is also bad"},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"invalid unaccepted from address[1]"},
		},
		{
			name: "empty coins",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerEmpty(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "nil coins",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerNil(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "bad coins",
			qf: &QuarantinedFunds{
				ToAddress:               testAddr0,
				UnacceptedFromAddresses: []string{testAddr1},
				Coins:                   coinMakerBad(),
				Declined:                false,
			},
			expectedInErr: []string{coinMakerBad().String(), "amount is not positive"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qfOrig := MakeCopyOfQuarantinedFunds(tc.qf)
			err := tc.qf.Validate()
			AssertErrorContents(t, err, tc.expectedInErr, "Validate")
			assert.Equal(t, qfOrig, tc.qf, "QuarantinedFunds before and after")
		})
	}
}

func TestNewAutoResponseEntry(t *testing.T) {
	testAddr0 := MakeTestAddr("nare", 0)
	testAddr1 := MakeTestAddr("nare", 1)

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		fromAddr sdk.AccAddress
		resp     AutoResponse
		expected *AutoResponseEntry
	}{
		{
			name:     "accept",
			toAddr:   testAddr0,
			fromAddr: testAddr1,
			resp:     AUTO_RESPONSE_ACCEPT,
			expected: &AutoResponseEntry{
				ToAddress:   testAddr0.String(),
				FromAddress: testAddr1.String(),
				Response:    AUTO_RESPONSE_ACCEPT,
			},
		},
		{
			name:     "decline",
			toAddr:   testAddr0,
			fromAddr: testAddr1,
			resp:     AUTO_RESPONSE_DECLINE,
			expected: &AutoResponseEntry{
				ToAddress:   testAddr0.String(),
				FromAddress: testAddr1.String(),
				Response:    AUTO_RESPONSE_DECLINE,
			},
		},
		{
			name:     "unspecified",
			toAddr:   testAddr0,
			fromAddr: testAddr1,
			resp:     AUTO_RESPONSE_UNSPECIFIED,
			expected: &AutoResponseEntry{
				ToAddress:   testAddr0.String(),
				FromAddress: testAddr1.String(),
				Response:    AUTO_RESPONSE_UNSPECIFIED,
			},
		},
		{
			name:     "nil to address",
			toAddr:   nil,
			fromAddr: testAddr1,
			resp:     AUTO_RESPONSE_ACCEPT,
			expected: &AutoResponseEntry{
				ToAddress:   "",
				FromAddress: testAddr1.String(),
				Response:    AUTO_RESPONSE_ACCEPT,
			},
		},
		{
			name:     "nil from address",
			toAddr:   testAddr0,
			fromAddr: nil,
			resp:     AUTO_RESPONSE_DECLINE,
			expected: &AutoResponseEntry{
				ToAddress:   testAddr0.String(),
				FromAddress: "",
				Response:    AUTO_RESPONSE_DECLINE,
			},
		},
		{
			name:     "weird response",
			toAddr:   testAddr1,
			fromAddr: testAddr0,
			resp:     -3,
			expected: &AutoResponseEntry{
				ToAddress:   testAddr1.String(),
				FromAddress: testAddr0.String(),
				Response:    -3,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewAutoResponseEntry(tc.toAddr, tc.fromAddr, tc.resp)
			assert.Equal(t, tc.expected, actual, "NewAutoResponseEntry")
		})
	}
}

func TestAutoResponseEntry_Validate(t *testing.T) {
	testAddr0 := MakeTestAddr("arev", 0).String()
	testAddr1 := MakeTestAddr("arev", 1).String()

	tests := []struct {
		name          string
		toAddr        string
		fromAddr      string
		resp          AutoResponse
		qf            QuarantinedFunds
		expectedInErr []string
	}{
		{
			name:          "accept",
			toAddr:        testAddr0,
			fromAddr:      testAddr1,
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: nil,
		},
		{
			name:          "decline",
			toAddr:        testAddr0,
			fromAddr:      testAddr1,
			resp:          AUTO_RESPONSE_DECLINE,
			expectedInErr: nil,
		},
		{
			name:          "unspecified",
			toAddr:        testAddr0,
			fromAddr:      testAddr1,
			resp:          AUTO_RESPONSE_UNSPECIFIED,
			expectedInErr: nil,
		},
		{
			name:          "bad to address",
			toAddr:        "notgonnawork",
			fromAddr:      testAddr1,
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "empty to address",
			toAddr:        "",
			fromAddr:      testAddr1,
			resp:          AUTO_RESPONSE_DECLINE,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "bad from address",
			toAddr:        testAddr0,
			fromAddr:      "alsonotgood",
			resp:          AUTO_RESPONSE_UNSPECIFIED,
			expectedInErr: []string{"invalid from address"},
		},
		{
			name:          "empty from address",
			toAddr:        testAddr0,
			fromAddr:      "",
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: []string{"invalid from address"},
		},
		{
			name:          "negative response",
			toAddr:        testAddr0,
			fromAddr:      testAddr1,
			resp:          -1,
			expectedInErr: []string{"unknown auto-response value", "-1"},
		},
		{
			name:          "response too large",
			toAddr:        testAddr0,
			fromAddr:      testAddr1,
			resp:          3,
			expectedInErr: []string{"unknown auto-response value", "3"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entryOrig := AutoResponseEntry{
				ToAddress:   tc.toAddr,
				FromAddress: tc.fromAddr,
				Response:    tc.resp,
			}
			entry := AutoResponseEntry{
				ToAddress:   tc.toAddr,
				FromAddress: tc.fromAddr,
				Response:    tc.resp,
			}
			err := entry.Validate()
			AssertErrorContents(t, err, tc.expectedInErr, "Validate")
			assert.Equal(t, entryOrig, entry, "AutoResponseEntry before and after")
		})
	}
}

func TestAutoResponseUpdate_Validate(t *testing.T) {
	testAddr0 := MakeTestAddr("arev", 0).String()
	testAddr1 := MakeTestAddr("arev", 1).String()

	tests := []struct {
		name          string
		fromAddr      string
		resp          AutoResponse
		qf            QuarantinedFunds
		expectedInErr []string
	}{
		{
			name:          "accept",
			fromAddr:      testAddr0,
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: nil,
		},
		{
			name:          "decline",
			fromAddr:      testAddr1,
			resp:          AUTO_RESPONSE_DECLINE,
			expectedInErr: nil,
		},
		{
			name:          "unspecified",
			fromAddr:      testAddr0,
			resp:          AUTO_RESPONSE_UNSPECIFIED,
			expectedInErr: nil,
		},
		{
			name:          "bad from address",
			fromAddr:      "yupnotgood",
			resp:          AUTO_RESPONSE_UNSPECIFIED,
			expectedInErr: []string{"invalid from address"},
		},
		{
			name:          "empty from address",
			fromAddr:      "",
			resp:          AUTO_RESPONSE_ACCEPT,
			expectedInErr: []string{"invalid from address"},
		},
		{
			name:          "negative response",
			fromAddr:      testAddr1,
			resp:          -1,
			expectedInErr: []string{"unknown auto-response value", "-1"},
		},
		{
			name:          "response too large",
			fromAddr:      testAddr0,
			resp:          3,
			expectedInErr: []string{"unknown auto-response value", "3"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updateOrig := AutoResponseUpdate{
				FromAddress: tc.fromAddr,
				Response:    tc.resp,
			}
			update := AutoResponseUpdate{
				FromAddress: tc.fromAddr,
				Response:    tc.resp,
			}
			err := update.Validate()
			AssertErrorContents(t, err, tc.expectedInErr, "Validate")
			assert.Equal(t, updateOrig, update, "AutoResponseUpdate before and after")
		})
	}
}

func TestAutoBValues(t *testing.T) {
	// If these were the same, it'd be bad.
	assert.NotEqual(t, NoAutoB, AutoAcceptB, "NoAutoB vs AutoAcceptB")
	assert.NotEqual(t, NoAutoB, AutoDeclineB, "NoAutoB vs AutoDeclineB")
	assert.NotEqual(t, AutoAcceptB, AutoDeclineB, "AutoAcceptB vs AutoDeclineB")
}

func TestToAutoB(t *testing.T) {
	tests := []struct {
		name     string
		r        AutoResponse
		expected byte
	}{
		{
			name:     "accept",
			r:        AUTO_RESPONSE_ACCEPT,
			expected: AutoAcceptB,
		},
		{
			name:     "decline",
			r:        AUTO_RESPONSE_DECLINE,
			expected: AutoDeclineB,
		},
		{
			name:     "unspecified",
			r:        AUTO_RESPONSE_UNSPECIFIED,
			expected: NoAutoB,
		},
		{
			name:     "negative",
			r:        -1,
			expected: NoAutoB,
		},
		{
			name:     "too large",
			r:        3,
			expected: NoAutoB,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := ToAutoB(tc.r)
			assert.Equal(t, tc.expected, actual, "ToAutoB(%s)", tc.r)
		})
	}
}

func TestAutoResponseValues(t *testing.T) {
	// If these were the same, it'd be bad.
	assert.NotEqual(t, AUTO_RESPONSE_UNSPECIFIED, AUTO_RESPONSE_ACCEPT, "AUTO_RESPONSE_UNSPECIFIED vs AUTO_RESPONSE_ACCEPT")
	assert.NotEqual(t, AUTO_RESPONSE_UNSPECIFIED, AUTO_RESPONSE_DECLINE, "AUTO_RESPONSE_UNSPECIFIED vs AUTO_RESPONSE_DECLINE")
	assert.NotEqual(t, AUTO_RESPONSE_ACCEPT, AUTO_RESPONSE_DECLINE, "AUTO_RESPONSE_ACCEPT vs AUTO_RESPONSE_DECLINE")
}

func TestToAutoResponse(t *testing.T) {
	tests := []struct {
		name     string
		bz       []byte
		expected AutoResponse
	}{
		{
			name:     "AutoAcceptB",
			bz:       []byte{AutoAcceptB},
			expected: AUTO_RESPONSE_ACCEPT,
		},
		{
			name:     "AutoDeclineB",
			bz:       []byte{AutoDeclineB},
			expected: AUTO_RESPONSE_DECLINE,
		},
		{
			name:     "NoAutoB",
			bz:       []byte{NoAutoB},
			expected: AUTO_RESPONSE_UNSPECIFIED,
		},
		{
			name:     "something else",
			bz:       []byte{0x7d},
			expected: AUTO_RESPONSE_UNSPECIFIED,
		},
		{
			name:     "nil",
			bz:       nil,
			expected: AUTO_RESPONSE_UNSPECIFIED,
		},
		{
			name:     "empty",
			bz:       []byte{},
			expected: AUTO_RESPONSE_UNSPECIFIED,
		},
		{
			name:     "too long",
			bz:       []byte{AutoAcceptB, AutoAcceptB},
			expected: AUTO_RESPONSE_UNSPECIFIED,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := ToAutoResponse(tc.bz)
			assert.Equal(t, tc.expected, actual, "ToAutoResponse(%v)", tc.bz)
		})
	}
}

func TestAutoResponse_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		r        AutoResponse
		expected bool
	}{
		{
			name:     "accept",
			r:        AUTO_RESPONSE_ACCEPT,
			expected: true,
		},
		{
			name:     "decline",
			r:        AUTO_RESPONSE_DECLINE,
			expected: true,
		},
		{
			name:     "unspecified",
			r:        AUTO_RESPONSE_UNSPECIFIED,
			expected: true,
		},
		{
			name:     "negative",
			r:        -1,
			expected: false,
		},
		{
			name:     "too large",
			r:        3,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.r
			actual := r.IsValid()
			assert.Equal(t, tc.expected, actual, "%s.IsValid", tc.r)
			assert.Equal(t, tc.r, r, "AutoResponse before and after")
		})
	}
}

func TestAutoResponse_IsAccept(t *testing.T) {
	tests := []struct {
		name     string
		r        AutoResponse
		expected bool
	}{
		{
			name:     "accept",
			r:        AUTO_RESPONSE_ACCEPT,
			expected: true,
		},
		{
			name:     "decline",
			r:        AUTO_RESPONSE_DECLINE,
			expected: false,
		},
		{
			name:     "unspecified",
			r:        AUTO_RESPONSE_UNSPECIFIED,
			expected: false,
		},
		{
			name:     "negative",
			r:        -1,
			expected: false,
		},
		{
			name:     "too large",
			r:        3,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.r
			actual := r.IsAccept()
			assert.Equal(t, tc.expected, actual, "%s.IsAccept", tc.r)
			assert.Equal(t, tc.r, r, "AutoResponse before and after")
		})
	}
}

func TestAutoResponse_IsDecline(t *testing.T) {
	tests := []struct {
		name     string
		r        AutoResponse
		expected bool
	}{
		{
			name:     "accept",
			r:        AUTO_RESPONSE_ACCEPT,
			expected: false,
		},
		{
			name:     "decline",
			r:        AUTO_RESPONSE_DECLINE,
			expected: true,
		},
		{
			name:     "unspecified",
			r:        AUTO_RESPONSE_UNSPECIFIED,
			expected: false,
		},
		{
			name:     "negative",
			r:        -1,
			expected: false,
		},
		{
			name:     "too large",
			r:        3,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.r
			actual := r.IsDecline()
			assert.Equal(t, tc.expected, actual, "%s.IsDecline", tc.r)
			assert.Equal(t, tc.r, r, "AutoResponse before and after")
		})
	}
}

func TestNewQuarantineRecord(t *testing.T) {
	testAddr0 := MakeTestAddr("nqr", 0)
	testAddr1 := MakeTestAddr("nqr", 1)

	tests := []struct {
		name        string
		uaFromAddrs []string
		coins       sdk.Coins
		declined    bool
		expected    *QuarantineRecord
		expPanic    string
	}{
		{
			name:        "control",
			uaFromAddrs: []string{testAddr0.String()},
			coins:       coinMakerOK(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name:        "declined",
			uaFromAddrs: []string{testAddr0.String()},
			coins:       coinMakerOK(),
			declined:    true,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name:        "multi coins",
			uaFromAddrs: []string{testAddr0.String()},
			coins:       coinMakerMulti(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerMulti(),
				Declined:                false,
			},
		},
		{
			name:        "empty coins",
			uaFromAddrs: []string{testAddr0.String()},
			coins:       coinMakerEmpty(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerEmpty(),
				Declined:                false,
			},
		},
		{
			name:        "nil coins",
			uaFromAddrs: []string{testAddr0.String()},
			coins:       coinMakerNil(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerNil(),
				Declined:                false,
			},
		},
		{
			name:        "bad coins",
			uaFromAddrs: []string{testAddr0.String()},
			coins:       coinMakerBad(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerBad(),
				Declined:                false,
			},
		},
		{
			name:        "bad unaccepted addr panics",
			uaFromAddrs: []string{"I'm a bad address"},
			coins:       coinMakerOK(),
			declined:    false,
			expPanic:    "decoding bech32 failed: string not all lowercase or all uppercase",
		},
		{
			name:        "empty unaccepted addr string panics",
			uaFromAddrs: []string{""},
			coins:       coinMakerOK(),
			declined:    false,
			expPanic:    "empty address string is not allowed",
		},
		{
			name:        "bad second unaccepted addr panics",
			uaFromAddrs: []string{testAddr0.String(), "I'm a bad address"},
			coins:       coinMakerOK(),
			declined:    false,
			expPanic:    "decoding bech32 failed: string not all lowercase or all uppercase",
		},
		{
			name:        "two unaccepted addresses",
			uaFromAddrs: []string{testAddr0.String(), testAddr1.String()},
			coins:       coinMakerOK(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name:        "empty unaccepted addresses",
			uaFromAddrs: []string{},
			coins:       coinMakerOK(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name:        "nil unaccepted addresses",
			uaFromAddrs: nil,
			coins:       coinMakerOK(),
			declined:    false,
			expected: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *QuarantineRecord
			testFunc := func() {
				actual = NewQuarantineRecord(tc.uaFromAddrs, tc.coins, tc.declined)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "NewQuarantineRecord") {
					assert.Equal(t, tc.expected, actual, "NewQuarantineRecord")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "NewQuarantineRecord")
			}
		})
	}
}

func TestQuarantineRecord_Validate(t *testing.T) {
	testAddr0 := MakeTestAddr("qrv", 0)
	testAddr1 := MakeTestAddr("qrv", 1)
	testAddr2 := MakeTestAddr("qrv", 2)

	tests := []struct {
		name          string
		qr            *QuarantineRecord
		expectedInErr []string
	}{
		{
			name: "control",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "declined",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			expectedInErr: nil,
		},
		{
			name: "no accepted addresses is ok",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "nil accepted addresses is ok",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "multi coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerMulti(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "empty coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerEmpty(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "nil coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerNil(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "bad coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerBad(),
				Declined:                false,
			},
			expectedInErr: []string{coinMakerBad().String(), "amount is not positive"},
		},
		{
			name: "nil unaccepted addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"at least one unaccepted from address is required"},
		},
		{
			name: "empty unaccepted addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: []string{"at least one unaccepted from address is required"},
		},
		{
			name: "two unaccepted addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
		{
			name: "three unaccepted addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1, testAddr2},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expectedInErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qrOrig := MakeCopyOfQuarantineRecord(tc.qr)
			err := tc.qr.Validate()
			AssertErrorContents(t, err, tc.expectedInErr, "Validate")
			assert.Equal(t, qrOrig, tc.qr, "QuarantineRecord before and after")
		})
	}
}

func TestQuarantineRecord_AddCoins(t *testing.T) {
	testAddr0 := MakeTestAddr("qrac", 0)
	testAddr1 := MakeTestAddr("qrac", 1)
	testAddr2 := MakeTestAddr("qrac", 2)
	testAddr3 := MakeTestAddr("qrac", 3)

	keyEmpty := "empty"
	keyNil := "nil"
	key0Acorn := "0acorn"
	key50Acorn := "50acorn"
	key32Almond := "32almond"
	key8acorn9Almond := "8acorn,9almond"
	coinMakerMap := map[string]coinMaker{
		keyEmpty:         coinMakerEmpty,
		keyNil:           coinMakerNil,
		key0Acorn:        func() sdk.Coins { return sdk.Coins{sdk.NewInt64Coin("acorn", 0)} },
		key50Acorn:       func() sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)) },
		key32Almond:      func() sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin("almond", 32)) },
		key8acorn9Almond: func() sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)) },
	}

	tests := []struct {
		qrCoinKey  string
		addCoinKey string
		expected   sdk.Coins
	}{
		// empty
		{
			qrCoinKey:  keyEmpty,
			addCoinKey: keyEmpty,
			expected:   nil,
		},
		{
			qrCoinKey:  keyEmpty,
			addCoinKey: keyNil,
			expected:   nil,
		},
		{
			qrCoinKey:  keyEmpty,
			addCoinKey: key0Acorn,
			expected:   nil,
		},
		{
			qrCoinKey:  keyEmpty,
			addCoinKey: key50Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  keyEmpty,
			addCoinKey: key32Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  keyEmpty,
			addCoinKey: key8acorn9Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},

		// nil
		{
			qrCoinKey:  keyNil,
			addCoinKey: keyEmpty,
			expected:   nil,
		},
		{
			qrCoinKey:  keyNil,
			addCoinKey: keyNil,
			expected:   nil,
		},
		{
			qrCoinKey:  keyNil,
			addCoinKey: key0Acorn,
			expected:   nil,
		},
		{
			qrCoinKey:  keyNil,
			addCoinKey: key50Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  keyNil,
			addCoinKey: key32Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  keyNil,
			addCoinKey: key8acorn9Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},

		// 0acorn
		{
			qrCoinKey:  key0Acorn,
			addCoinKey: keyEmpty,
			expected:   nil,
		},
		{
			qrCoinKey:  key0Acorn,
			addCoinKey: keyNil,
			expected:   nil,
		},
		{
			qrCoinKey:  key0Acorn,
			addCoinKey: key0Acorn,
			expected:   nil,
		},
		{
			qrCoinKey:  key0Acorn,
			addCoinKey: key50Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  key0Acorn,
			addCoinKey: key32Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  key0Acorn,
			addCoinKey: key8acorn9Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},

		// 50acorn
		{
			qrCoinKey:  key50Acorn,
			addCoinKey: keyEmpty,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  key50Acorn,
			addCoinKey: keyNil,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  key50Acorn,
			addCoinKey: key0Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50)),
		},
		{
			qrCoinKey:  key50Acorn,
			addCoinKey: key50Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 100)),
		},
		{
			qrCoinKey:  key50Acorn,
			addCoinKey: key32Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50), sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  key50Acorn,
			addCoinKey: key8acorn9Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 58), sdk.NewInt64Coin("almond", 9)),
		},

		// 32almond
		{
			qrCoinKey:  key32Almond,
			addCoinKey: keyEmpty,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  key32Almond,
			addCoinKey: keyNil,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  key32Almond,
			addCoinKey: key0Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  key32Almond,
			addCoinKey: key50Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 50), sdk.NewInt64Coin("almond", 32)),
		},
		{
			qrCoinKey:  key32Almond,
			addCoinKey: key32Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("almond", 64)),
		},
		{
			qrCoinKey:  key32Almond,
			addCoinKey: key8acorn9Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 41)),
		},

		// 8acorn,9almond
		{
			qrCoinKey:  key8acorn9Almond,
			addCoinKey: keyEmpty,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},
		{
			qrCoinKey:  key8acorn9Almond,
			addCoinKey: keyNil,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},
		{
			qrCoinKey:  key8acorn9Almond,
			addCoinKey: key0Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 9)),
		},
		{
			qrCoinKey:  key8acorn9Almond,
			addCoinKey: key50Acorn,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 58), sdk.NewInt64Coin("almond", 9)),
		},
		{
			qrCoinKey:  key8acorn9Almond,
			addCoinKey: key32Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("almond", 41)),
		},
		{
			qrCoinKey:  key8acorn9Almond,
			addCoinKey: key8acorn9Almond,
			expected:   sdk.NewCoins(sdk.NewInt64Coin("acorn", 16), sdk.NewInt64Coin("almond", 18)),
		},
	}

	addressCombos := []struct {
		name       string
		unaccepted []sdk.AccAddress
		accepted   []sdk.AccAddress
	}{
		{
			name:       "no addresses",
			unaccepted: nil,
			accepted:   nil,
		},
		{
			name:       "one unaccepted",
			unaccepted: []sdk.AccAddress{testAddr0},
			accepted:   nil,
		},
		{
			name:       "two unaccepted",
			unaccepted: []sdk.AccAddress{testAddr0, testAddr1},
			accepted:   nil,
		},
		{
			name:       "one accepted",
			unaccepted: nil,
			accepted:   []sdk.AccAddress{testAddr2},
		},
		{
			name:       "two accepted",
			unaccepted: nil,
			accepted:   []sdk.AccAddress{testAddr2, testAddr3},
		},
		{
			name:       "one unaccepted one accepted",
			unaccepted: []sdk.AccAddress{testAddr0},
			accepted:   []sdk.AccAddress{testAddr2},
		},
		{
			name:       "two unaccepted one accepted",
			unaccepted: []sdk.AccAddress{testAddr0, testAddr1},
			accepted:   []sdk.AccAddress{testAddr2},
		},
		{
			name:       "one unaccepted two accepted",
			unaccepted: []sdk.AccAddress{testAddr0},
			accepted:   []sdk.AccAddress{testAddr2, testAddr3},
		},
		{
			name:       "two unaccepted two accepted",
			unaccepted: []sdk.AccAddress{testAddr0, testAddr1},
			accepted:   []sdk.AccAddress{testAddr2, testAddr3},
		},
	}

	for _, tc := range tests {
		for _, ac := range addressCombos {
			for _, declined := range []bool{false, true} {
				name := fmt.Sprintf("%s+%s=%q %t %s", tc.qrCoinKey, tc.addCoinKey, tc.expected.String(), declined, ac.name)
				t.Run(name, func(t *testing.T) {
					expected := QuarantineRecord{
						UnacceptedFromAddresses: MakeCopyOfAccAddresses(ac.unaccepted),
						AcceptedFromAddresses:   MakeCopyOfAccAddresses(ac.accepted),
						Coins:                   tc.expected,
						Declined:                declined,
					}
					qr := QuarantineRecord{
						UnacceptedFromAddresses: ac.unaccepted,
						AcceptedFromAddresses:   ac.accepted,
						Coins:                   coinMakerMap[tc.qrCoinKey](),
						Declined:                declined,
					}
					addCoinsOrig := coinMakerMap[tc.addCoinKey]()
					addCoins := coinMakerMap[tc.addCoinKey]()
					qr.AddCoins(addCoins...)
					assert.Equal(t, expected, qr, "QuarantineRecord after AddCoins")
					assert.Equal(t, addCoinsOrig, addCoins, "Coins before and after")
				})
			}
		}
	}
}

func TestQuarantineRecord_IsFullyAccepted(t *testing.T) {
	testAddr0 := MakeTestAddr("qrifa", 0)
	testAddr1 := MakeTestAddr("qrifa", 1)

	tests := []struct {
		name     string
		qr       *QuarantineRecord
		expected bool
	}{
		{
			name: "no addresses at all",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: true,
		},
		{
			name: "one unaccepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: false,
		},
		{
			name: "one accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: true,
		},
		{
			name: "declined one accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			expected: true,
		},
		{
			name: "one accepted one unaccepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: false,
		},
		{
			name: "two unaccepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: false,
		},
		{
			name: "two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			expected: true,
		},
		{
			name: "declined two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := MakeCopyOfQuarantineRecord(tc.qr)
			actual := tc.qr.IsFullyAccepted()
			assert.Equal(t, tc.expected, actual, "IsFullyAccepted: %v", tc.qr)
			assert.Equal(t, orig, tc.qr, "QuarantineRecord before and after")
		})
	}
}

func TestQuarantineRecord_AcceptFrom(t *testing.T) {
	testAddr0 := MakeTestAddr("qraf", 0)
	testAddr1 := MakeTestAddr("qraf", 1)
	testAddr2 := MakeTestAddr("qraf", 2)
	testAddr3 := MakeTestAddr("qraf", 3)
	testAddr4 := MakeTestAddr("qraf", 4)

	tests := []struct {
		name  string
		qr    *QuarantineRecord
		addrs []sdk.AccAddress
		exp   bool
		expQr *QuarantineRecord
	}{
		{
			name: "control",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "nil addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: nil,
			exp:   false,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "empty addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{},
			exp:   false,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "one addrs only in accepted already",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr1},
			exp:   false,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "record has nil addresses",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   false,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "one address in both",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted two other provided",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr2, testAddr3},
			exp:   false,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted both provided",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0, testAddr1},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted both provided opposite order",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr1, testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted first provided first with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0, testAddr3, testAddr4},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted first provided second with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr3, testAddr0, testAddr4},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted first provided third with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr4, testAddr3, testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two same unaccepted provided once",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr0, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted second provided first with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr1, testAddr3, testAddr4},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted second provided second with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr3, testAddr1, testAddr4},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two unaccepted second provided third with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr4, testAddr3, testAddr1},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "one unaccepted provided thrice",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr4},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0, testAddr0, testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr4, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origInput := MakeCopyOfAccAddresses(tc.addrs)
			actual := tc.qr.AcceptFrom(tc.addrs)
			assert.Equal(t, tc.exp, actual, "AcceptFrom return value")
			assert.Equal(t, tc.expQr, tc.qr, "QuarantineRecord after AcceptFrom")
			assert.Equal(t, origInput, tc.addrs, "input address slice before and after AcceptFrom")
		})
	}
}

func TestQuarantineRecord_DeclineFrom(t *testing.T) {
	testAddr0 := MakeTestAddr("qrdf", 0)
	testAddr1 := MakeTestAddr("qrdf", 1)
	testAddr2 := MakeTestAddr("qrdf", 2)
	testAddr3 := MakeTestAddr("qrdf", 3)
	testAddr4 := MakeTestAddr("qrdf", 4)

	tests := []struct {
		name  string
		qr    *QuarantineRecord
		addrs []sdk.AccAddress
		exp   bool
		expQr *QuarantineRecord
	}{
		{
			name: "control",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "nil addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: nil,
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "nil addrs already declined",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			addrs: nil,
			exp:   false,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "empty addrs",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "one addrs only in unaccepted already",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "one addrs only in unaccepted already and already declined",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   false,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "record has nil addresses",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "one address in both",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "two accepted two other provided",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr2, testAddr3},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "two accepted both provided",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0, testAddr1},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2, testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "two accepted both provided previously declined",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			addrs: []sdk.AccAddress{testAddr0, testAddr1},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2, testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "two accepted both provided opposite order",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr1, testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2, testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "two accepted first provided first with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0, testAddr3, testAddr4},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2, testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "two accepted first provided second with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr3, testAddr0, testAddr4},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2, testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "two accepted first provided third with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			addrs: []sdk.AccAddress{testAddr4, testAddr3, testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2, testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "two same accepted provided once",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2, testAddr0, testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "two accepted second provided first with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr1, testAddr3, testAddr4},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "two accepted second provided second with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr3, testAddr1, testAddr4},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "two accepted second provided third with 2 others",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr4, testAddr3, testAddr1},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "one accepted provided thrice",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr4},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			addrs: []sdk.AccAddress{testAddr0, testAddr0, testAddr0},
			exp:   true,
			expQr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr4, testAddr0},
				AcceptedFromAddresses:   nil,
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origInput := MakeCopyOfAccAddresses(tc.addrs)
			actual := tc.qr.DeclineFrom(tc.addrs)
			assert.Equal(t, tc.exp, actual, "DeclineFrom return value")
			assert.Equal(t, tc.expQr, tc.qr, "QuarantineRecord after DeclineFrom")
			assert.Equal(t, origInput, tc.addrs, "input address slice before and after DeclineFrom")
		})
	}

}

func TestQuarantineRecord_GetAllFromAddrs(t *testing.T) {
	testAddr0 := MakeTestAddr("qrgafa", 0)
	testAddr1 := MakeTestAddr("qrgafa", 1)
	testAddr2 := MakeTestAddr("qrgafa", 2)
	testAddr3 := MakeTestAddr("qrgafa", 3)
	testAddr4 := MakeTestAddr("qrgafa", 4)

	tests := []struct {
		name string
		qr   *QuarantineRecord
		exp  []sdk.AccAddress
	}{
		{
			name: "nil unaccepted nil accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   nil,
			},
			exp: []sdk.AccAddress{},
		},
		{
			name: "nil unaccepted empty accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{},
			},
			exp: []sdk.AccAddress{},
		},
		{
			name: "empty unaccepted nil accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   nil,
			},
			exp: []sdk.AccAddress{},
		},
		{
			name: "empty unaccepted empty accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   []sdk.AccAddress{},
			},
			exp: []sdk.AccAddress{},
		},
		{
			name: "one unaccepted nil accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   nil,
			},
			exp: []sdk.AccAddress{testAddr0},
		},
		{
			name: "two unaccepted nil accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   nil,
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1},
		},
		{
			name: "one unaccepted empty accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{},
			},
			exp: []sdk.AccAddress{testAddr0},
		},
		{
			name: "two unaccepted empty accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1},
		},
		{
			name: "nil unaccepted one accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
			},
			exp: []sdk.AccAddress{testAddr0},
		},
		{
			name: "nil unaccepted two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1},
		},
		{
			name: "empty unaccepted one accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0},
			},
			exp: []sdk.AccAddress{testAddr0},
		},
		{
			name: "empty unaccepted two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr1},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1},
		},
		{
			name: "one unaccepted one accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1},
		},
		{
			name: "two unaccepted one accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1, testAddr2},
		},
		{
			name: "one unaccepted two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1, testAddr2},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr1, testAddr2},
		},
		{
			name: "two unaccepted two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr4, testAddr3},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1, testAddr2},
			},
			exp: []sdk.AccAddress{testAddr4, testAddr3, testAddr1, testAddr2},
		},
		{
			name: "three unaccepted two accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr2, testAddr3, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr0, testAddr4},
			},
			exp: []sdk.AccAddress{testAddr2, testAddr3, testAddr1, testAddr0, testAddr4},
		},
		{
			name: "two unaccepted three accepted",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr0, testAddr4},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr2, testAddr3, testAddr1},
			},
			exp: []sdk.AccAddress{testAddr0, testAddr4, testAddr2, testAddr3, testAddr1},
		},
		{
			name: "same address in both twice",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1, testAddr1},
				AcceptedFromAddresses:   []sdk.AccAddress{testAddr1, testAddr1},
			},
			exp: []sdk.AccAddress{testAddr1, testAddr1, testAddr1, testAddr1},
		},
	}

	// These shouldn't affect tests at all, but it's better to have
	// them set just in case, for some reason, they do.
	// But I didn't want to worry about them when defining the tests,
	// so I'm doing it here instead.
	for i, tc := range tests {
		if i%2 == 0 {
			tc.qr.Coins = coinMakerOK()
			tc.qr.Declined = true
		} else {
			tc.qr.Coins = coinMakerMulti()
			tc.qr.Declined = false
		}
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := MakeCopyOfQuarantineRecord(tc.qr)
			actual := tc.qr.GetAllFromAddrs()
			assert.Equal(t, tc.exp, actual, "GetAllFromAddrs result")
			assert.Equal(t, orig, tc.qr, "QuarantineRecord before and after")
		})
	}
}

func TestQuarantineRecord_AsQuarantinedFunds(t *testing.T) {
	testAddr0 := MakeTestAddr("qrasqf", 0)
	testAddr1 := MakeTestAddr("qrasqf", 1)
	testAddr2 := MakeTestAddr("qrasqf", 2)

	tests := []struct {
		name     string
		qr       *QuarantineRecord
		toAddr   sdk.AccAddress
		expected *QuarantinedFunds
	}{
		{
			name: "control",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "declined",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   coinMakerOK(),
				Declined:                true,
			},
		},
		{
			name: "bad coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerBad(),
				Declined:                false,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   coinMakerBad(),
				Declined:                false,
			},
		},
		{
			name: "empty coins",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerEmpty(),
				Declined:                false,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   coinMakerEmpty(),
				Declined:                false,
			},
		},
		{
			name: "no to address",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			toAddr: nil,
			expected: &QuarantinedFunds{
				ToAddress:               "",
				UnacceptedFromAddresses: []string{testAddr1.String()},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "nil from addresses",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: nil,
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "empty from addresses",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
		{
			name: "two from addresses",
			qr: &QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{testAddr1, testAddr2},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
			toAddr: testAddr0,
			expected: &QuarantinedFunds{
				ToAddress:               testAddr0.String(),
				UnacceptedFromAddresses: []string{testAddr1.String(), testAddr2.String()},
				Coins:                   coinMakerOK(),
				Declined:                false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qrOrig := MakeCopyOfQuarantineRecord(tc.qr)
			actual := tc.qr.AsQuarantinedFunds(tc.toAddr)
			assert.Equal(t, tc.expected, actual, "resulting QuarantinedFunds")
			assert.Equal(t, qrOrig, tc.qr, "QuarantineRecord before and after")
		})
	}
}

func TestQuarantineRecordSuffixIndex_AddSuffixes(t *testing.T) {
	suffixShort0 := []byte(MakeTestAddr("qrsias", 0))
	suffixShort1 := []byte(MakeTestAddr("qrsias", 1))
	suffixShort2 := []byte(MakeTestAddr("qrsias", 2))
	suffixShort3 := []byte(MakeTestAddr("qrsias", 3))
	suffixLong4 := []byte(MakeLongAddr("qrsias", 4))
	suffixLong5 := []byte(MakeLongAddr("qrsias", 5))
	suffixLong6 := []byte(MakeLongAddr("qrsias", 6))
	suffixLong7 := []byte(MakeLongAddr("qrsias", 7))
	suffixBad8 := []byte(MakeBadAddr("qrsias", 8))
	suffixEmpty := make([]byte, 0)

	tests := []struct {
		name  string
		qrsi  *QuarantineRecordSuffixIndex
		toAdd [][]byte
		exp   *QuarantineRecordSuffixIndex
	}{
		// nil + ...
		{
			name:  "nil + nil",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toAdd: nil,
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:  "nil + empty",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toAdd: [][]byte{},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:  "nil + short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toAdd: [][]byte{suffixShort1},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort1}},
		},
		{
			name:  "nil + long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toAdd: [][]byte{suffixLong5},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong5}},
		},
		{
			name:  "nil + short short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toAdd: [][]byte{suffixShort2, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2, suffixShort3}},
		},
		{
			name:  "nil + short long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toAdd: [][]byte{suffixShort2, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2, suffixLong6}},
		},
		{
			name:  "nil + long short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toAdd: [][]byte{suffixLong7, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong7, suffixShort3}},
		},
		{
			name:  "nil + long long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toAdd: [][]byte{suffixLong7, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong7, suffixLong6}},
		},

		// empty + ...
		{
			name:  "empty + nil",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toAdd: nil,
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
		},
		{
			name:  "empty + empty",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toAdd: [][]byte{},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
		},
		{
			name:  "empty + short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toAdd: [][]byte{suffixShort1},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort1}},
		},
		{
			name:  "empty + long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toAdd: [][]byte{suffixLong5},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong5}},
		},
		{
			name:  "empty + short short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toAdd: [][]byte{suffixShort2, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2, suffixShort3}},
		},
		{
			name:  "empty + short long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toAdd: [][]byte{suffixShort2, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2, suffixLong6}},
		},
		{
			name:  "empty + long short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toAdd: [][]byte{suffixLong7, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong7, suffixShort3}},
		},
		{
			name:  "empty + long long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toAdd: [][]byte{suffixLong7, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong7, suffixLong6}},
		},

		// short + ...
		{
			name:  "short + nil",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toAdd: nil,
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},
		{
			name:  "short + empty",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toAdd: [][]byte{},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},
		{
			name:  "short + short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toAdd: [][]byte{suffixShort1},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
		},
		{
			name:  "short + long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toAdd: [][]byte{suffixLong5},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong5}},
		},
		{
			name:  "short + short short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toAdd: [][]byte{suffixShort2, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2, suffixShort3}},
		},
		{
			name:  "short + short long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toAdd: [][]byte{suffixShort2, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2, suffixLong6}},
		},
		{
			name:  "short + long short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toAdd: [][]byte{suffixLong7, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong7, suffixShort3}},
		},
		{
			name:  "short + long long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toAdd: [][]byte{suffixLong7, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong7, suffixLong6}},
		},

		// long + ...
		{
			name:  "long + nil",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toAdd: nil,
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
		},
		{
			name:  "long + empty",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toAdd: [][]byte{},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
		},
		{
			name:  "long + short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toAdd: [][]byte{suffixShort1},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort1}},
		},
		{
			name:  "long + long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toAdd: [][]byte{suffixLong5},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5}},
		},
		{
			name:  "long + short short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toAdd: [][]byte{suffixShort2, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort2, suffixShort3}},
		},
		{
			name:  "long + short long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toAdd: [][]byte{suffixShort2, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort2, suffixLong6}},
		},
		{
			name:  "long + long short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toAdd: [][]byte{suffixLong7, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong7, suffixShort3}},
		},
		{
			name:  "long + long long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toAdd: [][]byte{suffixLong7, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong7, suffixLong6}},
		},

		// short short + ...
		{
			name:  "short short + nil",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
			toAdd: nil,
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
		},
		{
			name:  "short short + empty",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
			toAdd: [][]byte{},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
		},
		{
			name:  "short short + short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2}},
			toAdd: [][]byte{suffixShort1},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2, suffixShort1}},
		},
		{
			name:  "short short + long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
			toAdd: [][]byte{suffixLong5},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1, suffixLong5}},
		},
		{
			name:  "short short + short short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2}},
			toAdd: [][]byte{suffixShort1, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2, suffixShort1, suffixShort3}},
		},
		{
			name:  "short short + short long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2}},
			toAdd: [][]byte{suffixShort1, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2, suffixShort1, suffixLong6}},
		},
		{
			name:  "short short + long short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
			toAdd: [][]byte{suffixLong7, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1, suffixLong7, suffixShort3}},
		},
		{
			name:  "short short + long long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
			toAdd: [][]byte{suffixLong7, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1, suffixLong7, suffixLong6}},
		},

		// short long + ...
		{
			name:  "short long + nil",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4}},
			toAdd: nil,
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4}},
		},
		{
			name:  "short long + empty",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4}},
			toAdd: [][]byte{},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4}},
		},
		{
			name:  "short long + short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4}},
			toAdd: [][]byte{suffixShort1},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4, suffixShort1}},
		},
		{
			name:  "short long + long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4}},
			toAdd: [][]byte{suffixLong5},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4, suffixLong5}},
		},
		{
			name:  "short long + short short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4}},
			toAdd: [][]byte{suffixShort2, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4, suffixShort2, suffixShort3}},
		},
		{
			name:  "short long + short long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4}},
			toAdd: [][]byte{suffixShort2, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4, suffixShort2, suffixLong6}},
		},
		{
			name:  "short long + long short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4}},
			toAdd: [][]byte{suffixLong7, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4, suffixLong7, suffixShort3}},
		},
		{
			name:  "short long + long long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4}},
			toAdd: [][]byte{suffixLong7, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong4, suffixLong7, suffixLong6}},
		},

		// long short + ...
		{
			name:  "long short + nil",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0}},
			toAdd: nil,
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0}},
		},
		{
			name:  "long short + empty",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0}},
			toAdd: [][]byte{},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0}},
		},
		{
			name:  "long short + short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0}},
			toAdd: [][]byte{suffixShort1},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0, suffixShort1}},
		},
		{
			name:  "long short + long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0}},
			toAdd: [][]byte{suffixLong5},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0, suffixLong5}},
		},
		{
			name:  "long short + short short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0}},
			toAdd: [][]byte{suffixShort2, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0, suffixShort2, suffixShort3}},
		},
		{
			name:  "long short + short long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0}},
			toAdd: [][]byte{suffixShort2, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0, suffixShort2, suffixLong6}},
		},
		{
			name:  "long short + long short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0}},
			toAdd: [][]byte{suffixLong7, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0, suffixLong7, suffixShort3}},
		},
		{
			name:  "long short + long long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0}},
			toAdd: [][]byte{suffixLong7, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixShort0, suffixLong7, suffixLong6}},
		},

		// long long + ...
		{
			name:  "long long + nil",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5}},
			toAdd: nil,
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5}},
		},
		{
			name:  "long long + empty",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5}},
			toAdd: [][]byte{},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5}},
		},
		{
			name:  "long long + short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5}},
			toAdd: [][]byte{suffixShort1},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5, suffixShort1}},
		},
		{
			name:  "long long + long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong6}},
			toAdd: [][]byte{suffixLong5},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong6, suffixLong5}},
		},
		{
			name:  "long long + short short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5}},
			toAdd: [][]byte{suffixShort2, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5, suffixShort2, suffixShort3}},
		},
		{
			name:  "long long + short long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5}},
			toAdd: [][]byte{suffixShort2, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5, suffixShort2, suffixLong6}},
		},
		{
			name:  "long long + long short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5}},
			toAdd: [][]byte{suffixLong7, suffixShort3},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5, suffixLong7, suffixShort3}},
		},
		{
			name:  "long long + long long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5}},
			toAdd: [][]byte{suffixLong7, suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4, suffixLong5, suffixLong7, suffixLong6}},
		},

		// other ...
		{
			name:  "short long + bad",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort1, suffixLong7}},
			toAdd: [][]byte{suffixBad8},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort1, suffixLong7, suffixBad8}},
		},
		{
			name:  "long short + empty suffix",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong5, suffixShort0}},
			toAdd: [][]byte{suffixEmpty},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong5, suffixShort0, suffixEmpty}},
		},
		{
			name:  "bad + short short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixBad8}},
			toAdd: [][]byte{suffixShort3, suffixShort1},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixBad8, suffixShort3, suffixShort1}},
		},
		{
			name:  "empty suffix + long long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixEmpty}},
			toAdd: [][]byte{suffixLong7, suffixLong4},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixEmpty, suffixLong7, suffixLong4}},
		},
		{
			name:  "short + same short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toAdd: [][]byte{suffixShort0},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort0}},
		},
		{
			name:  "short long + same short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort1, suffixLong4}},
			toAdd: [][]byte{suffixShort1},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort1, suffixLong4, suffixShort1}},
		},
		{
			name:  "short long + same long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort1, suffixLong4}},
			toAdd: [][]byte{suffixLong4},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort1, suffixLong4, suffixLong4}},
		},
		{
			name:  "long short + same short",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong6, suffixShort2}},
			toAdd: [][]byte{suffixShort2},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong6, suffixShort2, suffixShort2}},
		},
		{
			name:  "long short + same long",
			qrsi:  &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong6, suffixShort2}},
			toAdd: [][]byte{suffixLong6},
			exp:   &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong6, suffixShort2, suffixLong6}},
		},
		{
			name: "shmorgishborg",
			qrsi: &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{
				suffixShort1, suffixShort3, suffixLong6, suffixBad8, suffixLong7,
				suffixLong4, suffixShort0, suffixLong5, suffixEmpty, suffixShort2,
			}},
			toAdd: [][]byte{
				suffixShort0, suffixBad8, suffixShort1, suffixShort1, suffixLong5,
				suffixEmpty, suffixLong4, suffixLong6, suffixShort0, suffixLong7,
			},
			exp: &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{
				suffixShort1, suffixShort3, suffixLong6, suffixBad8, suffixLong7,
				suffixLong4, suffixShort0, suffixLong5, suffixEmpty, suffixShort2,
				suffixShort0, suffixBad8, suffixShort1, suffixShort1, suffixLong5,
				suffixEmpty, suffixLong4, suffixLong6, suffixShort0, suffixLong7,
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.qrsi.AddSuffixes(tc.toAdd...)
			assert.Equal(t, tc.exp, tc.qrsi, "QuarantineRecordSuffixIndex after AddSuffixes")
		})
	}
}

func TestQuarantineRecordSuffixIndex_Simplify(t *testing.T) {
	suffixShort0 := []byte(MakeTestAddr("qrsis", 0))
	suffixShort1 := []byte(MakeTestAddr("qrsis", 1))
	suffixShort2 := []byte(MakeTestAddr("qrsis", 2))
	suffixShort3 := []byte(MakeTestAddr("qrsis", 3))
	suffixLong4 := []byte(MakeLongAddr("qrsis", 4))
	suffixLong5 := []byte(MakeLongAddr("qrsis", 5))
	suffixLong6 := []byte(MakeLongAddr("qrsis", 6))
	suffixLong7 := []byte(MakeLongAddr("qrsis", 7))
	suffixBad8 := []byte(MakeBadAddr("qrsis", 8))
	suffixEmpty := make([]byte, 0)

	tests := []struct {
		name     string
		qrsi     *QuarantineRecordSuffixIndex
		toRemove [][]byte
		exp      *QuarantineRecordSuffixIndex
	}{
		// nil - ...
		{
			name:     "nil - nil",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toRemove: nil,
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "nil - empty",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toRemove: [][]byte{},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "nil - short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toRemove: [][]byte{suffixShort1},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "nil - long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toRemove: [][]byte{suffixLong5},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "nil - short short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toRemove: [][]byte{suffixShort2, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "nil - short long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toRemove: [][]byte{suffixShort2, suffixLong6},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "nil - long short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toRemove: [][]byte{suffixLong7, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "nil - long long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
			toRemove: [][]byte{suffixLong7, suffixLong6},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},

		// empty - ...
		{
			name:     "empty - nil",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toRemove: nil,
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "empty - empty",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toRemove: [][]byte{},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "empty - short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toRemove: [][]byte{suffixShort1},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "empty - long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toRemove: [][]byte{suffixLong5},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "empty - short short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toRemove: [][]byte{suffixShort2, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "empty - short long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toRemove: [][]byte{suffixShort2, suffixLong6},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "empty - long short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toRemove: [][]byte{suffixLong7, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "empty - long long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{}},
			toRemove: [][]byte{suffixLong7, suffixLong6},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},

		// short - ...
		{
			name:     "short - nil",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toRemove: nil,
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},
		{
			name:     "short - empty",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toRemove: [][]byte{},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},
		{
			name:     "short - other short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toRemove: [][]byte{suffixShort1},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},
		{
			name:     "short - same short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toRemove: [][]byte{suffixShort0},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "short - long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toRemove: [][]byte{suffixLong5},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},
		{
			name:     "short - other short other short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toRemove: [][]byte{suffixShort2, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},
		{
			name:     "short - same short other short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toRemove: [][]byte{suffixShort0, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "short - other short same short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toRemove: [][]byte{suffixShort2, suffixShort0},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "short - same short same short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toRemove: [][]byte{suffixShort0, suffixShort0},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "short - short long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toRemove: [][]byte{suffixShort2, suffixLong6},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},
		{
			name:     "short - long short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toRemove: [][]byte{suffixLong7, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},
		{
			name:     "short - long long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
			toRemove: [][]byte{suffixLong7, suffixLong6},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},

		// long - ...
		{
			name:     "long - nil",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: nil,
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
		},
		{
			name:     "long - empty",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
		},
		{
			name:     "long - short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{suffixShort1},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
		},
		{
			name:     "long - other long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{suffixLong5},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
		},
		{
			name:     "long - same long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{suffixLong4},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "long - short short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{suffixShort2, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
		},
		{
			name:     "long - short other long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{suffixShort2, suffixLong6},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
		},
		{
			name:     "long - short same long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{suffixShort2, suffixLong4},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "long - other long short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{suffixLong7, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
		},
		{
			name:     "long - same long short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{suffixLong4, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "long - other long other long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{suffixLong7, suffixLong6},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
		},
		{
			name:     "long - other long same long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{suffixLong7, suffixLong4},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "long - same long other long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{suffixLong4, suffixLong6},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "long - same long same long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong4}},
			toRemove: [][]byte{suffixLong4, suffixLong4},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},

		// short short - ...
		{
			name:     "short short - nil",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
			toRemove: nil,
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
		},
		{
			name:     "short short - empty",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
			toRemove: [][]byte{},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
		},
		{
			name:     "short short - other short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2}},
			toRemove: [][]byte{suffixShort1},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2}},
		},
		{
			name:     "short short - same first short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2}},
			toRemove: [][]byte{suffixShort0},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2}},
		},
		{
			name:     "short short - same second short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2}},
			toRemove: [][]byte{suffixShort2},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},
		{
			name:     "short short - long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
			toRemove: [][]byte{suffixLong5},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
		},
		{
			name:     "short short - other short other short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2, suffixShort0}},
			toRemove: [][]byte{suffixShort1, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2}},
		},
		{
			name:     "short short - first short other short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2, suffixShort0}},
			toRemove: [][]byte{suffixShort2, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},
		{
			name:     "short short - second short other short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2, suffixShort0}},
			toRemove: [][]byte{suffixShort0, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2}},
		},
		{
			name:     "short short - other short first short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2, suffixShort0}},
			toRemove: [][]byte{suffixShort1, suffixShort2},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0}},
		},
		{
			name:     "short short - other short second short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2, suffixShort0}},
			toRemove: [][]byte{suffixShort1, suffixShort0},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2}},
		},
		{
			name:     "short short - first short second short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2, suffixShort0}},
			toRemove: [][]byte{suffixShort2, suffixShort0},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "short short - second short first short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort2, suffixShort0}},
			toRemove: [][]byte{suffixShort0, suffixShort2},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "short short - short long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2}},
			toRemove: [][]byte{suffixShort1, suffixLong6},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort2}},
		},
		{
			name:     "short short - long short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
			toRemove: [][]byte{suffixLong7, suffixShort3},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
		},
		{
			name:     "short short - long long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
			toRemove: [][]byte{suffixLong7, suffixLong6},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixShort1}},
		},

		// other ...
		{
			name:     "short long - bad",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort1, suffixLong7}},
			toRemove: [][]byte{suffixBad8},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort1, suffixLong7}},
		},
		{
			name:     "long short - empty suffix",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixLong5, suffixShort0}},
			toRemove: [][]byte{suffixEmpty},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixShort0, suffixLong5}},
		},
		{
			name:     "bad - short short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixBad8}},
			toRemove: [][]byte{suffixShort3, suffixShort1},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixBad8}},
		},
		{
			name:     "bad - short bad long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixBad8}},
			toRemove: [][]byte{suffixShort3, suffixBad8, suffixLong7},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name:     "empty suffix - long long",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixEmpty}},
			toRemove: [][]byte{suffixLong7, suffixLong4},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixEmpty}},
		},
		{
			name:     "empty suffix - long empty suffix short",
			qrsi:     &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{suffixEmpty}},
			toRemove: [][]byte{suffixLong7, suffixEmpty, suffixShort2},
			exp:      &QuarantineRecordSuffixIndex{RecordSuffixes: nil},
		},
		{
			name: "shmorgishborg",
			qrsi: &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{
				suffixShort1, suffixShort3, suffixLong6, suffixBad8, suffixLong7,
				suffixLong4, suffixShort0, suffixLong5, suffixEmpty, suffixShort2,
			}},
			toRemove: [][]byte{
				suffixShort0, suffixBad8, suffixShort1, suffixShort1, suffixLong4,
				suffixEmpty, suffixLong4, suffixLong6, suffixShort0, suffixLong7,
			},
			exp: &QuarantineRecordSuffixIndex{RecordSuffixes: [][]byte{
				suffixShort2, suffixShort3, suffixLong5,
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.toRemove == nil {
				tc.qrsi.Simplify()
			} else {
				tc.qrsi.Simplify(tc.toRemove...)
			}
			assert.Equal(t, tc.exp, tc.qrsi, "QuarantineRecordSuffixIndex after Simplify")
		})
	}
}
