package types

import (
	"encoding/json"

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

func EncodeHistoricalEntry(entry PubKeyHistory) []byte {
	// TODO: what steps are required to use MarshalAny for PubKeyHistory
	// bz, err := codec.MarshalAny(pk.cdc, entry)
	bz, err := json.Marshal(entry)
	if err != nil {
		panic(err)
	}
	return bz
}

func DecodeHistoricalEntry(bz []byte) PubKeyHistory {
	var entry PubKeyHistory
	// TODO: what steps are required to use UnmarshalAny for PubKeyHistory
	// err := codec.UnmarshalAny(pk.cdc, &entry, bz)
	err := json.Unmarshal(bz, &entry)
	if err != nil {
		panic(err)
	}
	return entry
}
