package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// GetInfoKey returns the key for OracleInfo
func GetInfoKey(p Payload, cdc *wire.Codec) []byte {
	bz := cdc.MustMarshalBinary(p)
	return append([]byte{0x00}, bz...)
}

// GetSignPrefix returns the prefix for signs
func GetSignPrefix(p Payload, cdc *wire.Codec) []byte {
	bz := cdc.MustMarshalBinary(p)
	return append([]byte{0x01}, bz...)
}

// GetSignKey returns the key for sign
func GetSignKey(p Payload, signer sdk.AccAddress, cdc *wire.Codec) []byte {
	return append(GetSignPrefix(p, cdc), signer...)
}
