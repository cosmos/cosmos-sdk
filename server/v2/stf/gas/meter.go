package gas

import (
	"cosmossdk.io/server/v2/core/stf"
)

var _ stf.GasMeter = (*Meter)(nil)

func NewMeter(gasLimit uint64) *Meter {
	return &Meter{
		limit:    gasLimit,
		consumed: 0,
	}
}

type Meter struct {
	limit    uint64
	consumed uint64
}

func (m *Meter) GasConsumed() stf.Gas {
	return m.consumed
}

func (m *Meter) Limit() stf.Gas {
	return m.limit
}

func (m *Meter) ConsumeGas(requested stf.Gas, _ string) error {
	remaining := m.limit - m.consumed
	if requested > remaining {
		return stf.ErrOutOfGas
	}
	m.consumed += requested
	return nil
}

func (m *Meter) RefundGas(amount stf.Gas, _ string) error {
	if amount < m.consumed {
		m.consumed -= amount
	}
	return nil
}
