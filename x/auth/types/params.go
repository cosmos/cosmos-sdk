package types

import (
	"fmt"
	"github.com/gogo/protobuf/proto"

	yaml "gopkg.in/yaml.v2"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	pt "github.com/gogo/protobuf/types"
)

// Default parameter values
var (
	DefaultMaxMemoCharacters      = pt.UInt64Value{Value: 256}
	DefaultTxSigLimit             = pt.UInt64Value{Value: 7}
	DefaultTxSizeCostPerByte      = pt.UInt64Value{Value: 10}
	DefaultSigVerifyCostED25519   = pt.UInt64Value{Value: 590}
	DefaultSigVerifyCostSecp256k1 = pt.UInt64Value{Value: 1000}
)

// Parameter keys
var (
	KeyMaxMemoCharacters      = []byte("MaxMemoCharacters")
	KeyTxSigLimit             = []byte("TxSigLimit")
	KeyTxSizeCostPerByte      = []byte("TxSizeCostPerByte")
	KeySigVerifyCostED25519   = []byte("SigVerifyCostED25519")
	KeySigVerifyCostSecp256k1 = []byte("SigVerifyCostSecp256k1")
)

var _ paramtypes.ParamSet = &Params{}

// NewParams creates a new Params object
func NewParams(
	maxMemoCharacters, txSigLimit, txSizeCostPerByte, sigVerifyCostED25519, sigVerifyCostSecp256k1 pt.UInt64Value,
) Params {
	return Params{
		MaxMemoCharacters:      &maxMemoCharacters,
		TxSigLimit:             &txSigLimit,
		TxSizeCostPerByte:      &txSizeCostPerByte,
		SigVerifyCostED25519:   &sigVerifyCostED25519,
		SigVerifyCostSecp256k1: &sigVerifyCostSecp256k1,
	}
}

// ParamKeyTable for auth module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
// nolint
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {

	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxMemoCharacters, p.MaxMemoCharacters, validateMaxMemoCharacters),
		paramtypes.NewParamSetPair(KeyTxSigLimit, p.TxSigLimit, validateTxSigLimit),
		paramtypes.NewParamSetPair(KeyTxSizeCostPerByte, p.TxSizeCostPerByte, validateTxSizeCostPerByte),
		paramtypes.NewParamSetPair(KeySigVerifyCostED25519, p.SigVerifyCostED25519, validateSigVerifyCostED25519),
		paramtypes.NewParamSetPair(KeySigVerifyCostSecp256k1, p.SigVerifyCostSecp256k1, validateSigVerifyCostSecp256k1),
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MaxMemoCharacters:      &DefaultMaxMemoCharacters,
		TxSigLimit:             &DefaultTxSigLimit,
		TxSizeCostPerByte:      &DefaultTxSizeCostPerByte,
		SigVerifyCostED25519:   &DefaultSigVerifyCostED25519,
		SigVerifyCostSecp256k1: &DefaultSigVerifyCostSecp256k1,
	}
}

// String implements the stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateTxSigLimit(m proto.Message) error {
	v, ok := m.(*pt.UInt64Value)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", m)
	}

	if v.Value == 0 {
		return fmt.Errorf("invalid tx signature limit: %d", v)
	}

	return nil
}

func validateSigVerifyCostED25519(m proto.Message) error {
	v, ok := m.(*pt.UInt64Value)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", m)
	}

	if v.Value == 0 {
		return fmt.Errorf("invalid ED25519 signature verification cost: %d", v)
	}

	return nil
}

func validateSigVerifyCostSecp256k1(m proto.Message) error {
	v, ok := m.(*pt.UInt64Value)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", m)
	}

	if v.Value == 0 {
		return fmt.Errorf("invalid SECK256k1 signature verification cost: %d", v)
	}

	return nil
}

func validateMaxMemoCharacters(m proto.Message) error {
	v, ok := m.(*pt.UInt64Value)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", m)
	}

	if v.Value == 0 {
		return fmt.Errorf("invalid max memo characters: %d", v)
	}

	return nil
}

func validateTxSizeCostPerByte(m proto.Message) error {
	v, ok := m.(*pt.UInt64Value)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", m)
	}

	if v.Value == 0 {
		return fmt.Errorf("invalid tx size cost per byte: %d", v)
	}

	return nil
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateTxSigLimit(p.TxSigLimit); err != nil {
		return err
	}
	if err := validateSigVerifyCostED25519(p.SigVerifyCostED25519); err != nil {
		return err
	}
	if err := validateSigVerifyCostSecp256k1(p.SigVerifyCostSecp256k1); err != nil {
		return err
	}
	if err := validateMaxMemoCharacters(p.MaxMemoCharacters); err != nil {
		return err
	}
	if err := validateTxSizeCostPerByte(p.TxSizeCostPerByte); err != nil {
		return err
	}

	return nil
}
