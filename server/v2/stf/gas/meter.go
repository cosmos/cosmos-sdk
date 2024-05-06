package gas

import (
	"cosmossdk.io/core/gas"
)

var _ gas.Meter = (*Meter)(nil)

type Meter struct {
	limit    uint64
	consumed uint64
}

// NewMeter creates a new gas meter with the given gas limit.
// The gas meter keeps track of the gas consumed during execution.
func NewMeter(gasLimit uint64) gas.Meter {
	return &Meter{
		limit:    gasLimit,
		consumed: 0,
	}
}

// Consumed returns the amount of gas consumed by the meter.
func (m *Meter) Consumed() gas.Gas {
	return m.consumed
}

// Limit returns the maximum gas limit allowed for the meter.
func (m *Meter) Limit() gas.Gas {
	return m.limit
}

// Consume consumes the specified amount of gas from the meter.
// It returns an error if the requested gas exceeds the remaining gas limit.
func (m *Meter) Consume(requested gas.Gas, _ string) error {
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
func (m *Meter) Refund(amount gas.Gas, _ string) error {
	if amount < m.consumed {
		m.consumed -= amount
	}
	return nil
}

// Remaining returns the remaining gas limit.
func (m *Meter) Remaining() gas.Gas {
	return m.limit - m.consumed
}
