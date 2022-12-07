package keeper_test

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// SendRestrictionArgs are the args provided to a SendRestrictionFn function.
type SendRestrictionArgs struct {
	FromAddr sdk.AccAddress
	ToAddr   sdk.AccAddress
	Coins    sdk.Coins
}

// SendRestrictionTestHelper is a struct with stuff helpful for testing the SendRestrictionFn stuff.
type SendRestrictionTestHelper struct {
	Calls []*SendRestrictionArgs
}

func NewSendRestrictionTestHelper() *SendRestrictionTestHelper {
	return &SendRestrictionTestHelper{Calls: make([]*SendRestrictionArgs, 0, 2)}
}

// RecordCall makes note that the provided args were used as a funcion call.
func (s *SendRestrictionTestHelper) RecordCall(fromAddr, toAddr sdk.AccAddress, coins sdk.Coins) {
	s.Calls = append(s.Calls, s.NewArgs(fromAddr, toAddr, coins))
}

// NewCalls is just a shorter way to create a []*SendRestrictionArgs.
func (s *SendRestrictionTestHelper) NewCalls(args ...*SendRestrictionArgs) []*SendRestrictionArgs {
	return args
}

// NewArgs creates a new SendRestrictionArgs.
func (s *SendRestrictionTestHelper) NewArgs(fromAddr, toAddr sdk.AccAddress, coins sdk.Coins) *SendRestrictionArgs {
	return &SendRestrictionArgs{
		FromAddr: fromAddr,
		ToAddr:   toAddr,
		Coins:    coins,
	}
}

// NoOpRestriction creates a new SendRestrictionFn function that records the arguments it's called with and returns the provided toAddr.
func (s *SendRestrictionTestHelper) NoOpRestriction() keeper.SendRestrictionFn {
	return func(_ sdk.Context, fromAddr, toAddr sdk.AccAddress, coins sdk.Coins) (sdk.AccAddress, error) {
		s.RecordCall(fromAddr, toAddr, coins)
		return toAddr, nil
	}
}

// NewToRestriction creates a new SendRestrictionFn function that returns a different toAddr than provided.
func (s *SendRestrictionTestHelper) NewToRestriction(addr sdk.AccAddress) keeper.SendRestrictionFn {
	return func(_ sdk.Context, fromAddr, toAddr sdk.AccAddress, coins sdk.Coins) (sdk.AccAddress, error) {
		s.RecordCall(fromAddr, toAddr, coins)
		return addr, nil
	}
}

// ErrorRestriction creates a new SendRestrictionFn function that returns a nil toAddr and an error.
func (s *SendRestrictionTestHelper) ErrorRestriction(message string) keeper.SendRestrictionFn {
	return func(_ sdk.Context, fromAddr, toAddr sdk.AccAddress, coins sdk.Coins) (sdk.AccAddress, error) {
		s.RecordCall(fromAddr, toAddr, coins)
		return nil, errors.New(message)
	}
}

// SendRestrictionTestParams are parameters to test regarding calling a SendRestrictionFn.
type SendRestrictionTestParams struct {
	// ExpNil is whether to expect the provided SendRestrictionFn to be nil.
	// If it is true, the rest of these test params are ignored.
	ExpNil bool
	// FromAddr is the SendRestrictionFn fromAddr input.
	FromAddr sdk.AccAddress
	// ToAddr is the SendRestrictionFn toAddr input.
	ToAddr sdk.AccAddress
	// Coins is the SendRestrictionFn coins input.
	Coins sdk.Coins
	// ExpAddr is the expected return address.
	ExpAddr sdk.AccAddress
	// ExpErr is the expected return error string.
	ExpErr string
	// ExpCalls is the args of all the SendRestrictionFn calls that end up being made.
	ExpCalls []*SendRestrictionArgs
}

// TestActual tests the provided SendRestrictionFn using the provided test parameters.
func (s *SendRestrictionTestHelper) TestActual(t *testing.T, tp *SendRestrictionTestParams, actual keeper.SendRestrictionFn) {
	t.Helper()
	if tp.ExpNil {
		require.Nil(t, actual, "ComposeSendRestrictions result")
	} else {
		require.NotNil(t, actual, "ComposeSendRestrictions result")
		s.Calls = s.Calls[:0]
		addr, err := actual(sdk.Context{}, tp.FromAddr, tp.ToAddr, tp.Coins)
		if len(tp.ExpErr) != 0 {
			assert.EqualError(t, err, tp.ExpErr, "composite SendRestrictionFn output error")
		} else {
			assert.NoError(t, err, "composite SendRestrictionFn output error")
		}
		assert.Equal(t, tp.ExpAddr, addr, "composite SendRestrictionFn output address")
		assert.Equal(t, tp.ExpCalls, s.Calls, "args given to funcs in composite SendRestrictionFn")
	}
}

func TestSendRestriction_Then(t *testing.T) {
	frAddr := sdk.AccAddress("fromaddr____________")
	addr0 := sdk.AccAddress("0addr_______________")
	addr1 := sdk.AccAddress("1addr_______________")
	addr2 := sdk.AccAddress("2addr_______________")
	addr3 := sdk.AccAddress("3addr_______________")
	addr4 := sdk.AccAddress("4addr_______________")
	coins := sdk.NewCoins(sdk.NewInt64Coin("acoin", 2), sdk.NewInt64Coin("bcoin", 4))

	h := NewSendRestrictionTestHelper()

	tests := []struct {
		name   string
		base   keeper.SendRestrictionFn
		second keeper.SendRestrictionFn
		exp    *SendRestrictionTestParams
	}{
		{
			name:   "nil nil",
			base:   nil,
			second: nil,
			exp: &SendRestrictionTestParams{
				ExpNil: true,
			},
		},
		{
			name:   "nil noop",
			base:   nil,
			second: h.NoOpRestriction(),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr1,
				Coins:    coins,
				ExpAddr:  addr1,
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr1, coins)),
			},
		},
		{
			name:   "noop nil",
			base:   h.NoOpRestriction(),
			second: nil,
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr1,
				Coins:    coins,
				ExpAddr:  addr1,
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr1, coins)),
			},
		},
		{
			name:   "noop noop",
			base:   h.NoOpRestriction(),
			second: h.NoOpRestriction(),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr1,
				Coins:    coins,
				ExpAddr:  addr1,
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr1, coins), h.NewArgs(frAddr, addr1, coins)),
			},
		},
		{
			name:   "setter setter",
			base:   h.NewToRestriction(addr2),
			second: h.NewToRestriction(addr3),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr1,
				Coins:    coins,
				ExpAddr:  addr3,
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr1, coins), h.NewArgs(frAddr, addr2, coins)),
			},
		},
		{
			name:   "setter error",
			base:   h.NewToRestriction(addr2),
			second: h.ErrorRestriction("this is a test error"),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr1,
				Coins:    coins,
				ExpAddr:  nil,
				ExpErr:   "this is a test error",
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr1, coins), h.NewArgs(frAddr, addr2, coins)),
			},
		},
		{
			name:   "error setter",
			base:   h.ErrorRestriction("another test error"),
			second: h.NewToRestriction(addr3),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr1,
				Coins:    coins,
				ExpAddr:  nil,
				ExpErr:   "another test error",
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr1, coins)),
			},
		},
		{
			name:   "error error",
			base:   h.ErrorRestriction("first test error"),
			second: h.ErrorRestriction("second test error"),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr1,
				Coins:    coins,
				ExpAddr:  nil,
				ExpErr:   "first test error",
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr1, coins)),
			},
		},
		{
			name:   "double chain",
			base:   keeper.ComposeSendRestrictions(h.NewToRestriction(addr1), h.NewToRestriction(addr2)),
			second: keeper.ComposeSendRestrictions(h.NewToRestriction(addr3), h.NewToRestriction(addr4)),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr0,
				Coins:    coins,
				ExpAddr:  addr4,
				ExpCalls: h.NewCalls(
					h.NewArgs(frAddr, addr0, coins),
					h.NewArgs(frAddr, addr1, coins),
					h.NewArgs(frAddr, addr2, coins),
					h.NewArgs(frAddr, addr3, coins),
				),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual keeper.SendRestrictionFn
			testFunc := func() {
				actual = tc.base.Then(tc.second)
			}
			require.NotPanics(t, testFunc, "SendRestrictionFn.Then")
			h.TestActual(t, tc.exp, actual)
		})
	}
}

func TestComposeSendRestrictions(t *testing.T) {
	rz := func(rs ...keeper.SendRestrictionFn) []keeper.SendRestrictionFn {
		return rs
	}
	frAddr := sdk.AccAddress("fromaddr____________")
	addr0 := sdk.AccAddress("0addr_______________")
	addr1 := sdk.AccAddress("1addr_______________")
	addr2 := sdk.AccAddress("2addr_______________")
	addr3 := sdk.AccAddress("3addr_______________")
	addr4 := sdk.AccAddress("4addr_______________")
	coins := sdk.NewCoins(sdk.NewInt64Coin("acoin", 2), sdk.NewInt64Coin("bcoin", 4))

	h := NewSendRestrictionTestHelper()

	tests := []struct {
		name  string
		input []keeper.SendRestrictionFn
		exp   *SendRestrictionTestParams
	}{
		{
			name:  "nil list",
			input: nil,
			exp: &SendRestrictionTestParams{
				ExpNil: true,
			},
		},
		{
			name:  "empty list",
			input: rz(),
			exp: &SendRestrictionTestParams{
				ExpNil: true,
			},
		},
		{
			name:  "only nil entry",
			input: rz(nil),
			exp: &SendRestrictionTestParams{
				ExpNil: true,
			},
		},
		{
			name:  "five nil entries",
			input: rz(nil, nil, nil, nil, nil),
			exp: &SendRestrictionTestParams{
				ExpNil: true,
			},
		},
		{
			name:  "only noop entry",
			input: rz(h.NoOpRestriction()),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr0,
				Coins:    coins,
				ExpAddr:  addr0,
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr0, coins)),
			},
		},
		{
			name:  "only error entry",
			input: rz(h.ErrorRestriction("test error")),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr0,
				Coins:    coins,
				ExpAddr:  nil,
				ExpErr:   "test error",
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr0, coins)),
			},
		},
		{
			name:  "noop nil nil",
			input: rz(h.NoOpRestriction(), nil, nil),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr0,
				Coins:    coins,
				ExpAddr:  addr0,
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr0, coins)),
			},
		},
		{
			name:  "nil noop nil",
			input: rz(nil, h.NoOpRestriction(), nil),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr1,
				Coins:    coins,
				ExpAddr:  addr1,
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr1, coins)),
			},
		},
		{
			name:  "nil nil noop",
			input: rz(nil, nil, h.NoOpRestriction()),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr2,
				Coins:    coins,
				ExpAddr:  addr2,
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr2, coins)),
			},
		},
		{
			name:  "noop noop nil",
			input: rz(h.NoOpRestriction(), h.NoOpRestriction(), nil),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr0,
				Coins:    coins,
				ExpAddr:  addr0,
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr0, coins), h.NewArgs(frAddr, addr0, coins)),
			},
		},
		{
			name:  "noop nil noop",
			input: rz(h.NoOpRestriction(), nil, h.NoOpRestriction()),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr1,
				Coins:    coins,
				ExpAddr:  addr1,
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr1, coins), h.NewArgs(frAddr, addr1, coins)),
			},
		},
		{
			name:  "nil noop noop",
			input: rz(nil, h.NoOpRestriction(), h.NoOpRestriction()),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr2,
				Coins:    coins,
				ExpAddr:  addr2,
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr2, coins), h.NewArgs(frAddr, addr2, coins)),
			},
		},
		{
			name:  "noop noop noop",
			input: rz(h.NoOpRestriction(), h.NoOpRestriction(), h.NoOpRestriction()),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr3,
				Coins:    coins,
				ExpAddr:  addr3,
				ExpCalls: h.NewCalls(
					h.NewArgs(frAddr, addr3, coins),
					h.NewArgs(frAddr, addr3, coins),
					h.NewArgs(frAddr, addr3, coins),
				),
			},
		},
		{
			name:  "err noop noop",
			input: rz(h.ErrorRestriction("first error"), h.NoOpRestriction(), h.NoOpRestriction()),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr4,
				Coins:    coins,
				ExpAddr:  nil,
				ExpErr:   "first error",
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr4, coins)),
			},
		},
		{
			name:  "noop err noop",
			input: rz(h.NoOpRestriction(), h.ErrorRestriction("second error"), h.NoOpRestriction()),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr4,
				Coins:    coins,
				ExpAddr:  nil,
				ExpErr:   "second error",
				ExpCalls: h.NewCalls(h.NewArgs(frAddr, addr4, coins), h.NewArgs(frAddr, addr4, coins)),
			},
		},
		{
			name:  "noop noop err",
			input: rz(h.NoOpRestriction(), h.NoOpRestriction(), h.ErrorRestriction("third error")),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr4,
				Coins:    coins,
				ExpAddr:  nil,
				ExpErr:   "third error",
				ExpCalls: h.NewCalls(
					h.NewArgs(frAddr, addr4, coins),
					h.NewArgs(frAddr, addr4, coins),
					h.NewArgs(frAddr, addr4, coins),
				),
			},
		},
		{
			name:  "new-to err err",
			input: rz(h.NewToRestriction(addr0), h.ErrorRestriction("second error"), h.ErrorRestriction("third error")),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr4,
				Coins:    coins,
				ExpAddr:  nil,
				ExpErr:   "second error",
				ExpCalls: h.NewCalls(
					h.NewArgs(frAddr, addr4, coins),
					h.NewArgs(frAddr, addr0, coins),
				),
			},
		},
		{
			name: "big bang",
			input: rz(
				h.NoOpRestriction(), nil, h.NewToRestriction(addr1), // Called with orig toAddr.
				nil, h.NoOpRestriction(), h.NewToRestriction(addr2), // Called with addr1 toAddr.
				h.NewToRestriction(addr3),                           // Called with addr2 toAddr.
				nil, h.NoOpRestriction(), h.NewToRestriction(addr4), // Called with addr3 toAddr.
				nil, h.NoOpRestriction(), nil, nil, h.ErrorRestriction("oops, an error"), // Called with addr4 toAddr.
				h.NewToRestriction(addr0), nil, h.NoOpRestriction(), // Not called.
			),
			exp: &SendRestrictionTestParams{
				FromAddr: frAddr,
				ToAddr:   addr0,
				Coins:    coins,
				ExpAddr:  nil,
				ExpErr:   "oops, an error",
				ExpCalls: h.NewCalls(
					h.NewArgs(frAddr, addr0, coins),
					h.NewArgs(frAddr, addr0, coins),
					h.NewArgs(frAddr, addr1, coins),
					h.NewArgs(frAddr, addr1, coins),
					h.NewArgs(frAddr, addr2, coins),
					h.NewArgs(frAddr, addr3, coins),
					h.NewArgs(frAddr, addr3, coins),
					h.NewArgs(frAddr, addr4, coins),
					h.NewArgs(frAddr, addr4, coins),
				),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual keeper.SendRestrictionFn
			testFunc := func() {
				actual = keeper.ComposeSendRestrictions(tc.input...)
			}
			require.NotPanics(t, testFunc, "ComposeSendRestrictions")
			h.TestActual(t, tc.exp, actual)
		})
	}
}
