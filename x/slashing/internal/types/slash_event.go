package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SlashEvent defines a recent slash for a validator
type SlashEvent struct {
	Address      sdk.ValAddress `json:"address" yaml:"address"`           // validator address
	Power        sdk.Dec        `json:"start_height" yaml:"start_height"` // height at which validator was first a candidate OR was unjailed
	SlashedSoFar sdk.Dec        `json:"index_offset" yaml:"index_offset"` // index offset into signed block bit array
}

// NewValidatorSigningInfo creates a new ValidatorSigningInfo instance
func NewSlashEvent(
	valAddress sdk.ValAddress, power sdk.Dec,
	slashedSoFar sdk.Dec,
) SlashEvent {

	return SlashEvent{
		Address:      valAddress,
		Power:        power,
		SlashedSoFar: slashedSoFar,
	}
}

// String implements the stringer interface for ValidatorSigningInfo
func (i SlashEvent) String() string {
	return fmt.Sprintf(`Slash Event:
	Address:               %s
	Power:                 %d
	Slashed So Far:        %d`,
		i.Address, i.Power, i.SlashedSoFar)
}
