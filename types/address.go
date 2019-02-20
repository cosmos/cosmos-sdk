package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/encoding/amino"

	"github.com/tendermint/tendermint/libs/bech32"
)

const (
	// AddrLen defines a valid address length
	AddrLen = 20
	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	Bech32MainPrefix = "cosmos"

	// PrefixAccount is the prefix for account keys
	PrefixAccount = "acc"
	// PrefixValidator is the prefix for validator keys
	PrefixValidator = "val"
	// PrefixConsensus is the prefix for consensus keys
	PrefixConsensus = "cons"
	// PrefixPublic is the prefix for public keys
	PrefixPublic = "pub"
	// PrefixOperator is the prefix for operator keys
	PrefixOperator = "oper"

	// PrefixAddress is the prefix for addresses
	PrefixAddress = "addr"

	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	Bech32PrefixAccAddr = Bech32MainPrefix
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	Bech32PrefixAccPub = Bech32MainPrefix + PrefixPublic
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = Bech32MainPrefix + PrefixValidator + PrefixOperator
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = Bech32MainPrefix + PrefixValidator + PrefixOperator + PrefixPublic
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = Bech32MainPrefix + PrefixValidator + PrefixConsensus
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = Bech32MainPrefix + PrefixValidator + PrefixConsensus + PrefixPublic
)

// Address is a common interface for different types of addresses used by the SDK
type Address interface {
	Equals(Address) bool
	Empty() bool
	Marshal() ([]byte, error)
	MarshalJSON() ([]byte, error)
	Bytes() []byte
	String() string
	Format(s fmt.State, verb rune)
}

// Ensure that different address types implement the interface
var _ Address = AccAddress{}
var _ Address = ValAddress{}
var _ Address = ConsAddress{}

// ----------------------------------------------------------------------------
// account
// ----------------------------------------------------------------------------

// AccAddress a wrapper around bytes meant to represent an account address.
// When marshaled to a string or JSON, it uses Bech32.
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
	if len(strings.TrimSpace(address)) == 0 {
		return AccAddress{}, nil
	}

	bech32PrefixAccAddr := GetConfig().GetBech32AccountAddrPrefix()

	bz, err := GetFromBech32(address, bech32PrefixAccAddr)
	if err != nil {
		return nil, err
	}

	if len(bz) != AddrLen {
		return nil, errors.New("Incorrect address length")
	}

	return AccAddress(bz), nil
}

// Returns boolean for whether two AccAddresses are Equal
func (aa AccAddress) Equals(aa2 Address) bool {
	if aa.Empty() && aa2.Empty() {
		return true
	}

	return bytes.Compare(aa.Bytes(), aa2.Bytes()) == 0
}

// Returns boolean for whether an AccAddress is empty
func (aa AccAddress) Empty() bool {
	if aa == nil {
		return true
	}

	aa2 := AccAddress{}
	return bytes.Compare(aa.Bytes(), aa2.Bytes()) == 0
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (aa AccAddress) Marshal() ([]byte, error) {
	return aa, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (aa *AccAddress) Unmarshal(data []byte) error {
	*aa = data
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (aa AccAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(aa.String())
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (aa *AccAddress) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	aa2, err := AccAddressFromBech32(s)
	if err != nil {
		return err
	}

	*aa = aa2
	return nil
}

// Bytes returns the raw address bytes.
func (aa AccAddress) Bytes() []byte {
	return aa
}

// String implements the Stringer interface.
func (aa AccAddress) String() string {
	if aa.Empty() {
		return ""
	}

	bech32PrefixAccAddr := GetConfig().GetBech32AccountAddrPrefix()

	bech32Addr, err := bech32.ConvertAndEncode(bech32PrefixAccAddr, aa.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Addr
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (aa AccAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", aa.String())))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", aa)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(aa))))
	}
}

// ----------------------------------------------------------------------------
// validator operator
// ----------------------------------------------------------------------------

// ValAddress defines a wrapper around bytes meant to present a validator's
// operator. When marshaled to a string or JSON, it uses Bech32.
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
	if len(strings.TrimSpace(address)) == 0 {
		return ValAddress{}, nil
	}

	bech32PrefixValAddr := GetConfig().GetBech32ValidatorAddrPrefix()

	bz, err := GetFromBech32(address, bech32PrefixValAddr)
	if err != nil {
		return nil, err
	}

	if len(bz) != AddrLen {
		return nil, errors.New("Incorrect address length")
	}

	return ValAddress(bz), nil
}

// Returns boolean for whether two ValAddresses are Equal
func (va ValAddress) Equals(va2 Address) bool {
	if va.Empty() && va2.Empty() {
		return true
	}

	return bytes.Compare(va.Bytes(), va2.Bytes()) == 0
}

// Returns boolean for whether an AccAddress is empty
func (va ValAddress) Empty() bool {
	if va == nil {
		return true
	}

	va2 := ValAddress{}
	return bytes.Compare(va.Bytes(), va2.Bytes()) == 0
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (va ValAddress) Marshal() ([]byte, error) {
	return va, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (va *ValAddress) Unmarshal(data []byte) error {
	*va = data
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (va ValAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(va.String())
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (va *ValAddress) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	va2, err := ValAddressFromBech32(s)
	if err != nil {
		return err
	}

	*va = va2
	return nil
}

// Bytes returns the raw address bytes.
func (va ValAddress) Bytes() []byte {
	return va
}

// String implements the Stringer interface.
func (va ValAddress) String() string {
	if va.Empty() {
		return ""
	}

	bech32PrefixValAddr := GetConfig().GetBech32ValidatorAddrPrefix()

	bech32Addr, err := bech32.ConvertAndEncode(bech32PrefixValAddr, va.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Addr
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (va ValAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", va.String())))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", va)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(va))))
	}
}

// ----------------------------------------------------------------------------
// consensus node
// ----------------------------------------------------------------------------

// ConsAddress defines a wrapper around bytes meant to present a consensus node.
// When marshaled to a string or JSON, it uses Bech32.
type ConsAddress []byte

// ConsAddressFromHex creates a ConsAddress from a hex string.
func ConsAddressFromHex(address string) (addr ConsAddress, err error) {
	if len(address) == 0 {
		return addr, errors.New("decoding Bech32 address failed: must provide an address")
	}

	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return ConsAddress(bz), nil
}

// ConsAddressFromBech32 creates a ConsAddress from a Bech32 string.
func ConsAddressFromBech32(address string) (addr ConsAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return ConsAddress{}, nil
	}

	bech32PrefixConsAddr := GetConfig().GetBech32ConsensusAddrPrefix()

	bz, err := GetFromBech32(address, bech32PrefixConsAddr)
	if err != nil {
		return nil, err
	}

	if len(bz) != AddrLen {
		return nil, errors.New("Incorrect address length")
	}

	return ConsAddress(bz), nil
}

// get ConsAddress from pubkey
func GetConsAddress(pubkey ConsPubKey) ConsAddress {
	return ConsAddress(pubkey.Address())
}

// Returns boolean for whether two ConsAddress are Equal
func (ca ConsAddress) Equals(ca2 Address) bool {
	if ca.Empty() && ca2.Empty() {
		return true
	}

	return bytes.Compare(ca.Bytes(), ca2.Bytes()) == 0
}

// Returns boolean for whether an ConsAddress is empty
func (ca ConsAddress) Empty() bool {
	if ca == nil {
		return true
	}

	ca2 := ConsAddress{}
	return bytes.Compare(ca.Bytes(), ca2.Bytes()) == 0
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (ca ConsAddress) Marshal() ([]byte, error) {
	return ca, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (ca *ConsAddress) Unmarshal(data []byte) error {
	*ca = data
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (ca ConsAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(ca.String())
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (ca *ConsAddress) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	ca2, err := ConsAddressFromBech32(s)
	if err != nil {
		return err
	}

	*ca = ca2
	return nil
}

// Bytes returns the raw address bytes.
func (ca ConsAddress) Bytes() []byte {
	return ca
}

// String implements the Stringer interface.
func (ca ConsAddress) String() string {
	if ca.Empty() {
		return ""
	}

	bech32PrefixConsAddr := GetConfig().GetBech32ConsensusAddrPrefix()

	bech32Addr, err := bech32.ConvertAndEncode(bech32PrefixConsAddr, ca.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Addr
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (ca ConsAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", ca.String())))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", ca)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", []byte(ca))))
	}
}

// ----------------------------------------------------------------------------
// auxiliary
// ----------------------------------------------------------------------------

// AccPubKey wrapper type around crypto.PubKey
type AccPubKey struct {
	crypto.PubKey
}

func NewEmptyAccPubKey() AccPubKey {
	return AccPubKey{}
}

// AccPubKeyFromHex creates an AccPubKey from a crypto.PubKey.
func AccPubKeyFromCryptoPubKey(cryptoPubKey crypto.PubKey) (pubKey AccPubKey) {
	return AccPubKey{cryptoPubKey}
}

// AccPubKeyFromHex creates an AccPubKey from a crypto.PubKey.
func (apk AccPubKey) CryptoPubKey() crypto.PubKey {
	return apk.PubKey
}

// AccPubKeyFromHex creates an AccPubKey from a hex string.
func AccPubKeyFromHex(hexPubKey string) (pubKey AccPubKey, err error) {
	if len(hexPubKey) == 0 {
		return pubKey, errors.New("decoding hexPubKey failed: must provide a pubkey")
	}

	bz, err := hex.DecodeString(hexPubKey)
	if err != nil {
		return pubKey, err
	}

	pk, err := cryptoAmino.PubKeyFromBytes(bz)
	if err != nil {
		return pubKey, err
	}

	return AccPubKeyFromCryptoPubKey(pk), nil
}

// AccPubKeyFromBech32 creates an AccPubKey from a Bech32 string.
func AccPubKeyFromBech32(bechPubKey string) (pubKey AccPubKey, err error) {
	if len(strings.TrimSpace(bechPubKey)) == 0 {
		return pubKey, nil
	}

	bech32PrefixAccPub := GetConfig().GetBech32AccountPubPrefix()

	bz, err := GetFromBech32(bechPubKey, bech32PrefixAccPub)
	if err != nil {
		return pubKey, err
	}

	pk, err := cryptoAmino.PubKeyFromBytes(bz)
	if err != nil {
		return pubKey, err
	}

	return AccPubKeyFromCryptoPubKey(pk), nil
}

// Returns boolean for whether two AccPubKey are Equal
func (apk AccPubKey) Equals(apk2 AccPubKey) bool {
	return (apk.Empty() && apk2.Empty()) ||
		apk.CryptoPubKey().Equals(apk2.CryptoPubKey())
}

// Returns boolean for whether an AccPubKey is empty
func (apk AccPubKey) Empty() bool {
	return apk == (NewEmptyAccPubKey()) ||
		apk.CryptoPubKey() == nil ||
		len(apk.Bytes()) == 0
}

// Marshal returns the raw pubkey bytes. It is needed for protobuf
// compatibility.
func (apk AccPubKey) Marshal() ([]byte, error) {
	return apk.Bytes(), nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (apk *AccPubKey) Unmarshal(data []byte) error {
	pk, err := cryptoAmino.PubKeyFromBytes(data)
	if err != nil {
		return err
	}
	*apk = AccPubKeyFromCryptoPubKey(pk)
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (apk AccPubKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(apk.String())
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (apk *AccPubKey) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	apk2, err := AccPubKeyFromBech32(s)
	if err != nil {
		return err
	}

	*apk = apk2
	return nil
}

// String implements the Stringer interface.
func (apk AccPubKey) String() string {
	if apk.Empty() {
		return ""
	}

	bech32PrefixAccPub := GetConfig().GetBech32AccountPubPrefix()

	bech32Pub, err := bech32.ConvertAndEncode(bech32PrefixAccPub, apk.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Pub
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (apk AccPubKey) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", apk.String())))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", apk.Bytes())))
	default:
		s.Write([]byte(fmt.Sprintf("%X", apk.Bytes())))
	}
}

// ValPubKey wrapper type around crypto.PubKey
type ValPubKey struct {
	crypto.PubKey
}

// Returns a new empty ValPubKey
func NewEmptyValPubKey() ValPubKey {
	return ValPubKey{}
}

// ValPubKeyFromCryptoPubKey creates an ValPubKey from a crypto.PubKey.
func ValPubKeyFromCryptoPubKey(cryptoPubKey crypto.PubKey) ValPubKey {
	return ValPubKey{cryptoPubKey}
}

// ValPubKeyFromHex creates an ValPubKey from a crypto.PubKey.
func (vpk ValPubKey) CryptoPubKey() crypto.PubKey {
	return vpk.PubKey
}

// AccPubKeyFromHex creates an AccPubKey from a hex string.
func ValPubKeyFromHex(hexPubKey string) (pubKey ValPubKey, err error) {
	if len(hexPubKey) == 0 {
		return pubKey, errors.New("decoding hexPubKey failed: must provide a pubkey")
	}

	bz, err := hex.DecodeString(hexPubKey)
	if err != nil {
		return pubKey, err
	}

	pk, err := cryptoAmino.PubKeyFromBytes(bz)
	if err != nil {
		return pubKey, err
	}

	return ValPubKeyFromCryptoPubKey(pk), nil
}

// AccPubKeyFromBech32 creates an AccPubKey from a Bech32 string.
func ValPubKeyFromBech32(bechPubKey string) (pubKey ValPubKey, err error) {
	if len(strings.TrimSpace(bechPubKey)) == 0 {
		return pubKey, nil
	}

	bech32ValidatorPubPrefix := GetConfig().GetBech32ValidatorPubPrefix()

	bz, err := GetFromBech32(bechPubKey, bech32ValidatorPubPrefix)
	if err != nil {
		return pubKey, err
	}

	pk, err := cryptoAmino.PubKeyFromBytes(bz)
	if err != nil {
		return pubKey, err
	}

	return ValPubKeyFromCryptoPubKey(pk), nil
}

// Returns boolean for whether two AccPubKey are Equal
func (vpk ValPubKey) Equals(vpk2 ValPubKey) bool {
	return (vpk.Empty() && vpk2.Empty()) ||
		vpk.CryptoPubKey().Equals(vpk2.CryptoPubKey())
}

// Returns boolean for whether an AccPubKey is empty
func (vpk ValPubKey) Empty() bool {
	return vpk == (NewEmptyValPubKey()) ||
		vpk.CryptoPubKey() == nil ||
		len(vpk.Bytes()) == 0
}

// Marshal returns the raw pubkey bytes. It is needed for protobuf
// compatibility.
func (vpk ValPubKey) Marshal() ([]byte, error) {
	return vpk.Bytes(), nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (vpk *ValPubKey) Unmarshal(data []byte) error {
	pk, err := cryptoAmino.PubKeyFromBytes(data)
	if err != nil {
		return err
	}
	*vpk = ValPubKeyFromCryptoPubKey(pk)
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (vpk ValPubKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(vpk.String())
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (vpk *ValPubKey) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	vpk2, err := ValPubKeyFromBech32(s)
	if err != nil {
		return err
	}

	*vpk = vpk2
	return nil
}

// String implements the Stringer interface.
func (vpk ValPubKey) String() string {
	if vpk.Empty() {
		return ""
	}

	bech32PrefixValPub := GetConfig().GetBech32ValidatorPubPrefix()

	bech32Pub, err := bech32.ConvertAndEncode(bech32PrefixValPub, vpk.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Pub
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (vpk ValPubKey) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", vpk.String())))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", vpk.Bytes())))
	default:
		s.Write([]byte(fmt.Sprintf("%X", vpk.Bytes())))
	}
}

// ConsPubKey wrapper type around crypto.PubKey
type ConsPubKey struct {
	crypto.PubKey
}

// Returns a new empty ConsPubKey
func NewEmptyConsPubKey() ConsPubKey {
	return ConsPubKey{}
}

// ConsPubKeyFromCryptoPubKey creates an ConsPubKey from a crypto.PubKey.
func ConsPubKeyFromCryptoPubKey(cryptoPubKey crypto.PubKey) ConsPubKey {
	return ConsPubKey{cryptoPubKey}
}

// ConsPubKeyFromHex returns the crypto.PubKey wrapped by a ConsPubKey.
func (cpk ConsPubKey) CryptoPubKey() crypto.PubKey {
	return cpk.PubKey
}

// ConsPubKeyFromHex creates an ConsPubKey from a hex string.
func ConsPubKeyFromHex(hexPubKey string) (pubKey ConsPubKey, err error) {
	if len(hexPubKey) == 0 {
		return pubKey, errors.New("decoding hexPubKey failed: must provide a pubkey")
	}

	bz, err := hex.DecodeString(hexPubKey)
	if err != nil {
		return pubKey, err
	}

	pk, err := cryptoAmino.PubKeyFromBytes(bz)
	if err != nil {
		return pubKey, err
	}

	return ConsPubKeyFromCryptoPubKey(pk), nil
}

// AccPubKeyFromBech32 creates an AccPubKey from a Bech32 string.
func ConsPubKeyFromBech32(bechPubKey string) (pubKey ConsPubKey, err error) {
	if len(strings.TrimSpace(bechPubKey)) == 0 {
		return pubKey, nil
	}

	bech32PrefixAccPub := GetConfig().GetBech32ConsensusPubPrefix()

	bz, err := GetFromBech32(bechPubKey, bech32PrefixAccPub)
	if err != nil {
		return pubKey, err
	}

	pk, err := cryptoAmino.PubKeyFromBytes(bz)
	if err != nil {
		return pubKey, err
	}

	return ConsPubKeyFromCryptoPubKey(pk), nil
}

// Returns boolean for whether two AccPubKey are Equal
func (cpk ConsPubKey) Equals(cpk2 ConsPubKey) bool {
	return (cpk.Empty() && cpk2.Empty()) ||
		cpk.CryptoPubKey().Equals(cpk2.CryptoPubKey())
}

// Returns boolean for whether an AccPubKey is empty
func (cpk ConsPubKey) Empty() bool {
	return cpk == (NewEmptyConsPubKey()) ||
		cpk.CryptoPubKey() == nil ||
		len(cpk.Bytes()) == 0
}

// Marshal returns the raw pubkey bytes. It is needed for protobuf
// compatibility.
func (cpk ConsPubKey) Marshal() ([]byte, error) {
	return cpk.Bytes(), nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (cpk *ConsPubKey) Unmarshal(data []byte) error {
	pk, err := cryptoAmino.PubKeyFromBytes(data)
	if err != nil {
		return err
	}
	*cpk = ConsPubKeyFromCryptoPubKey(pk)
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (cpk ConsPubKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(cpk.String())
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (cpk *ConsPubKey) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	cpk2, err := ConsPubKeyFromBech32(s)
	if err != nil {
		return err
	}

	*cpk = cpk2
	return nil
}

// String implements the Stringer interface.
func (cpk ConsPubKey) String() string {
	if cpk.Empty() {
		return ""
	}

	bech32PrefixConsensusPub := GetConfig().GetBech32ConsensusPubPrefix()

	bech32Pub, err := bech32.ConvertAndEncode(bech32PrefixConsensusPub, cpk.Bytes())
	if err != nil {
		panic(err)
	}

	return bech32Pub
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (cpk ConsPubKey) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(fmt.Sprintf("%s", cpk.String())))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", cpk.Bytes())))
	default:
		s.Write([]byte(fmt.Sprintf("%X", cpk.Bytes())))
	}
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
		return nil, fmt.Errorf("invalid Bech32 prefix; expected %s, got %s", prefix, hrp)
	}

	return bz, nil
}
