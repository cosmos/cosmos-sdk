package math

import (
	"errors"
)

var (
	// ErrIntOverflow is the error returned when an integer overflow occurs
	ErrIntOverflow = errors.New("Integer overflow")

	// ErrDivideByZero is the error returned when a divide by zero occurs
	ErrDivideByZero = errors.New("Divide by zero")

	// Decimal errors
	ErrLegacyEmptyDecimalStr      = errors.New("decimal string cannot be empty")
	ErrLegacyInvalidDecimalLength = errors.New("invalid decimal length")
	ErrLegacyInvalidDecimalStr    = errors.New("invalid decimal string")
)
