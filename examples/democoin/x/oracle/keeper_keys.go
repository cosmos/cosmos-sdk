package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetInfoKey returns the key for OracleInfo
func GetInfoKey(p Payload, cdc sdk.Codec) []byte {
	bz := cdc.MustMarshalBinary(p)
	return append([]byte{0x00}, bz...)
}

// GetSignPrefix returns the prefix for signs
func GetSignPrefix(p Payload, cdc sdk.Codec) []byte {
	bz := cdc.MustMarshalBinary(p)
	return append([]byte{0x01}, bz...)
}

// GetSignKey returns the key for sign
func GetSignKey(p Payload, signer sdk.AccAddress, cdc sdk.Codec) []byte {
	return append(GetSignPrefix(p, cdc), signer...)
}
