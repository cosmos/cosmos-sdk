package slashing

import (
	"bytes"
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Cap an infraction's slash amount by the slashing period in which it was committed
func (k Keeper) capBySlashingPeriod(ctx sdk.Context, address sdk.ConsAddress, fraction sdk.Dec, infractionHeight int64) (revisedFraction sdk.Dec) {

	// Fetch the newest slashing period starting before this infraction was committed
	slashingPeriod := k.getValidatorSlashingPeriodForHeight(ctx, address, infractionHeight)

	// Sanity check
	if slashingPeriod.EndHeight > 0 && slashingPeriod.EndHeight < infractionHeight {
		panic(fmt.Sprintf("slashing period ended before infraction: validator %s, infraction height %d, slashing period ended at %d", address, infractionHeight, slashingPeriod.EndHeight))
	}

	// Calculate the updated total slash amount
	// This is capped at the slashing fraction for the worst infraction within this slashing period
	totalToSlash := sdk.MaxDec(slashingPeriod.SlashedSoFar, fraction)

	// Calculate the remainder which we now must slash
	revisedFraction = totalToSlash.Sub(slashingPeriod.SlashedSoFar)

	// Update the slashing period struct
	slashingPeriod.SlashedSoFar = totalToSlash
	k.addOrUpdateValidatorSlashingPeriod(ctx, slashingPeriod)

	return
}

// Stored by validator Tendermint address (not operator address)
// This function retrieves the most recent slashing period starting
// before a particular height - so the slashing period that was "in effect"
// at the time of an infraction committed at that height.
func (k Keeper) getValidatorSlashingPeriodForHeight(ctx sdk.Context, address sdk.ConsAddress, height int64) (slashingPeriod ValidatorSlashingPeriod) {
	store := ctx.KVStore(k.storeKey)
	// Get the most recent slashing period at or before the infraction height
	start := GetValidatorSlashingPeriodPrefix(address)
	end := sdk.PrefixEndBytes(GetValidatorSlashingPeriodKey(address, height))
	fmt.Printf("start: %X, end: %X, diff: %v\n", start, end, bytes.Compare(start, end))
	// TODO
	itr := sdk.KVStorePrefixIterator(store, GetValidatorSlashingPeriodPrefix(address))
	for itr.Valid() {
		fmt.Printf("Key: %X\n", itr.Key())
		period := k.unmarshalSlashingPeriodKeyValue(itr.Key(), itr.Value())
		fmt.Printf("Found %X => %v\n", address, period)
		itr.Next()
	}
	// END TODO
	iterator := store.ReverseIterator(start, end)
	if !iterator.Valid() {
		panic(fmt.Sprintf("expected to find slashing period for validator %s before height %d, but none was found", address, height))
	}
	slashingPeriod = k.unmarshalSlashingPeriodKeyValue(iterator.Key(), iterator.Value())
	return
}

// Stored by validator Tendermint address (not operator address)
// This function sets a validator slashing period for a particular validator,
// start height, end height, and current slashed-so-far total, or updates
// an existing slashing period for the same validator and start height.
func (k Keeper) addOrUpdateValidatorSlashingPeriod(ctx sdk.Context, slashingPeriod ValidatorSlashingPeriod) {
	slashingPeriodValue := ValidatorSlashingPeriodValue{
		EndHeight:    slashingPeriod.EndHeight,
		SlashedSoFar: slashingPeriod.SlashedSoFar,
	}
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(slashingPeriodValue)
	fmt.Printf("Set slashing period for validator: %X => %s\n", GetValidatorSlashingPeriodKey(slashingPeriod.ValidatorAddr, slashingPeriod.StartHeight), slashingPeriod.ValidatorAddr)
	store.Set(GetValidatorSlashingPeriodKey(slashingPeriod.ValidatorAddr, slashingPeriod.StartHeight), bz)
}

// Unmarshal key/value into a ValidatorSlashingPeriod
func (k Keeper) unmarshalSlashingPeriodKeyValue(key []byte, value []byte) ValidatorSlashingPeriod {
	var slashingPeriodValue ValidatorSlashingPeriodValue
	k.cdc.MustUnmarshalBinary(value, &slashingPeriodValue)
	address := sdk.ConsAddress(key[1 : 1+sdk.AddrLen])
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
	ValidatorAddr sdk.ConsAddress `json:"validator_addr"` // validator which this slashing period is for
	StartHeight   int64           `json:"start_height"`   // starting height of the slashing period
	EndHeight     int64           `json:"end_height"`     // ending height of the slashing period, or sentinel value of 0 for in-progress
	SlashedSoFar  sdk.Dec         `json:"slashed_so_far"` // fraction of validator stake slashed so far in this slashing period
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
