package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// module name
	ModuleName = "slashing"

	// StoreKey is the store key string for slashing
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute is the querier route for slashing
	QuerierRoute = ModuleName
)

// Query endpoints supported by the slashing querier
const (
	QueryParameters   = "parameters"
	QuerySigningInfo  = "signingInfo"
	QuerySigningInfos = "signingInfos"
)

// Keys for slashing store
// Items are stored with the following key: values
//
// - 0x01<consAddress_Bytes>: ValidatorSigningInfo
//
// - 0x02<consAddress_Bytes><period_Bytes>: bool
//
// - 0x03<accAddr_Bytes>: crypto.PubKey
var (
	ValidatorSigningInfoKey         = []byte{0x01} // Prefix for signing info
	ValidatorMissedBlockBitArrayKey = []byte{0x02} // Prefix for missed block bit array
	AddrPubkeyRelationKey           = []byte{0x03} // Prefix for address-pubkey relation
)

// stored by *Consensus* address (not operator address)
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

// stored by *Consensus* address (not operator address)
func GetValidatorMissedBlockBitArrayPrefixKey(v sdk.ConsAddress) []byte {
	return append(ValidatorMissedBlockBitArrayKey, v.Bytes()...)
}

// stored by *Consensus* address (not operator address)
func GetValidatorMissedBlockBitArrayKey(v sdk.ConsAddress, i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return append(GetValidatorMissedBlockBitArrayPrefixKey(v), b...)
}

// get pubkey relation key used to get the pubkey from the address
func GetAddrPubkeyRelationKey(address []byte) []byte {
	return append(AddrPubkeyRelationKey, address...)
}
