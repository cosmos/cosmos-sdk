package types

import "fmt"

var _ GasMeter = &ProxyGasMeter{}

// ProxyGasMeter is like a basicGasMeter, but delegates the gas changes (refund and consume) to the parent GasMeter in
// realtime, ensuring accurate gas accounting even during panics.
type ProxyGasMeter struct {
	GasMeter

	parent GasMeter
}

// NewProxyGasMeter creates a ProxyGasMeter that wraps a parent gas meter, inheriting the minimum of the new limit and the parent's remaining gas.
// It delegates consumption to the parent in real time, ensuring accurate gas accounting even during panics.
//
// Returns the parent directly if the new limit is greater than or equal to its remaining gas.
func NewProxyGasMeter(gasMeter GasMeter, limit Gas) GasMeter {
	limit = min(limit, gasMeter.GasRemaining())
	return &ProxyGasMeter{
		GasMeter: NewGasMeter(limit),
		parent:   gasMeter,
	}
}

// RefundGas will also refund gas to parent gas meter.
func (pgm ProxyGasMeter) RefundGas(amount Gas, descriptor string) {
	pgm.parent.RefundGas(amount, descriptor)
	pgm.GasMeter.RefundGas(amount, descriptor)
}

// ConsumeGas will also consume gas from parent gas meter.
//
// it consume sub-gasmeter first, which means if sub-gasmeter runs out of gas,
// the gas is not charged in parent gas meter, the assumption for business logic
// is the gas is always charged before the work is done, so when out-of-gas panic happens,
// the actual work is not done yet, so we don't need to consume the gas in parent gas meter.
func (pgm ProxyGasMeter) ConsumeGas(amount Gas, descriptor string) {
	pgm.parent.ConsumeGas(amount, descriptor)
	pgm.GasMeter.ConsumeGas(amount, descriptor)
}

func (pgm ProxyGasMeter) String() string {
	return fmt.Sprintf("ProxyGasMeter{consumed: %d, limit: %d}", pgm.GasConsumed(), pgm.Limit())
}
