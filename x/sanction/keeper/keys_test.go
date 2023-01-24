package keeper_test

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

func TestPrefixValues(t *testing.T) {
	prefixes := []struct {
		name     string
		prefix   []byte
		expected []byte
	}{
		{name: "ParamsPrefix", prefix: keeper.ParamsPrefix, expected: []byte{0x00}},
		{name: "SanctionedPrefix", prefix: keeper.SanctionedPrefix, expected: []byte{0x01}},
		{name: "TemporaryPrefix", prefix: keeper.TemporaryPrefix, expected: []byte{0x02}},
		{name: "ProposalIndexPrefix", prefix: keeper.ProposalIndexPrefix, expected: []byte{0x03}},
	}

	for i, p := range prefixes {
		t.Run(fmt.Sprintf("%s expected value", p.name), func(t *testing.T) {
			assert.Equal(t, p.prefix, p.expected, "prefix value")

			for j, p2 := range prefixes {
				if i == j {
					continue
				}
				assert.NotEqual(t, p.prefix, p2.prefix, "%v = %s = %s", p.prefix, p.name, p2.name)
			}
		})
	}
}

func TestConstValues(t *testing.T) {
	consts := []struct {
		name      string
		value     string
		exptected string
	}{
		{
			name:      "ParamNameImmediateSanctionMinDeposit",
			value:     keeper.ParamNameImmediateSanctionMinDeposit,
			exptected: "immediate_sanction_min_deposit",
		},
		{
			name:      "ParamNameImmediateUnsanctionMinDeposit",
			value:     keeper.ParamNameImmediateUnsanctionMinDeposit,
			exptected: "immediate_unsanction_min_deposit",
		},
	}

	for i, c := range consts {
		t.Run(fmt.Sprintf("%s", c.name), func(t *testing.T) {
			assert.Equal(t, c.exptected, c.value, "variable value")

			for j, c2 := range consts {
				if i == j {
					continue
				}
				assert.NotEqual(t, c.value, c2.value, "%q = %s = %s", c.value, c.name, c2.name)
			}
		})
	}

}

func TestConcatBz(t *testing.T) {
	type testCase struct {
		name     string
		bz1      []byte
		bz2      []byte
		expected []byte
	}
	copyTestCase := func(tc testCase) testCase {
		rv := testCase{
			name:     tc.name,
			bz1:      nil,
			bz2:      nil,
			expected: nil,
		}
		if tc.bz1 != nil {
			rv.bz1 = make([]byte, len(tc.bz1), cap(tc.bz1))
			copy(rv.bz1, tc.bz1)
		}
		if tc.bz2 != nil {
			rv.bz2 = make([]byte, len(tc.bz2), cap(tc.bz2))
			copy(rv.bz2, tc.bz2)
		}
		if tc.expected != nil {
			rv.expected = make([]byte, len(tc.expected), cap(tc.expected))
			copy(rv.expected, tc.expected)
		}
		return rv
	}

	tests := []testCase{
		{
			name:     "nil nil",
			bz1:      nil,
			bz2:      nil,
			expected: []byte{},
		},
		{
			name:     "nil empty",
			bz1:      nil,
			bz2:      []byte{},
			expected: []byte{},
		},
		{
			name:     "empty nil",
			bz1:      []byte{},
			bz2:      nil,
			expected: []byte{},
		},
		{
			name:     "empty empty",
			bz1:      []byte{},
			bz2:      []byte{},
			expected: []byte{},
		},
		{
			name:     "nil 1 byte",
			bz1:      nil,
			bz2:      []byte{'a'},
			expected: []byte{'a'},
		},
		{
			name:     "empty 1 byte",
			bz1:      []byte{},
			bz2:      []byte{'a'},
			expected: []byte{'a'},
		},
		{
			name:     "nil 4 bytes",
			bz1:      nil,
			bz2:      []byte("test"),
			expected: []byte("test"),
		},
		{
			name:     "empty 4 bytes",
			bz1:      []byte{},
			bz2:      []byte("test"),
			expected: []byte("test"),
		},
		{
			name:     "1 byte nil",
			bz1:      []byte{'a'},
			bz2:      nil,
			expected: []byte{'a'},
		},
		{
			name:     "1 byte empty",
			bz1:      []byte{'a'},
			bz2:      []byte{},
			expected: []byte{'a'},
		},
		{
			name:     "4 bytes nil",
			bz1:      []byte("test"),
			bz2:      nil,
			expected: []byte("test"),
		},
		{
			name:     "4 bytes empty",
			bz1:      []byte("test"),
			bz2:      []byte{},
			expected: []byte("test"),
		},
		{
			name:     "1 byte 1 byte",
			bz1:      []byte{'a'},
			bz2:      []byte{'b'},
			expected: []byte{'a', 'b'},
		},
		{
			name:     "1 byte 4 bytes",
			bz1:      []byte{'a'},
			bz2:      []byte("test"),
			expected: []byte("atest"),
		},
		{
			name:     "4 bytes 1 byte",
			bz1:      []byte("word"),
			bz2:      []byte{'x'},
			expected: []byte("wordx"),
		},
		{
			name:     "5 bytes 5 bytes",
			bz1:      []byte("hello"),
			bz2:      []byte("world"),
			expected: []byte("helloworld"),
		},
	}

	for _, tc_orig := range tests {
		passes := t.Run(tc_orig.name, func(t *testing.T) {
			tc := copyTestCase(tc_orig)
			var actual []byte
			testFunc := func() {
				actual = keeper.ConcatBz(tc.bz1, tc.bz2)
			}
			require.NotPanics(t, testFunc, "ConcatBz")
			assert.Equal(t, tc.expected, actual, "ConcatBz result")
			assert.Equal(t, len(tc.expected), len(actual), "ConcatBz result length")
			assert.Equal(t, cap(tc.expected), cap(actual), "ConcatBz result capacity")
			assert.Equal(t, len(actual), cap(actual), "ConcatBz result length and capacity")
			assert.Equal(t, tc_orig.bz1, tc.bz1, "input 1 before and after ConcatBz")
			assert.Equal(t, len(tc_orig.bz1), len(tc.bz1), "input 1 length before and after ConcatBz")
			assert.Equal(t, cap(tc_orig.bz1), cap(tc.bz1), "input 1 capacity before and after ConcatBz")
			assert.Equal(t, tc_orig.bz2, tc.bz2, "input 2 before and after ConcatBz")
			assert.Equal(t, len(tc_orig.bz2), len(tc.bz2), "input 2 length before and after ConcatBz")
			assert.Equal(t, cap(tc_orig.bz2), cap(tc.bz2), "input 2 capacity before and after ConcatBz")
			if cap(tc.bz1) > 0 {
				if len(tc.bz1) > 0 {
					if tc.bz1[0] == 'x' {
						tc.bz1[0] = 'y'
					} else {
						tc.bz1[0] = 'x'
					}
					assert.Equal(t, tc.expected, actual, "ConcatBz result after changing original bz1 input")
				}
				if len(tc.bz1) < cap(tc.bz1) {
					tc.bz1 = tc.bz1[:len(tc.bz1)+1]
					tc.bz1[len(tc.bz1)] = 'x'
					assert.Equal(t, tc.expected, actual, "ConcatBz result after extending original bz1 input")
				}
			}
			if cap(tc.bz2) > 0 {
				if len(tc.bz2) > 0 {
					if tc.bz2[0] == 'x' {
						tc.bz2[0] = 'y'
					} else {
						tc.bz2[0] = 'x'
					}
					assert.Equal(t, tc.expected, actual, "ConcatBz result after changing original bz2 input")
				}
				if len(tc.bz2) < cap(tc.bz2) {
					tc.bz2 = tc.bz2[:len(tc.bz2)+1]
					tc.bz2[len(tc.bz2)] = 'x'
					assert.Equal(t, tc.expected, actual, "ConcatBz result after extending original bz2 input")
				}
			}
		})
		if !passes {
			continue
		}

		if len(tc_orig.expected) > 0 {
			t.Run(tc_orig.name+" changing result", func(t *testing.T) {
				tc := copyTestCase(tc_orig)
				actual := keeper.ConcatBz(tc.bz1, tc.bz2)
				if len(actual) > 0 {
					if actual[0] == 'x' {
						actual[0] = 'y'
					} else {
						actual[0] = 'x'
					}
					assert.Equal(t, tc_orig.bz1, tc.bz1, "original bz1 after changing first result byte")
					assert.Equal(t, len(tc_orig.bz1), len(tc.bz1), "original bz1 length after changing first result byte")
					assert.Equal(t, cap(tc_orig.bz1), cap(tc.bz1), "original bz1 capacity after changing first result byte")
				}
				if len(actual) > 1 {
					if actual[len(actual)-1] == 'x' {
						actual[len(actual)-1] = 'y'
					} else {
						actual[len(actual)-1] = 'x'
					}
					assert.Equal(t, tc_orig.bz2, tc.bz2, "original bz2 after changing last result byte")
					assert.Equal(t, len(tc_orig.bz2), len(tc.bz2), "original bz2 length after changing last result byte")
					assert.Equal(t, cap(tc_orig.bz2), cap(tc.bz2), "original bz2 capacity after changing last result byte")
				}
			})
		}

		t.Run(tc_orig.name+" plus cap", func(t *testing.T) {
			tc := copyTestCase(tc_orig)
			plusCap := 5
			actual := keeper.OnlyTestsConcatBzPlusCap(tc.bz1, tc.bz2, plusCap)
			assert.Equal(t, tc.expected, actual, "concatBzPlusCap result")
			assert.Equal(t, len(tc.expected), len(actual), "concatBzPlusCap result length")
			assert.Equal(t, cap(tc.expected)+plusCap, cap(actual), "concatBzPlusCap result capacity")
			actual = actual[:len(actual)+1]
			actual[len(actual)-1] = 'x'
			assert.Equal(t, tc_orig.bz1, tc.bz1, "input 1 after extending result from concatBzPlusCap")
			assert.Equal(t, tc_orig.bz2, tc.bz2, "input 2 after extending result from concatBzPlusCap")
		})
	}
}

func TestParseLengthPrefixedBz(t *testing.T) {
	tests := []struct {
		name      string
		bz        []byte
		expAddr   []byte
		expSuffix []byte
		expPanic  string
	}{
		{
			name:     "nil",
			bz:       nil,
			expPanic: "expected key of length at least 1, got 0",
		},
		{
			name:     "empty",
			bz:       []byte{},
			expPanic: "expected key of length at least 1, got 0",
		},
		{
			name:      "only length byte of 0",
			bz:        []byte{0},
			expAddr:   []byte{},
			expSuffix: nil,
		},
		{
			name:     "only length byte of 1",
			bz:       []byte{1},
			expPanic: "expected key of length at least 2, got 1",
		},
		{
			name: "length byte 20 but one short",
			bz: []byte{20,
				'1', '2', '3', '4', '5', '6', '7', '8', '9', '0',
				'1', '2', '3', '4', '5', '6', '7', '8', '9',
			},
			expPanic: "expected key of length at least 21, got 20",
		},
		{
			name:      "length byte of 0 with extra",
			bz:        []byte{0, 'a', 'b', 'c'},
			expAddr:   []byte{},
			expSuffix: []byte("abc"),
		},
		{
			name:      "20 bytes no suffix",
			bz:        address.MustLengthPrefix([]byte("test_20_byte_addr___")),
			expAddr:   []byte("test_20_byte_addr___"),
			expSuffix: nil,
		},
		{
			name:      "20 bytes with suffix",
			bz:        append(address.MustLengthPrefix([]byte("test_20_byte_addr_2_")), []byte("something")...),
			expAddr:   sdk.AccAddress("test_20_byte_addr_2_"),
			expSuffix: []byte("something"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addr, suffix []byte
			testFunc := func() {
				addr, suffix = keeper.ParseLengthPrefixedBz(tc.bz)
			}
			if len(tc.expPanic) > 0 {
				require.PanicsWithValue(t, tc.expPanic, testFunc, "ParseLengthPrefixedBz")
			} else {
				require.NotPanics(t, testFunc, "ParseLengthPrefixedBz")
				assert.Equal(t, tc.expAddr, addr, "ParseLengthPrefixedBz result addr")
				assert.Equal(t, tc.expSuffix, suffix, "ParseLengthPrefixedBz result suffix")
			}
		})
	}
}

func TestCreateParamKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   []byte
	}{
		{
			name:  "control",
			input: "a word",
			exp:   append([]byte{keeper.ParamsPrefix[0]}, "a word"...),
		},
		{
			name:  "empty",
			input: "",
			exp:   keeper.ParamsPrefix,
		},
		{
			name:  "ParamNameImmediateSanctionMinDeposit",
			input: keeper.ParamNameImmediateSanctionMinDeposit,
			exp:   append([]byte{keeper.ParamsPrefix[0]}, keeper.ParamNameImmediateSanctionMinDeposit...),
		},
		{
			name:  "ParamNameImmediateUnsanctionMinDeposit",
			input: keeper.ParamNameImmediateUnsanctionMinDeposit,
			exp:   append([]byte{keeper.ParamsPrefix[0]}, keeper.ParamNameImmediateUnsanctionMinDeposit...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.CreateParamKey(tc.input)
			}
			require.NotPanics(t, testFunc, "CreateParamKey")
			assert.Equal(t, tc.exp, actual, "CreateParamKey result")
		})
	}
}

func TestParseParamKey(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		exp   string
	}{
		{
			name:  "control",
			input: append([]byte{keeper.ParamsPrefix[0]}, "a word"...),
			exp:   "a word",
		},
		{
			name:  "empty",
			input: keeper.ParamsPrefix,
			exp:   "",
		},
		{
			name:  "ParamNameImmediateSanctionMinDeposit",
			input: keeper.CreateParamKey(keeper.ParamNameImmediateSanctionMinDeposit),
			exp:   keeper.ParamNameImmediateSanctionMinDeposit,
		},
		{
			name:  "ParamNameImmediateUnsanctionMinDeposit",
			input: keeper.CreateParamKey(keeper.ParamNameImmediateUnsanctionMinDeposit),
			exp:   keeper.ParamNameImmediateUnsanctionMinDeposit,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = keeper.ParseParamKey(tc.input)
			}
			require.NotPanics(t, testFunc, "ParseParamKey")
			assert.Equal(t, tc.exp, actual, "ParseParamKey result")
		})
	}
}

func TestCreateSanctionedAddrKey(t *testing.T) {
	tests := []struct {
		name string
		addr sdk.AccAddress
		exp  []byte
	}{
		{
			name: "nil addr",
			addr: nil,
			exp:  []byte{keeper.SanctionedPrefix[0]},
		},
		{
			name: "4 byte address",
			addr: sdk.AccAddress("test"),
			exp:  append([]byte{keeper.SanctionedPrefix[0], 4}, "test"...),
		},
		{
			name: "20 byte address",
			addr: sdk.AccAddress("test_20_byte_address"),
			exp:  append([]byte{keeper.SanctionedPrefix[0], 20}, "test_20_byte_address"...),
		},
		{
			name: "32 byte address",
			addr: sdk.AccAddress("test_____32_____byte_____address"),
			exp:  append([]byte{keeper.SanctionedPrefix[0], 32}, "test_____32_____byte_____address"...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.CreateSanctionedAddrKey(tc.addr)
			}
			require.NotPanics(t, testFunc, "CreateSanctionedAddrKey")
			assert.Equal(t, tc.exp, actual, "CreateSanctionedAddrKey result")
		})
	}
}

func TestParseSanctionedAddrKey(t *testing.T) {
	tests := []struct {
		name     string
		key      []byte
		exp      sdk.AccAddress
		expPanic string
	}{
		{
			name:     "nil",
			key:      nil,
			expPanic: "runtime error: slice bounds out of range [1:0]",
		},
		{
			name:     "empty",
			key:      []byte{},
			expPanic: "runtime error: slice bounds out of range [1:0]",
		},
		{
			name:     "just one byte",
			key:      []byte{'f'}, // doesn't matter what that byte is.
			expPanic: "expected key of length at least 1, got 0",
		},
		{
			name: "empty addr",
			key:  []byte{'g', 0},
			exp:  sdk.AccAddress{},
		},
		{
			name: "4 byte addr",
			key:  []byte{'P', 4, 't', 'e', 's', 't'},
			exp:  sdk.AccAddress("test"),
		},
		{
			name: "20 byte addr",
			key:  keeper.CreateSanctionedAddrKey(sdk.AccAddress("this_test_addr_is_20")),
			exp:  sdk.AccAddress("this_test_addr_is_20"),
		},
		{
			name: "32 byte addr",
			key:  keeper.CreateSanctionedAddrKey(sdk.AccAddress("this_test_addr_is_longer_with_32")),
			exp:  sdk.AccAddress("this_test_addr_is_longer_with_32"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.AccAddress
			testFunc := func() {
				actual = keeper.ParseSanctionedAddrKey(tc.key)
			}
			if len(tc.expPanic) > 0 {
				testutil.RequirePanicsWithMessage(t, tc.expPanic, testFunc, "ParseSanctionedAddrKey")
			} else {
				require.NotPanics(t, testFunc, "ParseSanctionedAddrKey")
				assert.Equal(t, tc.exp, actual, "ParseSanctionedAddrKey result")
			}
		})
	}
}

func TestCreateTemporaryAddrPrefix(t *testing.T) {
	tests := []struct {
		name   string
		addr   sdk.AccAddress
		exp    []byte
		expCap int
	}{
		{
			name:   "nil addr",
			addr:   nil,
			exp:    []byte{keeper.TemporaryPrefix[0]},
			expCap: 1,
		},
		{
			name:   "4 byte address",
			addr:   sdk.AccAddress("test"),
			exp:    append([]byte{keeper.TemporaryPrefix[0], 4}, "test"...),
			expCap: 14,
		},
		{
			name:   "20 byte address",
			addr:   sdk.AccAddress("test_20_byte_address"),
			exp:    append([]byte{keeper.TemporaryPrefix[0], 20}, "test_20_byte_address"...),
			expCap: 30,
		},
		{
			name:   "32 byte address",
			addr:   sdk.AccAddress("test_____32_____byte_____address"),
			exp:    append([]byte{keeper.TemporaryPrefix[0], 32}, "test_____32_____byte_____address"...),
			expCap: 42,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.CreateTemporaryAddrPrefix(tc.addr)
			}
			require.NotPanics(t, testFunc, "CreateSanctionedAddrKey")
			assert.Equal(t, tc.exp, actual, "CreateSanctionedAddrKey result")
			assert.Equal(t, tc.expCap, cap(actual), "CreateSanctionedAddrKey result capacity")
		})
	}
}

func TestCreateTemporaryKey(t *testing.T) {
	tests := []struct {
		name   string
		addr   sdk.AccAddress
		govId  uint64
		exp    []byte
		expCap int
	}{
		{
			name:   "nil addr id 0",
			addr:   nil,
			govId:  0,
			exp:    []byte{keeper.TemporaryPrefix[0], 0, 0, 0, 0, 0, 0, 0, 0},
			expCap: 9,
		},
		{
			name:   "4 byte address id 1",
			addr:   sdk.AccAddress("test"),
			govId:  1,
			exp:    append([]byte{keeper.TemporaryPrefix[0], 4}, append([]byte("test"), 0, 0, 0, 0, 0, 0, 0, 1)...),
			expCap: 14,
		},
		{
			name:   "20 byte address id 2",
			addr:   sdk.AccAddress("test_20_byte_address"),
			govId:  2,
			exp:    append([]byte{keeper.TemporaryPrefix[0], 20}, append([]byte("test_20_byte_address"), 0, 0, 0, 0, 0, 0, 0, 2)...),
			expCap: 30,
		},
		{
			name:   "32 byte address id 3",
			addr:   sdk.AccAddress("test_____32_____byte_____address"),
			govId:  3,
			exp:    append([]byte{keeper.TemporaryPrefix[0], 32}, append([]byte("test_____32_____byte_____address"), 0, 0, 0, 0, 0, 0, 0, 3)...),
			expCap: 42,
		},
		{
			name:   "20 byte address id 1000",
			addr:   sdk.AccAddress("test_20_byte_address"),
			govId:  1000,
			exp:    append([]byte{keeper.TemporaryPrefix[0], 20}, append([]byte("test_20_byte_address"), 0, 0, 0, 0, 0, 0, 3, 232)...),
			expCap: 30,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.CreateTemporaryKey(tc.addr, tc.govId)
			}
			require.NotPanics(t, testFunc, "CreateTemporaryKey")
			assert.Equal(t, tc.exp, actual, "CreateTemporaryKey result")
			assert.Equal(t, tc.expCap, cap(actual), "CreateTemporaryKey result capacity")
		})
	}
}

func TestParseTemporaryKey(t *testing.T) {
	tests := []struct {
		name     string
		key      []byte
		expAddr  sdk.AccAddress
		expId    uint64
		expPanic string
	}{
		{
			name:     "nil",
			key:      nil,
			expPanic: "runtime error: slice bounds out of range [1:0]",
		},
		{
			name:     "empty",
			key:      []byte{},
			expPanic: "runtime error: slice bounds out of range [1:0]",
		},
		{
			name:     "just one byte",
			key:      []byte{'f'}, // doesn't matter what that byte is.
			expPanic: "expected key of length at least 1, got 0",
		},
		{
			name:     "empty addr only 7 id bytes",
			key:      []byte{'g', 0, 0, 0, 0, 0, 0, 0, 77},
			expPanic: "runtime error: index out of range [7] with length 7",
		},
		{
			name:    "empty addr id 1",
			key:     []byte{'g', 0, 0, 0, 0, 0, 0, 0, 0, 1},
			expAddr: sdk.AccAddress{},
			expId:   1,
		},
		{
			name:    "4 byte addr id 2",
			key:     []byte{'P', 4, 't', 'e', 's', 't', 0, 0, 0, 0, 0, 0, 0, 2},
			expAddr: sdk.AccAddress("test"),
			expId:   2,
		},
		{
			name:    "20 byte addr id 3",
			key:     keeper.CreateTemporaryKey(sdk.AccAddress("this_test_addr_is_20"), 3),
			expAddr: sdk.AccAddress("this_test_addr_is_20"),
			expId:   3,
		},
		{
			name:    "32 byte addr id 4",
			key:     keeper.CreateTemporaryKey(sdk.AccAddress("this_test_addr_is_longer_with_32"), 4),
			expAddr: sdk.AccAddress("this_test_addr_is_longer_with_32"),
			expId:   4,
		},
		{
			name:    "20 byte addr id 1000",
			key:     keeper.CreateTemporaryKey(sdk.AccAddress("this_test_addr_is_20"), 1000),
			expAddr: sdk.AccAddress("this_test_addr_is_20"),
			expId:   1000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addr sdk.AccAddress
			var id uint64
			testFunc := func() {
				addr, id = keeper.ParseTemporaryKey(tc.key)
			}
			if len(tc.expPanic) > 0 {
				testutil.RequirePanicsWithMessage(t, tc.expPanic, testFunc, "ParseTemporaryKey")
			} else {
				require.NotPanics(t, testFunc, "ParseTemporaryKey")
				assert.Equal(t, tc.expAddr, addr, "ParseTemporaryKey address")
				assert.Equal(t, tc.expId, id, "ParseTemporaryKey gov prop id")
			}
		})
	}
}

func TestTempBValues(t *testing.T) {
	// If these were the same, it'd be bad.
	assert.NotEqual(t, keeper.SanctionB, keeper.UnsanctionB, "%v = SanctionB = UnsanctionB", keeper.SanctionB)
}

func TestIsSanctionBz(t *testing.T) {
	tests := []struct {
		name string
		bz   []byte
		exp  bool
	}{
		{name: "nil", bz: nil, exp: false},
		{name: "empty", bz: []byte{}, exp: false},
		{name: "SanctionB and 0", bz: []byte{keeper.SanctionB, 0}, exp: false},
		{name: "UnsanctionB and 0", bz: []byte{keeper.UnsanctionB, 0}, exp: false},
		{name: "0 and SanctionB", bz: []byte{0, keeper.SanctionB}, exp: false},
		{name: "0 and UnsanctionB", bz: []byte{0, keeper.UnsanctionB}, exp: false},
		{name: "the letter f", bz: []byte{'f'}, exp: false},
		{name: "SanctionB", bz: []byte{keeper.SanctionB}, exp: true},
		{name: "UnsanctionB", bz: []byte{keeper.UnsanctionB}, exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = keeper.IsSanctionBz(tc.bz)
			}
			require.NotPanics(t, testFunc, "IsSanctionBz")
			assert.Equal(t, tc.exp, actual, "IsSanctionBz result")
		})
	}
}

func TestIsUnsanctionBz(t *testing.T) {
	tests := []struct {
		name string
		bz   []byte
		exp  bool
	}{
		{name: "nil", bz: nil, exp: false},
		{name: "empty", bz: []byte{}, exp: false},
		{name: "SanctionB and 0", bz: []byte{keeper.SanctionB, 0}, exp: false},
		{name: "UnsanctionB and 0", bz: []byte{keeper.UnsanctionB, 0}, exp: false},
		{name: "0 and SanctionB", bz: []byte{0, keeper.SanctionB}, exp: false},
		{name: "0 and UnsanctionB", bz: []byte{0, keeper.UnsanctionB}, exp: false},
		{name: "the letter f", bz: []byte{'f'}, exp: false},
		{name: "SanctionB", bz: []byte{keeper.SanctionB}, exp: false},
		{name: "UnsanctionB", bz: []byte{keeper.UnsanctionB}, exp: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = keeper.IsUnsanctionBz(tc.bz)
			}
			require.NotPanics(t, testFunc, "IsUnsanctionBz")
			assert.Equal(t, tc.exp, actual, "IsUnsanctionBz result")
		})
	}
}

func TestToTempStatus(t *testing.T) {
	tests := []struct {
		name string
		bz   []byte
		exp  sanction.TempStatus
	}{
		{name: "nil", bz: nil, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "empty", bz: []byte{}, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "SanctionB and 0", bz: []byte{keeper.SanctionB, 0}, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "UnsanctionB and 0", bz: []byte{keeper.UnsanctionB, 0}, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "0 and SanctionB", bz: []byte{0, keeper.SanctionB}, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "0 and UnsanctionB", bz: []byte{0, keeper.UnsanctionB}, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "the letter f", bz: []byte{'f'}, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "SanctionB", bz: []byte{keeper.SanctionB}, exp: sanction.TEMP_STATUS_SANCTIONED},
		{name: "UnsanctionB", bz: []byte{keeper.UnsanctionB}, exp: sanction.TEMP_STATUS_UNSANCTIONED},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sanction.TempStatus
			testFunc := func() {
				actual = keeper.ToTempStatus(tc.bz)
			}
			require.NotPanics(t, testFunc, "ToTempStatus")
			assert.Equal(t, tc.exp, actual, "ToTempStatus result")
		})
	}
}

func TestNewTempEvent(t *testing.T) {
	tests := []struct {
		name     string
		typeVal  byte
		addr     sdk.AccAddress
		exp      proto.Message
		expPanic string
	}{
		{
			name:    "SanctionB nil addr",
			typeVal: keeper.SanctionB,
			addr:    nil,
			exp:     &sanction.EventTempAddressSanctioned{Address: ""},
		},
		{
			name:    "SanctionB empty addr",
			typeVal: keeper.SanctionB,
			addr:    sdk.AccAddress{},
			exp:     &sanction.EventTempAddressSanctioned{Address: ""},
		},
		{
			name:    "SanctionB 20 byte addr",
			typeVal: keeper.SanctionB,
			addr:    sdk.AccAddress("this_is_a_short_addr"),
			exp:     &sanction.EventTempAddressSanctioned{Address: sdk.AccAddress("this_is_a_short_addr").String()},
		},
		{
			name:    "SanctionB 32 byte addr",
			typeVal: keeper.SanctionB,
			addr:    sdk.AccAddress("this_is_a_longer_addr_for_tests_"),
			exp:     &sanction.EventTempAddressSanctioned{Address: sdk.AccAddress("this_is_a_longer_addr_for_tests_").String()},
		},
		{
			name:    "UnsanctionB nil addr",
			typeVal: keeper.UnsanctionB,
			addr:    nil,
			exp:     &sanction.EventTempAddressUnsanctioned{Address: ""},
		},
		{
			name:    "UnsanctionB empty addr",
			typeVal: keeper.UnsanctionB,
			addr:    sdk.AccAddress{},
			exp:     &sanction.EventTempAddressUnsanctioned{Address: ""},
		},
		{
			name:    "UnsanctionB 20 byte addr",
			typeVal: keeper.UnsanctionB,
			addr:    sdk.AccAddress("this_is_a_short_addr"),
			exp:     &sanction.EventTempAddressUnsanctioned{Address: sdk.AccAddress("this_is_a_short_addr").String()},
		},
		{
			name:    "UnsanctionB 32 byte addr",
			typeVal: keeper.UnsanctionB,
			addr:    sdk.AccAddress("this_is_a_longer_addr_for_tests_"),
			exp:     &sanction.EventTempAddressUnsanctioned{Address: sdk.AccAddress("this_is_a_longer_addr_for_tests_").String()},
		},
		{
			name:     "unknown type byte",
			typeVal:  42,
			addr:     sdk.AccAddress("does_not_matter"),
			expPanic: "unknown temp value byte: 2a",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual proto.Message
			testFunc := func() {
				actual = keeper.NewTempEvent(tc.typeVal, tc.addr)
			}
			if len(tc.expPanic) > 0 {
				testutil.RequirePanicsWithMessage(t, tc.expPanic, testFunc, "NewTempEvent")
			} else {
				require.NotPanics(t, testFunc, "NewTempEvent")
				assert.Equal(t, tc.exp, actual, "NewTempEvent result")
			}
		})
	}
}

func TestCreateProposalTempIndexPrefix(t *testing.T) {
	uint64p := func(v uint64) *uint64 {
		return &v
	}
	tests := []struct {
		name   string
		id     *uint64
		exp    []byte
		expCap int
	}{
		{
			name:   "nil id",
			id:     nil,
			exp:    []byte{keeper.ProposalIndexPrefix[0]},
			expCap: 1,
		},
		{
			name:   "id 0",
			id:     uint64p(0),
			exp:    []byte{keeper.ProposalIndexPrefix[0], 0, 0, 0, 0, 0, 0, 0, 0},
			expCap: 42,
		},
		{
			name:   "id 1",
			id:     uint64p(1),
			exp:    []byte{keeper.ProposalIndexPrefix[0], 0, 0, 0, 0, 0, 0, 0, 1},
			expCap: 42,
		},
		{
			name:   "id 100",
			id:     uint64p(100),
			exp:    []byte{keeper.ProposalIndexPrefix[0], 0, 0, 0, 0, 0, 0, 0, 100},
			expCap: 42,
		},
		{
			name:   "id 1000",
			id:     uint64p(1000),
			exp:    []byte{keeper.ProposalIndexPrefix[0], 0, 0, 0, 0, 0, 0, 3, 232},
			expCap: 42,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.CreateProposalTempIndexPrefix(tc.id)
			}
			require.NotPanics(t, testFunc, "CreateProposalTempIndexPrefix")
			require.Equal(t, tc.exp, actual, "CreateProposalTempIndexPrefix result")
			require.Equal(t, tc.expCap, cap(actual), "CreateProposalTempIndexPrefix result capacity")
		})
	}
}

func TestCreateProposalTempIndexKey(t *testing.T) {
	tests := []struct {
		name  string
		addr  sdk.AccAddress
		govId uint64
		exp   []byte
	}{
		{
			name:  "nil addr id 0",
			addr:  nil,
			govId: 0,
			exp:   []byte{keeper.ProposalIndexPrefix[0], 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:  "4 byte address id 1",
			addr:  sdk.AccAddress("test"),
			govId: 1,
			exp:   append([]byte{keeper.ProposalIndexPrefix[0], 0, 0, 0, 0, 0, 0, 0, 1, 4}, "test"...),
		},
		{
			name:  "20 byte address id 2",
			addr:  sdk.AccAddress("test_20_byte_address"),
			govId: 2,
			exp:   append([]byte{keeper.ProposalIndexPrefix[0], 0, 0, 0, 0, 0, 0, 0, 2, 20}, "test_20_byte_address"...),
		},
		{
			name:  "32 byte address id 3",
			addr:  sdk.AccAddress("test_____32_____byte_____address"),
			govId: 3,
			exp:   append([]byte{keeper.ProposalIndexPrefix[0], 0, 0, 0, 0, 0, 0, 0, 3, 32}, "test_____32_____byte_____address"...),
		},
		{
			name:  "20 byte address id 1000",
			addr:  sdk.AccAddress("test_20_byte_address"),
			govId: 1000,
			exp:   append([]byte{keeper.ProposalIndexPrefix[0], 0, 0, 0, 0, 0, 0, 3, 232, 20}, "test_20_byte_address"...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.CreateProposalTempIndexKey(tc.govId, tc.addr)
			}
			require.NotPanics(t, testFunc, "CreateProposalTempIndexKey")
			assert.Equal(t, tc.exp, actual, "CreateProposalTempIndexKey result")
			assert.Equal(t, 42, cap(actual), "CreateProposalTempIndexKey result capacity")
		})
	}
}

func TestParseProposalTempIndexKey(t *testing.T) {
	tests := []struct {
		name     string
		key      []byte
		expAddr  sdk.AccAddress
		expPanic string
		expId    uint64
	}{
		{
			name:     "nil",
			key:      nil,
			expPanic: "runtime error: slice bounds out of range [:9] with capacity 0",
		},
		{
			name:     "empty",
			key:      []byte{},
			expPanic: "runtime error: slice bounds out of range [:9] with capacity 0",
		},
		{
			name:     "just one byte",
			key:      []byte{'f'}, // doesn't matter what that byte is.
			expPanic: "runtime error: slice bounds out of range [:9] with capacity 1",
		},
		{
			name:     "only 7 id bytes empty addr",
			key:      []byte{'g', 0, 0, 0, 0, 0, 0, 77},
			expPanic: "runtime error: slice bounds out of range [:9] with capacity 8",
		},
		{
			name:    "id 1 empty addr",
			key:     []byte{'g', 0, 0, 0, 0, 0, 0, 0, 1, 0},
			expId:   1,
			expAddr: sdk.AccAddress{},
		},
		{
			name:    "id 2 4 byte addr",
			key:     []byte{'P', 0, 0, 0, 0, 0, 0, 0, 2, 4, 't', 'e', 's', 't'},
			expId:   2,
			expAddr: sdk.AccAddress("test"),
		},
		{
			name:    "id 3 20 byte addr",
			key:     keeper.CreateProposalTempIndexKey(3, sdk.AccAddress("this_test_addr_is_20")),
			expId:   3,
			expAddr: sdk.AccAddress("this_test_addr_is_20"),
		},
		{
			name:    "id 4 32 byte addr",
			key:     keeper.CreateProposalTempIndexKey(4, sdk.AccAddress("this_test_addr_is_longer_with_32")),
			expId:   4,
			expAddr: sdk.AccAddress("this_test_addr_is_longer_with_32"),
		},
		{
			name:    "id 1000 20 byte addr",
			key:     keeper.CreateProposalTempIndexKey(1000, sdk.AccAddress("this_test_addr_is_20")),
			expId:   1000,
			expAddr: sdk.AccAddress("this_test_addr_is_20"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addr sdk.AccAddress
			var id uint64
			testFunc := func() {
				id, addr = keeper.ParseProposalTempIndexKey(tc.key)
			}
			if len(tc.expPanic) > 0 {
				testutil.RequirePanicsWithMessage(t, tc.expPanic, testFunc, "ParseProposalTempIndexKey")
			} else {
				require.NotPanics(t, testFunc, "ParseProposalTempIndexKey")
				assert.Equal(t, tc.expId, id, "ParseProposalTempIndexKey gov prop id")
				assert.Equal(t, tc.expAddr, addr, "ParseProposalTempIndexKey address")
			}
		})
	}
}
