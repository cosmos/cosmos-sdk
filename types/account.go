package types

import (
	"encoding/hex"
	"errors"
	"fmt"

	bech32cosmos "github.com/cosmos/bech32cosmos/go"
	crypto "github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
)

//Address is a go crypto-style Address
type Address = cmn.HexBytes

// Bech32 prefixes
const (
	Bech32PrefixAccAddr = "cosmosaccaddr"
	Bech32PrefixAccPub  = "cosmosaccpub"
	Bech32PrefixValAddr = "cosmosvaladdr"
	Bech32PrefixValPub  = "cosmosvalpub"
)

// Bech32CosmosifyAcc takes Address and returns the Bech32Cosmos encoded string
func Bech32CosmosifyAcc(addr Address) (string, error) {
	return bech32cosmos.ConvertAndEncode(Bech32PrefixAccAddr, addr.Bytes())
}

// Bech32CosmosifyAccPub takes AccountPubKey and returns the Bech32Cosmos encoded string
func Bech32CosmosifyAccPub(pub crypto.PubKey) (string, error) {
	return bech32cosmos.ConvertAndEncode(Bech32PrefixAccPub, pub.Bytes())
}

// Bech32CosmosifyVal returns the Bech32Cosmos encoded string for a validator address
func Bech32CosmosifyVal(addr Address) (string, error) {
	return bech32cosmos.ConvertAndEncode(Bech32PrefixValAddr, addr.Bytes())
}

// Bech32CosmosifyValPub returns the Bech32Cosmos encoded string for a validator pubkey
func Bech32CosmosifyValPub(pub crypto.PubKey) (string, error) {
	return bech32cosmos.ConvertAndEncode(Bech32PrefixValPub, pub.Bytes())
}

// create an Address from a string
func GetAccAddressHex(address string) (addr Address, err error) {
	if len(address) == 0 {
		return addr, errors.New("must use provide address")
	}
	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

// create an Address from a string
func GetAccAddressBech32Cosmos(address string) (addr Address, err error) {
	bz, err := getFromBech32Cosmos(address, Bech32PrefixAccAddr)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

// create an Address from a hex string
func GetValAddressHex(address string) (addr Address, err error) {
	if len(address) == 0 {
		return addr, errors.New("must use provide address")
	}
	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

// create an Address from a bech32cosmos string
func GetValAddressBech32Cosmos(address string) (addr Address, err error) {
	bz, err := getFromBech32Cosmos(address, Bech32PrefixValAddr)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

//Decode a validator publickey into a public key
func GetValPubKeyBech32Cosmos(pubkey string) (pk crypto.PubKey, err error) {
	bz, err := getFromBech32Cosmos(pubkey, Bech32PrefixValPub)
	if err != nil {
		return nil, err
	}

	pk, err = crypto.PubKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

func getFromBech32Cosmos(bech32, prefix string) ([]byte, error) {
	if len(bech32) == 0 {
		return nil, errors.New("must provide non-empty string")
	}
	hrp, bz, err := bech32cosmos.DecodeAndConvert(bech32)
	if err != nil {
		return nil, err
	}

	if hrp != prefix {
		return nil, fmt.Errorf("Invalid bech32 prefix. Expected %s, Got %s", prefix, hrp)
	}

	return bz, nil
}
