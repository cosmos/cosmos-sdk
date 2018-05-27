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

// Bech32CosmosifyAcc takes Address and returns the Bech32Cosmos encoded string
func Bech32CosmosifyAcc(addr Address) (string, error) {
	return bech32cosmos.ConvertAndEncode("cosmosaccaddr", addr.Bytes())
}

// Bech32CosmosifyAccPub takes AccountPubKey and returns the Bech32Cosmos encoded string
func Bech32CosmosifyAccPub(pub crypto.PubKey) (string, error) {
	return bech32cosmos.ConvertAndEncode("cosmosaccpub", pub.Bytes())
}

// Bech32CosmosifyVal returns the Bech32Cosmos encoded string for a validator address
func Bech32CosmosifyVal(addr Address) (string, error) {
	return bech32cosmos.ConvertAndEncode("cosmosvaladdr", addr.Bytes())
}

// Bech32CosmosifyValPub returns the Bech32Cosmos encoded string for a validator pubkey
func Bech32CosmosifyValPub(pub crypto.PubKey) (string, error) {
	return bech32cosmos.ConvertAndEncode("cosmosvalpub", pub.Bytes())
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
	if len(address) == 0 {
		return addr, errors.New("must use provide address")
	}

	hrp, bz, err := bech32cosmos.DecodeAndConvert(address)

	if hrp != "cosmosaccaddr" {
		return addr, fmt.Errorf("Invalid Address Prefix. Expected cosmosaccaddr, Got %s", hrp)
	}

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
	if len(address) == 0 {
		return addr, errors.New("must use provide address")
	}

	hrp, bz, err := bech32cosmos.DecodeAndConvert(address)

	if hrp != "cosmosvaladdr" {
		return addr, fmt.Errorf("Invalid Address Prefix. Expected cosmosvaladdr, Got %s", hrp)
	}

	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

//Decode a validator publickey into a public key
func GetValPubKeyBech32Cosmos(pubkey string) (pk crypto.PubKey, err error) {
	if len(pubkey) == 0 {
		return pk, errors.New("must use provide pubkey")
	}
	hrp, bz, err := bech32cosmos.DecodeAndConvert(pubkey)

	if hrp != "cosmosvalpub" {
		return pk, fmt.Errorf("Invalid Validator Pubkey Prefix. Expected cosmosvalpub, Got %s", hrp)
	}

	if err != nil {
		return nil, err
	}

	pk, err = crypto.PubKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}
