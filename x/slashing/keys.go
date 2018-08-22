package slashing

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ValidatorSigningInfoKey     = []byte{0x01}
	ValidatorSigningBitArrayKey = []byte{0x02}
	ValidatorSlashingPeriodKey  = []byte{0x03}
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

// Stored by *validator* address (not owner address) followed by start height
func GetValidatorSlashingPeriodKey(v sdk.ValAddress, startHeight int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(startHeight))
	return append([]byte{0x03}, append(v.Bytes(), b...)...)
}
