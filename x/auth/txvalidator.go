package auth

import (
	"fmt"

	"cosmossdk.io/x/auth/ante"
	"cosmossdk.io/x/auth/types"
)

// TxValidatorOptions defines the keepers needed to run checks in TxValidator
type TxValidatorOptions struct {
	BankKeeper     types.BankKeeper
	FeegrantKeeper ante.FeegrantKeeper
}

// Validate validates the options of tx validator
func (options TxValidatorOptions) Validate() error {
	if options.BankKeeper == nil {
		return fmt.Errorf("no bank keeper found for tx validator")
	}

	if options.FeegrantKeeper == nil {
		return fmt.Errorf("no feegrant keeper found for tx validator")
	}

	return nil
}
