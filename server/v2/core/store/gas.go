package store

import "fmt"

// Gas defines type alias of uint64 for gas consumption. Gas is measured by the
// SDK for store operations such as Get and Set calls. In addition, callers have
// the ability to explicitly charge gas for costly operations such as signature
// verification.
type Gas = uint64

// GasMeter defines an interface for gas consumption tracking.
type GasMeter interface {
	// GasConsumed returns the amount of gas consumed so far.
	GasConsumed() Gas
	// GasConsumedToLimit returns the gas limit if gas consumed is past the limit,
	// otherwise it returns the consumed gas so far.
	GasConsumedToLimit() Gas
	// GasRemaining returns the gas left in the GasMeter.
	GasRemaining() Gas
	// Limit returns the gas limit (if any).
	Limit() Gas
	// ConsumeGas adds the given amount of gas to the gas consumed and should error
	// if it overflows the gas limit (if any).
	// contract: total consumption must be checked in order to exit early from execution
	ConsumeGas(amount Gas, descriptor string) error
	// RefundGas will deduct the given amount from the gas consumed so far. If the
	// amount is greater than the gas consumed, the function should error.
	RefundGas(amount Gas, descriptor string) error
	// IsPastLimit returns <true> if the gas consumed so far is past the limit (if any),
	// otherwise it returns <false>.
	IsPastLimit() bool
	// IsOutOfGas returns <true> if the gas consumed so far is greater than or equal
	// to gas limit (if any), otherwise it returns <false>.
	IsOutOfGas() bool
}

// GasConfig defines gas cost for each operation on a KVStore.
type GasConfig interface {
	// HasCost should reflect a fixed cost for a Has() call on a store.
	HasCost() Gas
	// DeleteCost should reflect a fixed cost for a Delete() call on a store.
	DeleteCost() Gas
	// ReadCostFlat should reflect a fixed cost for a Get() call on a store.
	ReadCostFlat() Gas
	// ReadCostPerByte should reflect a fixed cost, per-byte on the key and value,
	// for a Get() call on a store. Note, this cost can also be used on iteration
	// seeks.
	ReadCostPerByte() Gas
	// WriteCostFlat should reflect a fixed cost for a Set() call on a store.
	WriteCostFlat() Gas
	// WriteCostPerByte should reflect a fixed cost, per-byte on the key and value,
	// for a Set() call on a store.
	WriteCostPerByte() Gas
	// IterNextCostFlat should reflect a fixed cost for each call to Next() on an
	// iterator.
	IterNextCostFlat() Gas
}

type (
	// ErrorNegativeGasConsumed defines an error thrown when the amount of gas refunded
	// results in a negative gas consumed amount.
	ErrorNegativeGasConsumed struct {
		Message string
	}

	// ErrorOutOfGas defines an error thrown when an action results in out of gas.
	ErrorOutOfGas struct {
		Message string
	}

	// ErrorGasOverflow defines an error thrown when an action results gas consumption
	// unsigned integer overflow.
	ErrorGasOverflow struct {
		Message string
	}
)

func (e ErrorNegativeGasConsumed) Error() string {
	return fmt.Sprintf("negative gas consumed: %s", e.Message)
}

func (e ErrorOutOfGas) Error() string {
	return fmt.Sprintf("out of gas: %s", e.Message)
}

func (e ErrorGasOverflow) Error() string {
	return fmt.Sprintf("gas overflow: %s", e.Message)
}
