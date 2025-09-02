package types

import (
	fmt "fmt"
)

var _ GasMeter = &ProxyGasMeter{}

// ProxyGasMeter wraps another GasMeter, but enforces a lower gas limit.
// Gas consumption is delegated to the wrapped GasMeter, so it won't risk losing gas accounting compared to standalone
// gas meter.
type ProxyGasMeter struct {
	GasMeter

	limit Gas
}

// NewProxyGasMeter returns a new GasMeter which wraps the provided gas meter.
// The remaining is the maximum gas that can be consumed on top of current consumed
// gas of the wrapped gas meter.
//
// If the new remaining is greater than or equal to the existing remaining gas, no wrapping is needed
// and the original gas meter is returned.
func NewProxyGasMeter(gasMeter GasMeter, remaining Gas) GasMeter {
	if remaining >= gasMeter.GasRemaining() {
		return gasMeter
	}

	return &ProxyGasMeter{
		GasMeter: gasMeter,
		limit:    remaining + gasMeter.GasConsumed(),
	}
}

func (pgm ProxyGasMeter) GasRemaining() Gas {
	if pgm.IsPastLimit() {
		return 0
	}
	return pgm.limit - pgm.GasConsumed()
}

func (pgm ProxyGasMeter) Limit() Gas {
	return pgm.limit
}

func (pgm ProxyGasMeter) IsPastLimit() bool {
	return pgm.GasConsumed() > pgm.limit
}

func (pgm ProxyGasMeter) IsOutOfGas() bool {
	return pgm.GasConsumed() >= pgm.limit
}

func (pgm ProxyGasMeter) ConsumeGas(amount Gas, descriptor string) {
	consumed, overflow := addUint64Overflow(pgm.GasMeter.GasConsumed(), amount)
	if overflow {
		panic(ErrorGasOverflow{Descriptor: descriptor})
	}

	if consumed > pgm.limit {
		panic(ErrorOutOfGas{Descriptor: descriptor})
	}

	pgm.GasMeter.ConsumeGas(amount, descriptor)
}

func (pgm ProxyGasMeter) String() string {
	return fmt.Sprintf("ProxyGasMeter{consumed: %d, limit: %d}", pgm.GasConsumed(), pgm.limit)
}
