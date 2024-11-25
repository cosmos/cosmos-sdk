package types

// NewParams creates a new parameter configuration for the bank/v2 module
func NewParams() Params {
	return Params{}
}

// DefaultParams is the default parameter configuration for the bank/v2 module
func DefaultParams() Params {
	return NewParams()
}

// Validate all bank/v2 module parameters
func (p Params) Validate() error {
	return nil
}
