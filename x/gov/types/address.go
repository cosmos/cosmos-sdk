package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/stretchr/testify/assert/yaml"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Address = GovernorAddress{}

const (
	// Prefix for governor addresses
	// Full prefix is defined as `bech32AccountPrefix + PrefixGovernor`
	PrefixGovernor = "gov"
)

type GovernorAddress []byte

// GovernorAddressFromHex creates a GovernorAddress from a hex string.
func GovernorAddressFromHex(address string) (addr GovernorAddress, err error) {
	bz, err := addressBytesFromHexString(address)
	return GovernorAddress(bz), err
}

// GovernorAddressFromBech32 creates a GovernorAddress from a Bech32 string.
func GovernorAddressFromBech32(address string) (addr GovernorAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return GovernorAddress{}, errors.New("empty address string is not allowed")
	}

	bech32PrefixGovAddr := sdk.GetConfig().GetBech32AccountAddrPrefix() + PrefixGovernor

	bz, err := sdk.GetFromBech32(address, bech32PrefixGovAddr)
	if err != nil {
		return nil, err
	}

	err = sdk.VerifyAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return GovernorAddress(bz), nil
}

// MustGovernorAddressFromBech32 creates a GovernorAddress from a Bech32 string.
// If the address is invalid, it panics.
func MustGovernorAddressFromBech32(address string) GovernorAddress {
	addr, err := GovernorAddressFromBech32(address)
	if err != nil {
		panic(err)
	}

	return addr
}

// Returns boolean for whether two GovernorAddresses are Equal
func (ga GovernorAddress) Equals(ga2 sdk.Address) bool {
	if ga.Empty() && ga2.Empty() {
		return true
	}

	return bytes.Equal(ga.Bytes(), ga2.Bytes())
}

// Returns boolean for whether a GovernorAddress is empty
func (ga GovernorAddress) Empty() bool {
	return len(ga) == 0
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (ga GovernorAddress) Marshal() ([]byte, error) {
	return ga, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (ga *GovernorAddress) Unmarshal(data []byte) error {
	*ga = data
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (ga GovernorAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(ga.String())
}

// MarshalYAML marshals to YAML using Bech32.
func (ga GovernorAddress) MarshalYAML() (interface{}, error) {
	return ga.String(), nil
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (ga *GovernorAddress) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		*ga = GovernorAddress{}
		return nil
	}

	ga2, err := GovernorAddressFromBech32(s)
	if err != nil {
		return err
	}

	*ga = ga2
	return nil
}

// UnmarshalYAML unmarshals from YAML assuming Bech32 encoding.
func (ga *GovernorAddress) UnmarshalYAML(data []byte) error {
	var s string
	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		*ga = GovernorAddress{}
		return nil
	}

	ga2, err := GovernorAddressFromBech32(s)
	if err != nil {
		return err
	}

	*ga = ga2
	return nil
}

// Bytes returns the raw address bytes.
func (ga GovernorAddress) Bytes() []byte {
	return ga
}

// String implements the Stringer interface.
func (ga GovernorAddress) String() string {
	if ga.Empty() {
		return ""
	}

	bech32Addr, err := sdk.Bech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix()+PrefixGovernor, ga.Bytes())
	if err != nil {
		panic(err)
	}
	return bech32Addr
}

// Format implements the fmt.Formatter interface.
func (ga GovernorAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		_, err := s.Write([]byte(ga.String()))
		if err != nil {
			panic(err)
		}
	case 'p':
		_, err := s.Write([]byte(fmt.Sprintf("%p", ga)))
		if err != nil {
			panic(err)
		}
	default:
		_, err := s.Write([]byte(fmt.Sprintf("%X", []byte(ga))))
		if err != nil {
			panic(err)
		}
	}
}

func addressBytesFromHexString(address string) ([]byte, error) {
	if len(address) == 0 {
		return nil, sdk.ErrEmptyHexAddress
	}

	return hex.DecodeString(address)
}

type governorAddressKey struct {
	stringDecoder func(string) (GovernorAddress, error)
	keyType       string
}

func (a governorAddressKey) Encode(buffer []byte, key GovernorAddress) (int, error) {
	return collections.BytesKey.Encode(buffer, key)
}

func (a governorAddressKey) Decode(buffer []byte) (int, GovernorAddress, error) {
	return collections.BytesKey.Decode(buffer)
}

func (a governorAddressKey) Size(key GovernorAddress) int {
	return collections.BytesKey.Size(key)
}

func (a governorAddressKey) EncodeJSON(value GovernorAddress) ([]byte, error) {
	return collections.StringKey.EncodeJSON(value.String())
}

func (a governorAddressKey) DecodeJSON(b []byte) (v GovernorAddress, err error) {
	s, err := collections.StringKey.DecodeJSON(b)
	if err != nil {
		return
	}
	v, err = a.stringDecoder(s)
	return
}

func (a governorAddressKey) Stringify(key GovernorAddress) string {
	return key.String()
}

func (a governorAddressKey) KeyType() string {
	return a.keyType
}

func (a governorAddressKey) EncodeNonTerminal(buffer []byte, key GovernorAddress) (int, error) {
	return collections.BytesKey.EncodeNonTerminal(buffer, key)
}

func (a governorAddressKey) DecodeNonTerminal(buffer []byte) (int, GovernorAddress, error) {
	return collections.BytesKey.DecodeNonTerminal(buffer)
}

func (a governorAddressKey) SizeNonTerminal(key GovernorAddress) int {
	return collections.BytesKey.SizeNonTerminal(key)
}
