package coretesting

import (
	"errors"
	"math"

	"cosmossdk.io/core/gas"
)

var _ gas.Meter = (*BasicMeter)(nil)

type BasicMeter struct {
	limit    uint64
	consumed uint64
}

// NewMeter creates a new gas meter with the given gas limit.
// The gas meter keeps track of the gas consumed during execution.
func NewMeter(gasLimit uint64) gas.Meter {
	return &BasicMeter{
		limit:    gasLimit,
		consumed: 0,
	}
}

// Consumed returns the amount of gas consumed by the meter.
func (m *BasicMeter) Consumed() gas.Gas {
	return m.consumed
}

// Limit returns the maximum gas limit allowed for the meter.
func (m *BasicMeter) Limit() gas.Gas {
	return m.limit
}

// Consume consumes the specified amount of gas from the meter.
// It returns an error if the requested gas exceeds the remaining gas limit.
func (m *BasicMeter) Consume(requested gas.Gas, _ string) error {
	remaining := m.limit - m.consumed
	if requested > remaining {
		return gas.ErrOutOfGas
	}
	m.consumed += requested
	return nil
}

// Refund refunds the specified amount of gas.
// If the amount is less than the consumed gas, it subtracts the amount from the consumed gas.
// It returns nil error.
func (m *BasicMeter) Refund(amount gas.Gas, _ string) error {
	if amount < m.consumed {
		m.consumed -= amount
	}
	return nil
}

// Remaining returns the remaining gas limit.
func (m *BasicMeter) Remaining() gas.Gas {
	return m.limit - m.consumed
}

type infiniteGasMeter struct {
	consumed uint64
}

// NewInfiniteGasMeter returns a new gas meter without a limit.
func NewInfiniteGasMeter() gas.Meter {
	return &infiniteGasMeter{
		consumed: 0,
	}
}

// Consumed returns the amount of gas consumed by the meter.
func (m *infiniteGasMeter) Consumed() gas.Gas {
	return m.consumed
}

// Remaining returns MaxUint64 since limit is not confined in infiniteGasMeter.
func (g *infiniteGasMeter) Remaining() gas.Gas {
	return math.MaxUint64
}

// Limit returns MaxUint64 since limit is not confined in infiniteGasMeter.
func (g *infiniteGasMeter) Limit() gas.Gas {
	return math.MaxUint64
}

// Consume adds the given amount of gas to the gas consumed and panics if it overflows the limit.
func (g *infiniteGasMeter) Consume(amount gas.Gas, descriptor string) error {
	var overflow bool
	g.consumed, overflow = addUint64Overflow(g.consumed, amount)
	if overflow {
		return errors.New("gas overflowed")
	}

	return nil
}

// Refund will deduct the given amount from the gas consumed. If the amount is greater than the
// gas consumed, the function will panic.
func (g *infiniteGasMeter) Refund(amount gas.Gas, descriptor string) error {
	if g.consumed < amount {
		return errors.New("negative gas consumed")
	}

	g.consumed -= amount

	return nil
}

// addUint64Overflow performs the addition operation on two uint64 integers and
// returns a boolean on whether or not the result overflows.
func addUint64Overflow(a, b uint64) (uint64, bool) {
	if math.MaxUint64-a < b {
		return 0, true
	}

	return a + b, false
}
