package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/crypto"
)

func EncodePubKey(cdc codec.BinaryMarshaler, pubkey crypto.PubKey) []byte {
	if pubkey == nil {
		return []byte{}
	}
	bz, err := codec.MarshalAny(cdc, pubkey)
	if err != nil {
		panic(err)
	}
	return bz
}

func DecodePubKey(cdc codec.BinaryMarshaler, bz []byte) crypto.PubKey {
	if len(bz) == 0 {
		return nil
	}
	var pubkey crypto.PubKey
	err := codec.UnmarshalAny(cdc, &pubkey, bz)
	if err != nil {
		panic(err)
	}
	return pubkey
}

func EncodeHistoricalEntry(cdc codec.BinaryMarshaler, entry PubKeyHistory) []byte {
	return cdc.MustMarshalBinaryBare(&entry)
}

func DecodeHistoricalEntry(cdc codec.BinaryMarshaler, bz []byte) PubKeyHistory {
	var entry PubKeyHistory
	cdc.MustUnmarshalBinaryBare(bz, &entry)
	return entry
}
