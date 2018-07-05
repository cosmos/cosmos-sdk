package types

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/bech32"
)

// Bech32 prefixes
const (
	// expected address length
	AddrLen = 20

	// Bech32 prefixes
	Bech32PrefixAccAddr = "cosmosaccaddr"
	Bech32PrefixAccPub  = "cosmosaccpub"
	Bech32PrefixValAddr = "cosmosvaladdr"
	Bech32PrefixValPub  = "cosmosvalpub"
)

// Address is a go crypto-style Address
type Address []byte

func NewAddress(bech32String string) (Address, error) {
	return GetAccAddressHex(bech32String)
}

// Marshal needed for protobuf compatibility
func (bz Address) Marshal() ([]byte, error) {
	return bz, nil
}

// Unmarshal needed for protobuf compatibility
func (bz *Address) Unmarshal(data []byte) error {
	*bz = data
	return nil
}

// Marshals to JSON using Bech32
func (bz Address) MarshalJSON() ([]byte, error) {
	s := bz.String()
	jbz := make([]byte, len(s)+2)
	jbz[0] = '"'
	copy(jbz[1:], []byte(s))
	jbz[len(jbz)-1] = '"'
	return jbz, nil
}

// Unmarshals from JSON assuming Bech32 encoding
func (bz *Address) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("Invalid bech32 string: %s", data)
	}

	bz2, err := GetAccAddressHex(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*bz = bz2
	return nil
}

// Allow it to fulfill various interfaces in light-client, etc...
func (bz Address) Bytes() []byte {
	return bz
}

func (bz Address) String() string {
	bech32Addr, err := bech32.ConvertAndEncode(Bech32PrefixAccAddr, bz.Bytes())
	if err != nil {
		panic(err)
	}
	return bech32Addr
}

// For Printf / Sprintf, returns bech32 when using %s
func (bz Address) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", bz.String())))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", bz)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(bz))))
	}
}

// Bech32ifyAccPub takes AccountPubKey and returns the bech32 encoded string
func Bech32ifyAccPub(pub crypto.PubKey) (string, error) {
	return bech32.ConvertAndEncode(Bech32PrefixAccPub, pub.Bytes())
}

// MustBech32ifyAccPub panics on bech32-encoding failure
func MustBech32ifyAccPub(pub crypto.PubKey) string {
	enc, err := Bech32ifyAccPub(pub)
	if err != nil {
		panic(err)
	}
	return enc
}

// Bech32ifyVal returns the bech32 encoded string for a validator address
func Bech32ifyVal(bz []byte) (string, error) {
	return bech32.ConvertAndEncode(Bech32PrefixValAddr, bz)
}

// MustBech32ifyVal panics on bech32-encoding failure
func MustBech32ifyVal(bz []byte) string {
	enc, err := Bech32ifyVal(bz)
	if err != nil {
		panic(err)
	}
	return enc
}

// Bech32ifyValPub returns the bech32 encoded string for a validator pubkey
func Bech32ifyValPub(pub crypto.PubKey) (string, error) {
	return bech32.ConvertAndEncode(Bech32PrefixValPub, pub.Bytes())
}

// MustBech32ifyValPub pancis on bech32-encoding failure
func MustBech32ifyValPub(pub crypto.PubKey) string {
	enc, err := Bech32ifyValPub(pub)
	if err != nil {
		panic(err)
	}
	return enc
}

// create an Address from a string
func GetAccAddressHex(address string) (addr Address, err error) {
	if len(address) == 0 {
		return addr, errors.New("decoding bech32 address failed: must provide an address")
	}
	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

// create an Address from a string
func GetAccAddressBech32(address string) (addr Address, err error) {
	bz, err := GetFromBech32(address, Bech32PrefixAccAddr)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

// create an Address from a string, panics on error
func MustGetAccAddressBech32(address string) (addr Address) {
	addr, err := GetAccAddressBech32(address)
	if err != nil {
		panic(err)
	}
	return addr
}

// create a Pubkey from a string
func GetAccPubKeyBech32(address string) (pk crypto.PubKey, err error) {
	bz, err := GetFromBech32(address, Bech32PrefixAccPub)
	if err != nil {
		return nil, err
	}

	pk, err = crypto.PubKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// create an Pubkey from a string, panics on error
func MustGetAccPubKeyBech32(address string) (pk crypto.PubKey) {
	pk, err := GetAccPubKeyBech32(address)
	if err != nil {
		panic(err)
	}
	return pk
}

// create an Address from a hex string
func GetValAddressHex(address string) (addr Address, err error) {
	if len(address) == 0 {
		return addr, errors.New("decoding bech32 address failed: must provide an address")
	}
	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

// create an Address from a bech32 string
func GetValAddressBech32(address string) (addr Address, err error) {
	bz, err := GetFromBech32(address, Bech32PrefixValAddr)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

// create an Address from a string, panics on error
func MustGetValAddressBech32(address string) (addr Address) {
	addr, err := GetValAddressBech32(address)
	if err != nil {
		panic(err)
	}
	return addr
}

// decode a validator public key into a PubKey
func GetValPubKeyBech32(pubkey string) (pk crypto.PubKey, err error) {
	bz, err := GetFromBech32(pubkey, Bech32PrefixValPub)
	if err != nil {
		return nil, err
	}

	pk, err = crypto.PubKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// create an Pubkey from a string, panics on error
func MustGetValPubKeyBech32(address string) (pk crypto.PubKey) {
	pk, err := GetValPubKeyBech32(address)
	if err != nil {
		panic(err)
	}
	return pk
}

// decode a bytestring from a bech32-encoded string
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
