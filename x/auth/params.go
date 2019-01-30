package auth

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// DefaultParamspace defines the default auth module parameter subspace
const DefaultParamspace = "auth"

// Default parameter values
const (
	DefaultMaxMemoCharacters      uint64 = 256
	DefaultTxSigLimit             uint64 = 7
	DefaultTxSizeCostPerByte      uint64 = 10
	DefaultSigVerifyCostED25519   uint64 = 590
	DefaultSigVerifyCostSecp256k1 uint64 = 1000
)

// Parameter keys
var (
	KeyMaxMemoCharacters      = []byte("MaxMemoCharacters")
	KeyTxSigLimit             = []byte("TxSigLimit")
	KeyTxSizeCostPerByte      = []byte("TxSizeCostPerByte")
	KeySigVerifyCostED25519   = []byte("SigVerifyCostED25519")
	KeySigVerifyCostSecp256k1 = []byte("SigVerifyCostSecp256k1")
)

var _ params.ParamSet = &Params{}

// Params defines the parameters for the auth module.
type Params struct {
	MaxMemoCharacters      uint64
	TxSigLimit             uint64 // max total number of signatures per tx
	TxSizeCostPerByte      uint64
	SigVerifyCostED25519   uint64
	SigVerifyCostSecp256k1 uint64
}

// ParamTable for staking module
func ParamTypeTable() params.TypeTable {
	return params.NewTypeTable().RegisterParamSet(&Params{})
}

// KeyValuePairs implements the ParamSet interface and returns all the key/value
// pairs of auth module's parameters.
// nolint
func (p *Params) KeyValuePairs() params.KeyValuePairs {
	return params.KeyValuePairs{
		{KeyMaxMemoCharacters, &p.MaxMemoCharacters},
		{KeyTxSigLimit, &p.TxSigLimit},
		{KeyTxSizeCostPerByte, &p.TxSizeCostPerByte},
		{KeySigVerifyCostED25519, &p.SigVerifyCostED25519},
		{KeySigVerifyCostSecp256k1, &p.SigVerifyCostSecp256k1},
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := msgCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := msgCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MaxMemoCharacters:      DefaultMaxMemoCharacters,
		TxSigLimit:             DefaultTxSigLimit,
		TxSizeCostPerByte:      DefaultTxSizeCostPerByte,
		SigVerifyCostED25519:   DefaultSigVerifyCostED25519,
		SigVerifyCostSecp256k1: DefaultSigVerifyCostSecp256k1,
	}
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder

	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("MaxMemoCharacters: %d\n", p.MaxMemoCharacters))
	sb.WriteString(fmt.Sprintf("TxSigLimit: %d\n", p.TxSigLimit))
	sb.WriteString(fmt.Sprintf("TxSizeCostPerByte: %d\n", p.TxSizeCostPerByte))
	sb.WriteString(fmt.Sprintf("SigVerifyCostED25519: %d\n", p.SigVerifyCostED25519))
	sb.WriteString(fmt.Sprintf("SigVerifyCostSecp256k1: %d\n", p.SigVerifyCostSecp256k1))

	return sb.String()
}

// MustUnmarshalParams deserializes raw Params bytes into a Params structure. It
// will panic upon failure.
func MustUnmarshalParams(cdc *codec.Codec, value []byte) Params {
	params, err := UnmarshalParams(cdc, value)
	if err != nil {
		panic(err)
	}

	return params
}

// UnmarshalParams deserializes raw Params bytes into a Params structure. It will
// return an error upon failure.
func UnmarshalParams(cdc *codec.Codec, value []byte) (params Params, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &params)
	if err != nil {
		return Params{}, err
	}

	return
}
