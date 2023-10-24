package store

import "fmt"

// Gas defines type alias of uint64 for gas consumption. Gas is measured by the
// SDK for store operations such as Get and Set calls. In addition, callers have
// the ability to explicitly charge gas for costly operations such as signature
// verification.
type Gas uint64

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

type (
	// ErrorNegativeGasConsumed defines an error thrown when the amount of gas refunded
	// results in a negative gas consumed amount.
	ErrorNegativeGasConsumed struct {
		Descriptor string
	}

	// ErrorOutOfGas defines an error thrown when an action results in out of gas.
	ErrorOutOfGas struct {
		Descriptor string
	}

	// ErrorGasOverflow defines an error thrown when an action results gas consumption
	// unsigned integer overflow.
	ErrorGasOverflow struct {
		Descriptor string
	}
)

// GasMeter defines an interface for gas consumption tracking.
type GasMeter interface {
	GasConsumed() Gas
	GasConsumedToLimit() Gas
	GasRemaining() Gas
	Limit() Gas
	ConsumeGas(amount Gas, descriptor string)
	RefundGas(amount Gas, descriptor string)
	IsPastLimit() bool
	IsOutOfGas() bool

	fmt.Stringer
}
