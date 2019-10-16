package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SlashEvent defines a recent slash for a validator
type SlashEvent struct {
	Address          sdk.ConsAddress `json:"address" yaml:"address"` // validator's cons address
	Power            int64           `json:"power" yaml:"power"`
	InfractionHeight int64           `json:"infraction_height" yaml:"infraction_height"`
	PercentPower     sdk.Dec         `json:"voting_percent" yaml:"voting_percent"`
	SlashedSoFar     sdk.Dec         `json:"slashed_so_far" yaml:"slashed_so_far"`
	EndTime          time.Time       `json:"end_time" yaml:"end_time"` // time when SlashEvent gets pruned
}

// NewValidatorSigningInfo creates a new ValidatorSigningInfo instance
func NewSlashEvent(
	consAddress sdk.ConsAddress, power int64, infractionHeight int64,
	percentPower sdk.Dec, slashedSoFar sdk.Dec, endTime time.Time,
) SlashEvent {

	return SlashEvent{
		Address:          consAddress,
		Power:            power,
		InfractionHeight: infractionHeight,
		PercentPower:     percentPower,
		SlashedSoFar:     slashedSoFar,
		EndTime:          endTime,
	}
}

// String implements the stringer interface for ValidatorSigningInfo
func (i SlashEvent) String() string {
	return fmt.Sprintf(`Slash Event:
	Address:               %s
	Power:                 %d
	Infraction Height:     %d
	Percent Power:         &d
	Slashed So Far:        %d
	End Time:              %s`,
		i.Address, i.Power, i.InfractionHeight, i.PercentPower, i.SlashedSoFar, i.EndTime)
}

func (i SlashEvent) StoreKey() []byte {
	return append(sdk.FormatTimeBytes(i.EndTime), i.Address.Bytes()...)
}
