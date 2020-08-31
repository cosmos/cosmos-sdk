package types

import (
	"encoding"
	"encoding/json"
	"fmt"
	"testing"

	"math/big"
)

const maxBitLen = 255

func newIntegerFromString(s string) (*big.Int, bool) {
	return new(big.Int).SetString(s, 0)
}

func equal(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) == 0 }

func gt(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) == 1 }

func gte(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) >= 0 }

func lt(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) == -1 }

func lte(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) <= 0 }

func add(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Add(i, i2) }

func sub(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Sub(i, i2) }

func mul(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Mul(i, i2) }

func div(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Quo(i, i2) }

func mod(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Mod(i, i2) }

func neg(i *big.Int) *big.Int { return new(big.Int).Neg(i) }

func min(i *big.Int, i2 *big.Int) *big.Int {
	if i.Cmp(i2) == 1 {
		return new(big.Int).Set(i2)
	}

	return new(big.Int).Set(i)
}

func max(i *big.Int, i2 *big.Int) *big.Int {
	if i.Cmp(i2) == -1 {
		return new(big.Int).Set(i2)
	}

	return new(big.Int).Set(i)
}

func unmarshalText(i *big.Int, text string) error {
	if err := i.UnmarshalText([]byte(text)); err != nil {
		return err
	}

	if i.BitLen() > maxBitLen {
		return fmt.Errorf("integer out of range: %s", text)
	}

	return nil
}

var _ CustomProtobufType = (*Int)(nil)

// Int wraps integer with 256 bit range bound
// Checks overflow, underflow and division by zero
// Exists in range from -(2^maxBitLen-1) to 2^maxBitLen-1
type Int struct {
	i *big.Int
}

// BigInt converts Int to big.Int
func (i Int) BigInt() *big.Int {
	if i.IsNil() {
		return nil
	}
	return new(big.Int).Set(i.i)
}

// IsNil returns true if Int is uninitialized
func (i Int) IsNil() bool {
	return i.i == nil
}

// NewInt constructs Int from int64
func NewInt(n int64) Int {
	return Int{big.NewInt(n)}
}

// NewIntFromUint64 constructs an Int from a uint64.
func NewIntFromUint64(n uint64) Int {
	b := big.NewInt(0)
	b.SetUint64(n)
	return Int{b}
}

// NewIntFromBigInt constructs Int from big.Int
func NewIntFromBigInt(i *big.Int) Int {
	if i.BitLen() > maxBitLen {
		panic("NewIntFromBigInt() out of bound")
	}
	return Int{i}
}

// NewIntFromString constructs Int from string
func NewIntFromString(s string) (res Int, ok bool) {
	i, ok := newIntegerFromString(s)
	if !ok {
		return
	}
	// Check overflow
	if i.BitLen() > maxBitLen {
		ok = false
		return
	}
	return Int{i}, true
}

// NewIntWithDecimal constructs Int with decimal
// Result value is n*10^dec
func NewIntWithDecimal(n int64, dec int) Int {
	if dec < 0 {
		panic("NewIntWithDecimal() decimal is negative")
	}
	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(dec)), nil)
	i := new(big.Int)
	i.Mul(big.NewInt(n), exp)

	// Check overflow
	if i.BitLen() > maxBitLen {
		panic("NewIntWithDecimal() out of bound")
	}
	return Int{i}
}

// ZeroInt returns Int value with zero
func ZeroInt() Int { return Int{big.NewInt(0)} }

// OneInt returns Int value with one
func OneInt() Int { return Int{big.NewInt(1)} }

// ToDec converts Int to Dec
func (i Int) ToDec() Dec {
	return NewDecFromInt(i)
}

// Int64 converts Int to int64
// Panics if the value is out of range
func (i Int) Int64() int64 {
	if !i.i.IsInt64() {
		panic("Int64() out of bound")
	}
	return i.i.Int64()
}

// IsInt64 returns true if Int64() not panics
func (i Int) IsInt64() bool {
	return i.i.IsInt64()
}

// Uint64 converts Int to uint64
// Panics if the value is out of range
func (i Int) Uint64() uint64 {
	if !i.i.IsUint64() {
		panic("Uint64() out of bounds")
	}
	return i.i.Uint64()
}

// IsUint64 returns true if Uint64() not panics
func (i Int) IsUint64() bool {
	return i.i.IsUint64()
}

// IsZero returns true if Int is zero
func (i Int) IsZero() bool {
	return i.i.Sign() == 0
}

// IsNegative returns true if Int is negative
func (i Int) IsNegative() bool {
	return i.i.Sign() == -1
}

// IsPositive returns true if Int is positive
func (i Int) IsPositive() bool {
	return i.i.Sign() == 1
}

// Sign returns sign of Int
func (i Int) Sign() int {
	return i.i.Sign()
}

// Equal compares two Ints
func (i Int) Equal(i2 Int) bool {
	return equal(i.i, i2.i)
}

// GT returns true if first Int is greater than second
func (i Int) GT(i2 Int) bool {
	return gt(i.i, i2.i)
}

// GTE returns true if receiver Int is greater than or equal to the parameter
// Int.
func (i Int) GTE(i2 Int) bool {
	return gte(i.i, i2.i)
}

// LT returns true if first Int is lesser than second
func (i Int) LT(i2 Int) bool {
	return lt(i.i, i2.i)
}

// LTE returns true if first Int is less than or equal to second
func (i Int) LTE(i2 Int) bool {
	return lte(i.i, i2.i)
}

// Add adds Int from another
func (i Int) Add(i2 Int) (res Int) {
	res = Int{add(i.i, i2.i)}
	// Check overflow
	if res.i.BitLen() > maxBitLen {
		panic("Int overflow")
	}
	return
}

// AddRaw adds int64 to Int
func (i Int) AddRaw(i2 int64) Int {
	return i.Add(NewInt(i2))
}

// Sub subtracts Int from another
func (i Int) Sub(i2 Int) (res Int) {
	res = Int{sub(i.i, i2.i)}
	// Check overflow
	if res.i.BitLen() > maxBitLen {
		panic("Int overflow")
	}
	return
}

// SubRaw subtracts int64 from Int
func (i Int) SubRaw(i2 int64) Int {
	return i.Sub(NewInt(i2))
}

// Mul multiples two Ints
func (i Int) Mul(i2 Int) (res Int) {
	// Check overflow
	if i.i.BitLen()+i2.i.BitLen()-1 > maxBitLen {
		panic("Int overflow")
	}
	res = Int{mul(i.i, i2.i)}
	// Check overflow if sign of both are same
	if res.i.BitLen() > maxBitLen {
		panic("Int overflow")
	}
	return
}

// MulRaw multipies Int and int64
func (i Int) MulRaw(i2 int64) Int {
	return i.Mul(NewInt(i2))
}

// Quo divides Int with Int
func (i Int) Quo(i2 Int) (res Int) {
	// Check division-by-zero
	if i2.i.Sign() == 0 {
		panic("Division by zero")
	}
	return Int{div(i.i, i2.i)}
}

// QuoRaw divides Int with int64
func (i Int) QuoRaw(i2 int64) Int {
	return i.Quo(NewInt(i2))
}

// Mod returns remainder after dividing with Int
func (i Int) Mod(i2 Int) Int {
	if i2.Sign() == 0 {
		panic("division-by-zero")
	}
	return Int{mod(i.i, i2.i)}
}

// ModRaw returns remainder after dividing with int64
func (i Int) ModRaw(i2 int64) Int {
	return i.Mod(NewInt(i2))
}

// Neg negates Int
func (i Int) Neg() (res Int) {
	return Int{neg(i.i)}
}

// return the minimum of the ints
func MinInt(i1, i2 Int) Int {
	return Int{min(i1.BigInt(), i2.BigInt())}
}

// MaxInt returns the maximum between two integers.
func MaxInt(i, i2 Int) Int {
	return Int{max(i.BigInt(), i2.BigInt())}
}

// Human readable string
func (i Int) String() string {
	return i.i.String()
}

// MarshalJSON defines custom encoding scheme
func (i Int) MarshalJSON() ([]byte, error) {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return marshalJSON(i.i)
}

// UnmarshalJSON defines custom decoding scheme
func (i *Int) UnmarshalJSON(bz []byte) error {
	if i.i == nil { // Necessary since default Int initialization has i.i as nil
		i.i = new(big.Int)
	}
	return unmarshalJSON(i.i, bz)
}

// MarshalJSON for custom encoding scheme
// Must be encoded as a string for JSON precision
func marshalJSON(i encoding.TextMarshaler) ([]byte, error) {
	text, err := i.MarshalText()
	if err != nil {
		return nil, err
	}

	return json.Marshal(string(text))
}

// UnmarshalJSON for custom decoding scheme
// Must be encoded as a string for JSON precision
func unmarshalJSON(i *big.Int, bz []byte) error {
	var text string
	if err := json.Unmarshal(bz, &text); err != nil {
		return err
	}

	return unmarshalText(i, text)
}

// MarshalYAML returns the YAML representation.
func (i Int) MarshalYAML() (interface{}, error) {
	return i.String(), nil
}

// Marshal implements the gogo proto custom type interface.
func (i Int) Marshal() ([]byte, error) {
	if i.i == nil {
		i.i = new(big.Int)
	}
	return i.i.MarshalText()
}

// MarshalTo implements the gogo proto custom type interface.
func (i *Int) MarshalTo(data []byte) (n int, err error) {
	if i.i == nil {
		i.i = new(big.Int)
	}
	if len(i.i.Bytes()) == 0 {
		copy(data, []byte{0x30})
		return 1, nil
	}

	bz, err := i.Marshal()
	if err != nil {
		return 0, err
	}

	copy(data, bz)
	return len(bz), nil
}

// Unmarshal implements the gogo proto custom type interface.
func (i *Int) Unmarshal(data []byte) error {
	if len(data) == 0 {
		i = nil
		return nil
	}

	if i.i == nil {
		i.i = new(big.Int)
	}

	if err := i.i.UnmarshalText(data); err != nil {
		return err
	}

	if i.i.BitLen() > maxBitLen {
		return fmt.Errorf("integer out of range; got: %d, max: %d", i.i.BitLen(), maxBitLen)
	}

	return nil
}

// Size implements the gogo proto custom type interface.
func (i *Int) Size() int {
	bz, _ := i.Marshal()
	return len(bz)
}

// Override Amino binary serialization by proxying to protobuf.
func (i Int) MarshalAmino() ([]byte, error)   { return i.Marshal() }
func (i *Int) UnmarshalAmino(bz []byte) error { return i.Unmarshal(bz) }

// intended to be used with require/assert:  require.True(IntEq(...))
func IntEq(t *testing.T, exp, got Int) (*testing.T, bool, string, string, string) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}

func (ip IntProto) String() string {
	return ip.Int.String()
}
