package gas

import (
	"cosmossdk.io/server/v2/core/stf"
	"cosmossdk.io/server/v2/core/store"
)

// DefaultWrapWithGasMeter defines the default wrap with gas meter function in stf. In case
// the meter sets as limit stf.NoGasLimit, then a fast path is taken and the store.WriterMap
// is returned.
func DefaultWrapWithGasMeter(meter stf.GasMeter, state store.WriterMap) store.WriterMap {
	if meter.Limit() == stf.NoGasLimit {
		return state
	}
	return NewMeteredWriterMap(DefaultConfig, meter, state)
}

// DefaultGetMeter returns the default gas meter. In case it is stf.NoGasLimit a NoOpMeter is returned.
func DefaultGetMeter(gasLimit uint64) stf.GasMeter {
	if gasLimit == stf.NoGasLimit {
		return NoOpMeter{}
	}
	return NewMeter(gasLimit)
}

var DefaultConfig = StoreConfig{
	HasCost:          1000,
	DeleteCostFlat:   1000,
	ReadCostFlat:     1000,
	ReadCostPerByte:  3,
	WriteCostFlat:    2000,
	WriteCostPerByte: 30,
	IterNextCostFlat: 30,
}

type NoOpMeter struct {
}

func (n NoOpMeter) GasConsumed() stf.Gas { return 0 }

func (n NoOpMeter) Limit() stf.Gas { return stf.NoGasLimit }

func (n NoOpMeter) ConsumeGas(_ stf.Gas, _ string) error { return nil }

func (n NoOpMeter) RefundGas(_ stf.Gas, _ string) error { return nil }
