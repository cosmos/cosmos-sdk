package slashing

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// Prefix for signing info
	ValidatorSigningInfoKey = []byte{0x01}
	// Prefix for signature bit array
	ValidatorSigningBitArrayKey = []byte{0x02}
	// Prefix for slashing period
	ValidatorSlashingPeriodKey = []byte{0x03}
)

// Stored by *validator* address (not owner address)
func GetValidatorSigningInfoKey(v sdk.ValAddress) []byte {
	return append(ValidatorSigningInfoKey, v.Bytes()...)
}

// Stored by *validator* address (not owner address)
func GetValidatorSigningBitArrayKey(v sdk.ValAddress, i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return append(ValidatorSigningBitArrayKey, append(v.Bytes(), b...)...)
}

// Stored by *validator* address (not owner address)
func GetValidatorSlashingPeriodPrefix(v sdk.ValAddress) []byte {
	return append(ValidatorSlashingPeriodKey, v.Bytes()...)
}

// Stored by *validator* address (not owner address) followed by start height
func GetValidatorSlashingPeriodKey(v sdk.ValAddress, startHeight int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(startHeight))
	return append(GetValidatorSlashingPeriodPrefix(v), b...)
}
