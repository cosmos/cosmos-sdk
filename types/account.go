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
	// AddrLen defines a valid address length
	AddrLen = 20

	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	Bech32PrefixAccAddr = "cosmosaccaddr"
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	Bech32PrefixAccPub = "cosmosaccpub"
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = "cosmosvaladdr"
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = "cosmosvalpub"
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = "cosmosconsaddr"
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = "cosmosconspub"
)

// ----------------------------------------------------------------------------
// account
// ----------------------------------------------------------------------------

// AccAddress a wrapper around bytes meant to represent an account address
// When marshaled to a string or json, it uses bech32
type AccAddress []byte

// AccAddressFromHex creates an AccAddress from a hex string.
func AccAddressFromHex(address string) (addr AccAddress, err error) {
	if len(address) == 0 {
		return addr, errors.New("decoding Bech32 address failed: must provide an address")
	}

	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return AccAddress(bz), nil
}

// AccAddressFromBech32 creates an AccAddress from a Bech32 string.
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
		return addr, errors.New("decoding Bech32 address failed: must provide an address")
	}

	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return ValAddress(bz), nil
}

// ValAddressFromBech32 creates a ValAddress from a Bech32 string.
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

// Bech32ifyAccPub returns a Bech32 encoded string for a given account PubKey.
func Bech32ifyAccPub(pub crypto.PubKey) (string, error) {
	return bech32.ConvertAndEncode(Bech32PrefixAccPub, pub.Bytes())
}

// MustBech32ifyAccPub returns the result of Bech32ifyAccPub panicing on failure.
func MustBech32ifyAccPub(pub crypto.PubKey) string {
	enc, err := Bech32ifyAccPub(pub)
	if err != nil {
		panic(err)
	}

	return enc
}

// Bech32ifyValPub returns a Bech32 encoded string for a given validator
// operator's PubKey.
func Bech32ifyValPub(pub crypto.PubKey) (string, error) {
	return bech32.ConvertAndEncode(Bech32PrefixValPub, pub.Bytes())
}

// MustBech32ifyValPub returns the result of Bech32ifyValPub panicing on failure.
func MustBech32ifyValPub(pub crypto.PubKey) string {
	enc, err := Bech32ifyValPub(pub)
	if err != nil {
		panic(err)
	}

	return enc
}

// Bech32ifyConsPub returns a Bech32 encoded string for a given consensus node's
// PubKey.
func Bech32ifyConsPub(pub crypto.PubKey) (string, error) {
	return bech32.ConvertAndEncode(Bech32PrefixConsPub, pub.Bytes())
}

// MustBech32ifyConsPub returns the result of Bech32ifyConsPub panicing on
// failure.
func MustBech32ifyConsPub(pub crypto.PubKey) string {
	enc, err := Bech32ifyConsPub(pub)
	if err != nil {
		panic(err)
	}

	return enc
}

// GetAccPubKeyBech32 creates a PubKey from a string.
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

// MustGetAccPubKeyBech32 creates a PubKey from a string. It panics on error.
func MustGetAccPubKeyBech32(address string) (pk crypto.PubKey) {
	pk, err := GetAccPubKeyBech32(address)
	if err != nil {
		panic(err)
	}

	return pk
}

// GetValPubKeyBech32 decodes a validator public key into a PubKey.
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

// MustGetValPubKeyBech32 creates a PubKey from a string. It panics on error.
func MustGetValPubKeyBech32(address string) (pk crypto.PubKey) {
	pk, err := GetValPubKeyBech32(address)
	if err != nil {
		panic(err)
	}

	return pk
}

// GetFromBech32 decodes a bytestring from a Bech32 encoded string.
func GetFromBech32(bech32str, prefix string) ([]byte, error) {
	if len(bech32str) == 0 {
		return nil, errors.New("decoding Bech32 address failed: must provide an address")
	}

	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return nil, err
	}

	if hrp != prefix {
		return nil, fmt.Errorf("invalid Bech32 prefix. Expected %s, Got %s", prefix, hrp)
	}

	return bz, nil
}
