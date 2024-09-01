package gas

import (
	coregas "cosmossdk.io/core/gas"
	"cosmossdk.io/core/store"
)

// DefaultWrapWithGasMeter defines the default wrap with gas meter function in stf. In case
// the meter sets as limit stf.NoGasLimit, then a fast path is taken and the store.WriterMap
// is returned.
func DefaultWrapWithGasMeter(meter coregas.Meter, state store.WriterMap) store.WriterMap {
	if meter.Limit() == coregas.NoGasLimit {
		return state
	}
	return NewMeteredWriterMap(DefaultConfig, meter, state)
}

// DefaultGasMeter returns the default gas meter. In case it is coregas.NoGasLimit a NoOpMeter is returned.
func DefaultGasMeter(gasLimit uint64) coregas.Meter {
	if gasLimit == coregas.NoGasLimit {
		return NoOpMeter{}
	}
	return NewMeter(gasLimit)
}

// DefaultGasConfig returns the default gas config.
// Unless overridden, the default gas costs are:
var DefaultConfig = coregas.GasConfig{
	HasCost:          1000,
	DeleteCost:       1000,
	ReadCostFlat:     1000,
	ReadCostPerByte:  3,
	WriteCostFlat:    2000,
	WriteCostPerByte: 30,
	IterNextCostFlat: 30,
}

type NoOpMeter struct{}

func (n NoOpMeter) Consumed() coregas.Gas { return 0 }

func (n NoOpMeter) Limit() coregas.Gas { return coregas.NoGasLimit }

func (n NoOpMeter) Consume(_ coregas.Gas, _ string) error { return nil }

func (n NoOpMeter) Refund(_ coregas.Gas, _ string) error { return nil }

func (n NoOpMeter) Remaining() coregas.Gas { return coregas.NoGasLimit }
