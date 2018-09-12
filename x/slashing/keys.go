package slashing

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// key prefix bytes
var (
	ValidatorSigningInfoKey     = []byte{0x01} // Prefix for signing info
	ValidatorSigningBitArrayKey = []byte{0x02} // Prefix for signature bit array
	ValidatorSlashingPeriodKey  = []byte{0x03} // Prefix for slashing period
	AddrPubkeyRelationKey       = []byte{0x04} // Prefix for address-pubkey relation
)

// stored by *Tendermint* address (not operator address)
func GetValidatorSigningInfoKey(v sdk.ConsAddress) []byte {
	return append(ValidatorSigningInfoKey, v.Bytes()...)
}

// stored by *Tendermint* address (not operator address)
func GetValidatorSigningBitArrayKey(v sdk.ConsAddress, i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return append(ValidatorSigningBitArrayKey, append(v.Bytes(), b...)...)
}

// stored by *Tendermint* address (not operator address)
func GetValidatorSlashingPeriodPrefix(v sdk.ConsAddress) []byte {
	return append(ValidatorSlashingPeriodKey, v.Bytes()...)
}

// stored by *Tendermint* address (not operator address) followed by start height
func GetValidatorSlashingPeriodKey(v sdk.ConsAddress, startHeight int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(startHeight))
	return append(GetValidatorSlashingPeriodPrefix(v), b...)
}

func getAddrPubkeyRelationKey(address []byte) []byte {
	return append(AddrPubkeyRelationKey, address...)
}
