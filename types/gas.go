package types

import ()

// Gas measured by the SDK
type Gas = int64

// GasMeter interface to track gas consumption
type GasMeter interface {
	GasExceeded() bool
	GasConsumed() Gas
	ConsumeGas(amount Gas)
	ConsumeGasOrFail(amount Gas) bool
}

type basicGasMeter struct {
	limit    Gas
	consumed Gas
}

func NewGasMeter(limit Gas) GasMeter {
	return &basicGasMeter{
		limit:    limit,
		consumed: 0,
	}
}

func (g *basicGasMeter) GasExceeded() bool {
	return g.consumed > g.limit
}

func (g *basicGasMeter) GasConsumed() Gas {
	return g.consumed
}

func (g *basicGasMeter) ConsumeGas(amount Gas) {
	g.consumed += amount
}

func (g *basicGasMeter) ConsumeGasOrFail(amount Gas) bool {
	g.ConsumeGas(amount)
	return g.GasExceeded()
}
