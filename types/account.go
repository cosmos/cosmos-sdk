package types

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/encoding/amino"

	"github.com/tendermint/tendermint/libs/bech32"
)

const (
	// AddrLen is the expected address length
	AddrLen = 20

	// Bech32PrefixAccAddr defines the bech32 prefix of an account's address
	Bech32PrefixAccAddr = "cosmosaccaddr"
	// Bech32PrefixAccPub defines the bech32 prefix of an account's public key
	Bech32PrefixAccPub = "cosmosaccpub"
	// Bech32PrefixValAddr defines the bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = "cosmosvaladdr"
	// Bech32PrefixValPub defines the bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = "cosmosvalpub"
	// Bech32PrefixTmAddr defines the bech32 prefix of a Tendermint node's address
	Bech32PrefixTmAddr = "cosmostmaddr"
	// Bech32PrefixTmPub defines the bech32 prefix of a Tendermint node's public key
	Bech32PrefixTmPub = "cosmostmpub"
)

// ----------------------------------------------------------------------------
// account
// ----------------------------------------------------------------------------

// AccAddress a wrapper around bytes meant to represent an account address
// When marshaled to a string or json, it uses bech32
type AccAddress []byte

// AccAddressFromHex createS an AccAddress from a hex string.
func AccAddressFromHex(address string) (addr AccAddress, err error) {
	if len(address) == 0 {
		return addr, errors.New("decoding bech32 address failed: must provide an address")
	}

	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return AccAddress(bz), nil
}

// AccAddressFromBech32 createS an AccAddress from a bech32 string.
func AccAddressFromBech32(address string) (addr AccAddress, err error) {
	bz, err := GetFromBech32(address, Bech32PrefixAccAddr)
	if err != nil {
		return nil, err
	}

	return AccAddress(bz), nil
}

// Marshal needed for protobuf compatibility.
func (bz AccAddress) Marshal() ([]byte, error) {
	return bz, nil
}

// Unmarshal needed for protobuf compatibility.
func (bz *AccAddress) Unmarshal(data []byte) error {
	*bz = data
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (bz AccAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(bz.String())
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (bz *AccAddress) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	bz2, err := AccAddressFromBech32(s)
	if err != nil {
		return err
	}

	*bz = bz2
	return nil
}

// Bytes returns the raw bytes.
func (bz AccAddress) Bytes() []byte {
	return bz
}

// String implements the Stringer interface.
func (bz AccAddress) String() string {
	bech32Addr, err := bech32.ConvertAndEncode(Bech32PrefixAccAddr, bz.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Addr
}

// Format implements the fmt.Formatter interface.
func (bz AccAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", bz.String())))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", bz)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(bz))))
	}
}

// ----------------------------------------------------------------------------
// validator owner
// ----------------------------------------------------------------------------

// AccAddress a wrapper around bytes meant to represent a validator address
// (from over ABCI).  When marshaled to a string or json, it uses bech32
type ValAddress []byte

// ValAddressFromHex creates a ValAddress from a hex string.
func ValAddressFromHex(address string) (addr ValAddress, err error) {
	if len(address) == 0 {
		return addr, errors.New("decoding bech32 address failed: must provide an address")
	}

	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return ValAddress(bz), nil
}

// ValAddressFromBech32 creates a ValAddress from a bech32 string.
func ValAddressFromBech32(address string) (addr ValAddress, err error) {
	bz, err := GetFromBech32(address, Bech32PrefixValAddr)
	if err != nil {
		return nil, err
	}

	return ValAddress(bz), nil
}

// Marshal needed for protobuf compatibility.
func (bz ValAddress) Marshal() ([]byte, error) {
	return bz, nil
}

// Unmarshal needed for protobuf compatibility.
func (bz *ValAddress) Unmarshal(data []byte) error {
	*bz = data
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (bz ValAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(bz.String())
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (bz *ValAddress) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	bz2, err := ValAddressFromBech32(s)
	if err != nil {
		return err
	}

	*bz = bz2
	return nil
}

// Bytes returns the raw bytes.
func (bz ValAddress) Bytes() []byte {
	return bz
}

// String implements the Stringer interface.
func (bz ValAddress) String() string {
	bech32Addr, err := bech32.ConvertAndEncode(Bech32PrefixValAddr, bz.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Addr
}

// Format implements the fmt.Formatter interface.
func (bz ValAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", bz.String())))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", bz)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(bz))))
	}
}

// ----------------------------------------------------------------------------
// auxiliary
// ----------------------------------------------------------------------------

// Bech32ifyAccPub takes AccountPubKey and returns the bech32 encoded string.
func Bech32ifyAccPub(pub crypto.PubKey) (string, error) {
	return bech32.ConvertAndEncode(Bech32PrefixAccPub, pub.Bytes())
}

// MustBech32ifyAccPub panics on bech32-encoding failure.
func MustBech32ifyAccPub(pub crypto.PubKey) string {
	enc, err := Bech32ifyAccPub(pub)
	if err != nil {
		panic(err)
	}

	return enc
}

// Bech32ifyValPub returns the bech32 encoded string for a validator's
// crypto.PubKey.
func Bech32ifyValPub(pub crypto.PubKey) (string, error) {
	return bech32.ConvertAndEncode(Bech32PrefixValPub, pub.Bytes())
}

// MustBech32ifyValPub panics on bech32-encoding failure.
func MustBech32ifyValPub(pub crypto.PubKey) string {
	enc, err := Bech32ifyValPub(pub)
	if err != nil {
		panic(err)
	}

	return enc
}

// GetAccPubKeyBech32 creates a crypto.PubKey from a string.
func GetAccPubKeyBech32(address string) (pk crypto.PubKey, err error) {
	bz, err := GetFromBech32(address, Bech32PrefixAccPub)
	if err != nil {
		return nil, err
	}

	pk, err = cryptoAmino.PubKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// MustGetAccPubKeyBech32 creates a crypto.PubKey from a string. It panics on
// error.
func MustGetAccPubKeyBech32(address string) (pk crypto.PubKey) {
	pk, err := GetAccPubKeyBech32(address)
	if err != nil {
		panic(err)
	}

	return pk
}

// GetValPubKeyBech32 decodes a validator public key into a crypto.PubKey.
func GetValPubKeyBech32(pubkey string) (pk crypto.PubKey, err error) {
	bz, err := GetFromBech32(pubkey, Bech32PrefixValPub)
	if err != nil {
		return nil, err
	}

	pk, err = cryptoAmino.PubKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// MustGetValPubKeyBech32 creates a crypto.PubKey from a string. It panics on
// error.
func MustGetValPubKeyBech32(address string) (pk crypto.PubKey) {
	pk, err := GetValPubKeyBech32(address)
	if err != nil {
		panic(err)
	}

	return pk
}

// GetFromBech32 decodes a bytestring from a bech32-encoded string.
func GetFromBech32(bech32str, prefix string) ([]byte, error) {
	if len(bech32str) == 0 {
		return nil, errors.New("decoding bech32 address failed: must provide an address")
	}

	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return nil, err
	}

	if hrp != prefix {
		return nil, fmt.Errorf("invalid bech32 prefix. Expected %s, Got %s", prefix, hrp)
	}

	return bz, nil
}
