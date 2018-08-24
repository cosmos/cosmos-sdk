package slashing

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Cap an infraction's slash amount by the slashing period in which it was committed
func (k Keeper) capBySlashingPeriod(ctx sdk.Context, address sdk.ValAddress, fraction sdk.Dec, infractionHeight int64) (revisedFraction sdk.Dec) {

	// Calculate total amount to be slashed
	slashingPeriod := k.getValidatorSlashingPeriodForHeight(ctx, address, infractionHeight)
	totalToSlash := sdk.MaxDec(slashingPeriod.SlashedSoFar, fraction)
	slashingPeriod.SlashedSoFar = totalToSlash
	k.setValidatorSlashingPeriod(ctx, slashingPeriod)

	// Calculate remainder
	revisedFraction = slashingPeriod.SlashedSoFar.Sub(totalToSlash)
	return

}

// Stored by *validator* address (not owner address)
func (k Keeper) getValidatorSlashingPeriodForHeight(ctx sdk.Context, address sdk.ValAddress, height int64) (slashingPeriod ValidatorSlashingPeriod) {
	store := ctx.KVStore(k.storeKey)
	start := GetValidatorSlashingPeriodKey(address, height)
	end := sdk.PrefixEndBytes(GetValidatorSlashingPeriodPrefix(address))
	iterator := store.Iterator(start, end)
	if !iterator.Valid() {
		panic("expected to find slashing period, but none was found")
	}
	slashingPeriod = k.unmarshalSlashingPeriodKeyValue(iterator.Key(), iterator.Value())
	if slashingPeriod.EndHeight < height {
		panic("slashing period ended before infraction")
	}
	return
}

// Stored by *validator* address (not owner address)
func (k Keeper) setValidatorSlashingPeriod(ctx sdk.Context, slashingPeriod ValidatorSlashingPeriod) {
	slashingPeriodValue := ValidatorSlashingPeriodValue{
		EndHeight:    slashingPeriod.EndHeight,
		SlashedSoFar: slashingPeriod.SlashedSoFar,
	}
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(slashingPeriodValue)
	store.Set(GetValidatorSlashingPeriodKey(slashingPeriod.ValidatorAddr, slashingPeriod.StartHeight), bz)
}

// Unmarshal key/value into a ValidatorSlashingPeriod
func (k Keeper) unmarshalSlashingPeriodKeyValue(key []byte, value []byte) ValidatorSlashingPeriod {
	var slashingPeriodValue ValidatorSlashingPeriodValue
	k.cdc.MustUnmarshalBinary(value, &slashingPeriodValue)
	address := sdk.ValAddress(key[1 : 1+sdk.AddrLen])
	startHeight := int64(binary.LittleEndian.Uint64(key[1+sdk.AddrLen : 1+sdk.AddrLen+8]))
	return ValidatorSlashingPeriod{
		ValidatorAddr: address,
		StartHeight:   startHeight,
		EndHeight:     slashingPeriodValue.EndHeight,
		SlashedSoFar:  slashingPeriodValue.SlashedSoFar,
	}
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
	ValidatorAddr sdk.ValAddress `json:"validator"`      // validator which this slashing period is for
	StartHeight   int64          `json:"start_height"`   // starting height of the slashing period
	EndHeight     int64          `json:"end_height"`     // ending height of the slashing period, or sentinel value of 0 for in-progress
	SlashedSoFar  sdk.Dec        `json:"slashed_so_far"` // fraction of validator stake slashed so far in this slashing period
}

// Value part of slashing period (validator address & start height are stored in the key)
type ValidatorSlashingPeriodValue struct {
	EndHeight    int64   `json:"end_height"`
	SlashedSoFar sdk.Dec `json:"slashed_so_far"`
}

// Return human readable slashing period
func (p ValidatorSlashingPeriod) HumanReadableString() string {
	return fmt.Sprintf("Start height: %d, end height: %d, slashed so far: %v",
		p.StartHeight, p.EndHeight, p.SlashedSoFar)
}
