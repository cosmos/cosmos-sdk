package gas

import (
	"cosmossdk.io/server/v2/core/stf"
)

var _ stf.GasMeter = (*Meter)(nil)

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

func (m *Meter) ConsumeGas(amount stf.Gas, _ string) error {
	if m.limit-m.consumed > amount {
		return stf.ErrOutOfGas
	}
	m.consumed += amount
	return nil
}

func (m *Meter) RefundGas(amount stf.Gas, _ string) error {
	if amount < m.consumed {
		m.consumed -= amount
	}
	return nil
}
