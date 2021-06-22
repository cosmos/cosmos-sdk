package types

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

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

var _ paramtypes.ParamSet = &Params{}

// NewParams creates a new Params object
func NewParams(
	maxMemoCharacters, txSigLimit, txSizeCostPerByte, sigVerifyCostED25519, sigVerifyCostSecp256k1 uint64,
) Params {
	return Params{
		MaxMemoCharacters:      maxMemoCharacters,
		TxSigLimit:             txSigLimit,
		TxSizeCostPerByte:      txSizeCostPerByte,
		SigVerifyCostED25519:   sigVerifyCostED25519,
		SigVerifyCostSecp256k1: sigVerifyCostSecp256k1,
	}
}

// ParamKeyTable for auth module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxMemoCharacters, &p.MaxMemoCharacters, validateMaxMemoCharacters),
		paramtypes.NewParamSetPair(KeyTxSigLimit, &p.TxSigLimit, validateTxSigLimit),
		paramtypes.NewParamSetPair(KeyTxSizeCostPerByte, &p.TxSizeCostPerByte, validateTxSizeCostPerByte),
		paramtypes.NewParamSetPair(KeySigVerifyCostED25519, &p.SigVerifyCostED25519, validateSigVerifyCostED25519),
		paramtypes.NewParamSetPair(KeySigVerifyCostSecp256k1, &p.SigVerifyCostSecp256k1, validateSigVerifyCostSecp256k1),
	}
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
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateTxSigLimit(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid tx signature limit: %d", v)
	}

	return nil
}

func validateSigVerifyCostED25519(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid ED25519 signature verification cost: %d", v)
	}

	return nil
}

func validateSigVerifyCostSecp256k1(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid SECK256k1 signature verification cost: %d", v)
	}

	return nil
}

func validateMaxMemoCharacters(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid max memo characters: %d", v)
	}

	return nil
}

func validateTxSizeCostPerByte(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
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
