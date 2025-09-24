package types

import (
	"fmt"
)

var _ GasMeter = &ProxyGasMeter{}

// ProxyGasMeter is like a basicGasMeter, but delegates the gas changes (refund and consume) to the parent GasMeter in
// realtime, so it won't risk losing gas accounting in face of error return or panics.
type ProxyGasMeter struct {
	GasMeter

	parent GasMeter
}

// NewProxyGasMeter returns a new ProxyGasMeter which is like a basic gas meter with minimum of new limit and remaining gas
// of the parent gas meter, it also delegate the gas consumption to parent gas meter in real time, so it won't risk
// losing gas accounting in face of panics or other unexpected errors.
//
// If limit is greater than or equal to the remaining gas, no wrapping is needed and the original gas meter is returned.
func NewProxyGasMeter(gasMeter GasMeter, limit Gas) GasMeter {
	limit = min(limit, gasMeter.GasRemaining())
	return &ProxyGasMeter{
		GasMeter: NewGasMeter(limit),
		parent:   gasMeter,
	}
}

// RefundGas will also refund gas to parent gas meter.
func (pgm ProxyGasMeter) RefundGas(amount Gas, descriptor string) {
	pgm.GasMeter.RefundGas(amount, descriptor)
	pgm.parent.RefundGas(amount, descriptor)
}

// ConsumeGas will also consume gas from parent gas meter.
//
// it consume sub-gasmeter first, which means if sub-gasmeter runs out of gas,
// the gas is not charged in parent gas meter, the assumption for business logic
// is the gas is always charged before the work is done, so when out-of-gas panic happens,
// the actual work is not done yet, so we don't need to consume the gas in parent gas meter.
func (pgm ProxyGasMeter) ConsumeGas(amount Gas, descriptor string) {
	pgm.GasMeter.ConsumeGas(amount, descriptor)
	pgm.parent.ConsumeGas(amount, descriptor)
}

func (pgm ProxyGasMeter) String() string {
	return fmt.Sprintf("ProxyGasMeter{consumed: %d, limit: %d}", pgm.GasConsumed(), pgm.Limit())
}
