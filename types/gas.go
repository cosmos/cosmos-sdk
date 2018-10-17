package types

// Gas consumption descriptors.
const (
	GasIterNextCostFlatDesc = "IterNextFlat"
	GasValuePerByteDesc     = "ValuePerByte"
	GasWritePerByteDesc     = "WritePerByte"
	GasReadPerByteDesc      = "ReadPerByte"
	GasWriteCostFlatDesc    = "WriteFlat"
	GasReadCostFlatDesc     = "ReadFlat"
	GasHasDesc              = "Has"
	GasDeleteDesc           = "Delete"
)

var (
	cachedKVGasConfig        = KVGasConfig()
	cachedTransientGasConfig = TransientGasConfig()
)

// Gas measured by the SDK
type Gas = int64

// ErrorOutOfGas defines an error thrown when an action results in out of gas.
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

// NewGasMeter returns a reference to a new basicGasMeter.
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

// NewInfiniteGasMeter returns a reference to a new infiniteGasMeter.
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
	DeleteCost       Gas
	ReadCostFlat     Gas
	ReadCostPerByte  Gas
	WriteCostFlat    Gas
	WriteCostPerByte Gas
	ValueCostPerByte Gas
	IterNextCostFlat Gas
}

// KVGasConfig returns a default gas config for KVStores.
func KVGasConfig() GasConfig {
	return GasConfig{
		HasCost:          10,
		DeleteCost:       10,
		ReadCostFlat:     10,
		ReadCostPerByte:  1,
		WriteCostFlat:    10,
		WriteCostPerByte: 10,
		ValueCostPerByte: 1,
		IterNextCostFlat: 15,
	}
}

// TransientGasConfig returns a default gas config for TransientStores.
func TransientGasConfig() GasConfig {
	// TODO: define gasconfig for transient stores
	return KVGasConfig()
}
