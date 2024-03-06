package gas

import (
	"fmt"

	"cosmossdk.io/core/gas"
)

var _ gas.Meter = (*Meter)(nil)

func NewMeter(gasLimit uint64) gas.Meter {
	return &Meter{
		limit:    gasLimit,
		consumed: 0,
	}
}

type Meter struct {
	limit    uint64
	consumed uint64
}

func (m *Meter) Consumed() gas.Gas {
	return m.consumed
}

func (m *Meter) Limit() gas.Gas {
	return m.limit
}

func (m *Meter) Consume(requested gas.Gas, _ string) error {
	remaining := m.limit - m.consumed
	if requested > remaining {
		m.consumed = m.limit // this sets it as fully consumed aka remaining is 0.
		return fmt.Errorf("%w: requested %d remaining %d", gas.ErrOutOfGas, requested, remaining)
	}
	m.consumed += requested
	return nil
}

func (m *Meter) Refund(amount gas.Gas, _ string) error {
	if amount < m.consumed {
		m.consumed -= amount
	}
	return nil
}
