package slashing

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Stored by *validator* address (not owner address)
func (k Keeper) getValidatorSlashingPeriod(ctx sdk.Context, address sdk.ValAddress, startHeight int64) (slashingPeriod ValidatorSlashingPeriod) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(GetValidatorSlashingPeriodKey(address, startHeight))
	k.cdc.MustUnmarshalBinary(bz, &slashingPeriod)
	return
}

// Stored by *validator* address (not owner address)
func (k Keeper) setValidatorSlashingPeriod(ctx sdk.Context, address sdk.ValAddress, startHeight int64, slashingPeriod ValidatorSlashingPeriod) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(slashingPeriod)
	store.Set(GetValidatorSlashingPeriodKey(address, startHeight), bz)
}

// Construct a new `ValidatorSlashingPeriod` struct
func NewValidatorSlashingPeriod(startHeight int64, endHeight int64, slashedSoFar sdk.Dec) ValidatorSlashingPeriod {
	return ValidatorSlashingPeriod{
		StartHeight:  startHeight,
		EndHeight:    endHeight,
		SlashedSoFar: slashedSoFar,
	}
}

// Slashing period for a validator
type ValidatorSlashingPeriod struct {
	StartHeight  int64   `json:"start_height"`   // starting height of the slashing period
	EndHeight    int64   `json:"end_height"`     // ending height of the slashing period, or sentinel value of 0 for in-progress
	SlashedSoFar sdk.Dec `json:"slashed_so_far"` // fraction of validator stake slashed so far in this slashing period
}

// Return human readable slashing period
func (p ValidatorSlashingPeriod) HumanReadableString() string {
	return fmt.Sprintf("Start height: %d, end height: %d, slashed so far: %v",
		p.StartHeight, p.EndHeight, p.SlashedSoFar)
}
