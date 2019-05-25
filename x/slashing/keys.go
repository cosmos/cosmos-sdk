package slashing

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "slashing"

	// StoreKey is the store key string for slashing
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute is the querier route for slashing
	QuerierRoute = ModuleName
)

// key prefix bytes
var (
	ValidatorSigningInfoKey         = []byte{0x01} // Prefix for signing info
	ValidatorMissedBlockBitArrayKey = []byte{0x02} // Prefix for missed block bit array
	ValidatorSlashingPeriodKey      = []byte{0x03} // Prefix for slashing period
	AddrPubkeyRelationKey           = []byte{0x04} // Prefix for address-pubkey relation
)

// stored by *Tendermint* address (not operator address)
func GetValidatorSigningInfoKey(v sdk.ConsAddress) []byte {
	return append(ValidatorSigningInfoKey, v.Bytes()...)
}

// extract the address from a validator signing info key
func GetValidatorSigningInfoAddress(key []byte) (v sdk.ConsAddress) {
	addr := key[1:]
	if len(addr) != sdk.AddrLen {
		panic("unexpected key length")
	}
	return sdk.ConsAddress(addr)
}

// stored by *Tendermint* address (not operator address)
func GetValidatorMissedBlockBitArrayPrefixKey(v sdk.ConsAddress) []byte {
	return append(ValidatorMissedBlockBitArrayKey, v.Bytes()...)
}

// stored by *Tendermint* address (not operator address)
func GetValidatorMissedBlockBitArrayKey(v sdk.ConsAddress, i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return append(GetValidatorMissedBlockBitArrayPrefixKey(v), b...)
}

// stored by *Tendermint* address (not operator address)
func GetValidatorSlashingPeriodPrefix(v sdk.ConsAddress) []byte {
	return append(ValidatorSlashingPeriodKey, v.Bytes()...)
}

// stored by *Tendermint* address (not operator address) followed by start height
func GetValidatorSlashingPeriodKey(v sdk.ConsAddress, startHeight int64) []byte {
	b := make([]byte, 8)
	// this needs to be height + ValidatorUpdateDelay because the slashing period for genesis validators starts at height -ValidatorUpdateDelay
	binary.BigEndian.PutUint64(b, uint64(startHeight+sdk.ValidatorUpdateDelay))
	return append(GetValidatorSlashingPeriodPrefix(v), b...)
}

func getAddrPubkeyRelationKey(address []byte) []byte {
	return append(AddrPubkeyRelationKey, address...)
}
