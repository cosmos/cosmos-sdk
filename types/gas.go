package types

import (
	storetypes "cosmossdk.io/store/types"
)

// Wrapper Error for store/v1 ErrorOutOfGas, ErrorNegativeGasConsumed and ErrorGasOverflow so that
// modules don't have to import storev1 directly

// ErrorNegativeGasConsumed defines an error thrown when the amount of gas refunded results in a
// negative gas consumed amount.
type ErrorNegativeGasConsumed = storetypes.ErrorNegativeGasConsumed

// ErrorOutOfGas defines an error thrown when an action results in out of gas.
type ErrorOutOfGas = storetypes.ErrorOutOfGas

// ErrorGasOverflow defines an error thrown when an action results gas consumption
// unsigned integer overflow.
type ErrorGasOverflow = storetypes.ErrorGasOverflow
