package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate performs a basic validation of the coin metadata fields
func (m Metadata) Validate() error {
	if err := sdk.ValidateDenom(m.Base); err != nil {
		return fmt.Errorf("invalid metadata base denom: %w", err)
	}

	if err := sdk.ValidateDenom(m.Display); err != nil {
		return fmt.Errorf("invalid metadata display denom: %w", err)
	}

	seenUnits := make(map[string]bool, 0)
	for _, denomUnit := range m.DenomUnits {
		if seenUnits[denomUnit.Denom] {
			return fmt.Errorf("duplicate denomination unit %s", denomUnit.Denom)
		}

		if err := denomUnit.Validate(); err != nil {
			return err
		}

		seenUnits[denomUnit.Denom] = true
	}

	return nil
}

// Validate performs a basic validation of the denomination unit fields
func (du DenomUnit) Validate() error {
	if err := sdk.ValidateDenom(du.Denom); err != nil {
		return fmt.Errorf("invalid denom unit: %w", err)
	}

	for _, alias := range du.Aliases {
		if strings.TrimSpace(alias) == "" {
			return fmt.Errorf("alias for denom unit %s cannot be blank", du.Denom)
		}
	}

	return nil
}
