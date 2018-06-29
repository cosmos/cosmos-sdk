package types

// Gas measured by the SDK
type Gas = int64

// Error thrown when out of gas
type ErrorOutOfGas struct {
	Descriptor string
}

// GasMeter interface to track gas consumption
type GasMeter interface {
	GasConsumed() Gas
	ConsumeGas(amount Gas, descriptor string)
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

func (g *basicGasMeter) GasConsumed() Gas {
	return g.consumed
}

func (g *basicGasMeter) ConsumeGas(amount Gas, descriptor string) {
	g.consumed += amount
	if g.consumed > g.limit {
		panic(ErrorOutOfGas{descriptor})
	}
}

type infiniteGasMeter struct {
	consumed Gas
}

func NewInfiniteGasMeter() GasMeter {
	return &infiniteGasMeter{
		consumed: 0,
	}
}

func (g *infiniteGasMeter) GasConsumed() Gas {
	return g.consumed
}

func (g *infiniteGasMeter) ConsumeGas(amount Gas, descriptor string) {
	g.consumed += amount
}

// GasConfig defines gas cost for each operation on KVStores
type GasConfig struct {
	HasCost          Gas
	ReadCostFlat     Gas
	ReadCostPerByte  Gas
	WriteCostFlat    Gas
	WriteCostPerByte Gas
	KeyCostFlat      Gas
	ValueCostFlat    Gas
	ValueCostPerByte Gas
}

var (
	cachedDefaultGasConfig   = DefaultGasConfig()
	cachedTransientGasConfig = TransientGasConfig()
)

// Default gas config for KVStores
func DefaultGasConfig() GasConfig {
	return GasConfig{
		HasCost:          10,
		ReadCostFlat:     10,
		ReadCostPerByte:  1,
		WriteCostFlat:    10,
		WriteCostPerByte: 10,
		KeyCostFlat:      5,
		ValueCostFlat:    10,
		ValueCostPerByte: 1,
	}
}

// Default gas config for TransientStores
func TransientGasConfig() GasConfig {
	// TODO: reduce the gas cost in transient store
	return GasConfig{
		HasCost:          10,
		ReadCostFlat:     10,
		ReadCostPerByte:  1,
		WriteCostFlat:    10,
		WriteCostPerByte: 10,
		KeyCostFlat:      5,
		ValueCostFlat:    10,
		ValueCostPerByte: 1,
	}

}
