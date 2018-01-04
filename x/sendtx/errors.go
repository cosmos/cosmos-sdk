package sendtx

import "fmt"

// TODO! Deal coherently with this and coinstore/errors.go

func ErrNoInputs() error {
	return fmt.Errorf("No inputs")
}

func ErrNoOutputs() error {
	return fmt.Errorf("No outputs")
}

func ErrInvalidSequence(seq int64) error {
	return fmt.Errorf("Bad sequence %d", seq)
}
