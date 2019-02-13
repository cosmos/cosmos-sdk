package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"math/big"
	"math/rand"
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

func div(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Div(i, i2) }

func mod(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Mod(i, i2) }

func neg(i *big.Int) *big.Int { return new(big.Int).Neg(i) }

func random(i *big.Int) *big.Int { return new(big.Int).Rand(rand.New(rand.NewSource(rand.Int63())), i) }

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

// MarshalAmino for custom encoding scheme
func marshalAmino(i *big.Int) (string, error) {
	bz, err := i.MarshalText()
	return string(bz), err
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

// UnmarshalAmino for custom decoding scheme
func unmarshalAmino(i *big.Int, text string) (err error) {
	return unmarshalText(i, text)
}

// MarshalJSON for custom encoding scheme
// Must be encoded as a string for JSON precision
func marshalJSON(i *big.Int) ([]byte, error) {
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
	err := json.Unmarshal(bz, &text)
	if err != nil {
		return err
	}

	return unmarshalText(i, text)
}

// Int wraps integer with 256 bit range bound
// Checks overflow, underflow and division by zero
// Exists in range from -(2^maxBitLen-1) to 2^maxBitLen-1
type Int struct {
	i *big.Int
}

// BigInt converts Int to big.Int
func (i Int) BigInt() *big.Int {
	return new(big.Int).Set(i.i)
}

// NewInt constructs Int from int64
func NewInt(n int64) Int {
	return Int{big.NewInt(n)}
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

// Div divides Int with Int
func (i Int) Div(i2 Int) (res Int) {
	// Check division-by-zero
	if i2.i.Sign() == 0 {
		panic("Division by zero")
	}
	return Int{div(i.i, i2.i)}
}

// DivRaw divides Int with int64
func (i Int) DivRaw(i2 int64) Int {
	return i.Div(NewInt(i2))
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

// Testing purpose random Int generator
func randomInt(i Int) Int {
	return NewIntFromBigInt(random(i.BigInt()))
}

// MarshalAmino defines custom encoding scheme
func (i Int) MarshalAmino() (string, error) {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return marshalAmino(i.i)
}

// UnmarshalAmino defines custom decoding scheme
func (i *Int) UnmarshalAmino(text string) error {
	if i.i == nil { // Necessary since default Int initialization has i.i as nil
		i.i = new(big.Int)
	}
	return unmarshalAmino(i.i, text)
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

// Int wraps integer with 256 bit range bound
// Checks overflow, underflow and division by zero
// Exists in range from 0 to 2^256-1
type Uint struct {
	i *big.Int
}

// BigInt converts Uint to big.Unt
func (i Uint) BigInt() *big.Int {
	return new(big.Int).Set(i.i)
}

// NewUint constructs Uint from int64
func NewUint(n uint64) Uint {
	i := new(big.Int)
	i.SetUint64(n)
	return Uint{i}
}

// NewUintFromBigUint constructs Uint from big.Uint
func NewUintFromBigInt(i *big.Int) Uint {
	res := Uint{i}
	if UintOverflow(res) {
		panic("Uint overflow")
	}
	return res
}

// NewUintFromString constructs Uint from string
func NewUintFromString(s string) (res Uint, ok bool) {
	i, ok := newIntegerFromString(s)
	if !ok {
		return
	}
	// Check overflow
	if i.Sign() == -1 || i.Sign() == 1 && i.BitLen() > 256 {
		ok = false
		return
	}
	return Uint{i}, true
}

// NewUintWithDecimal constructs Uint with decimal
// Result value is n*10^dec
func NewUintWithDecimal(n uint64, dec int) Uint {
	if dec < 0 {
		panic("NewUintWithDecimal() decimal is negative")
	}
	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(dec)), nil)
	i := new(big.Int)
	i.Mul(new(big.Int).SetUint64(n), exp)

	res := Uint{i}
	if UintOverflow(res) {
		panic("NewUintWithDecimal() out of bound")
	}

	return res
}

// ZeroUint returns Uint value with zero
func ZeroUint() Uint { return Uint{big.NewInt(0)} }

// OneUint returns Uint value with one
func OneUint() Uint { return Uint{big.NewInt(1)} }

// Uint64 converts Uint to uint64
// Panics if the value is out of range
func (i Uint) Uint64() uint64 {
	if !i.i.IsUint64() {
		panic("Uint64() out of bound")
	}
	return i.i.Uint64()
}

// IsUint64 returns true if Uint64() not panics
func (i Uint) IsUint64() bool {
	return i.i.IsUint64()
}

// IsZero returns true if Uint is zero
func (i Uint) IsZero() bool {
	return i.i.Sign() == 0
}

// Sign returns sign of Uint
func (i Uint) Sign() int {
	return i.i.Sign()
}

// Equal compares two Uints
func (i Uint) Equal(i2 Uint) bool {
	return equal(i.i, i2.i)
}

// GT returns true if first Uint is greater than second
func (i Uint) GT(i2 Uint) bool {
	return gt(i.i, i2.i)
}

// LT returns true if first Uint is lesser than second
func (i Uint) LT(i2 Uint) bool {
	return lt(i.i, i2.i)
}

// Add adds Uint from another
func (i Uint) Add(i2 Uint) (res Uint) {
	res = Uint{add(i.i, i2.i)}
	if UintOverflow(res) {
		panic("Uint overflow")
	}
	return
}

// AddRaw adds uint64 to Uint
func (i Uint) AddRaw(i2 uint64) Uint {
	return i.Add(NewUint(i2))
}

// Sub subtracts Uint from another
func (i Uint) Sub(i2 Uint) (res Uint) {
	res = Uint{sub(i.i, i2.i)}
	if UintOverflow(res) {
		panic("Uint overflow")
	}
	return
}

// SafeSub attempts to subtract one Uint from another. A boolean is also returned
// indicating if the result contains integer overflow.
func (i Uint) SafeSub(i2 Uint) (Uint, bool) {
	res := Uint{sub(i.i, i2.i)}
	if UintOverflow(res) {
		return res, true
	}

	return res, false
}

// SubRaw subtracts uint64 from Uint
func (i Uint) SubRaw(i2 uint64) Uint {
	return i.Sub(NewUint(i2))
}

// Mul multiples two Uints
func (i Uint) Mul(i2 Uint) (res Uint) {
	if i.i.BitLen()+i2.i.BitLen()-1 > 256 {
		panic("Uint overflow")
	}

	res = Uint{mul(i.i, i2.i)}
	if UintOverflow(res) {
		panic("Uint overflow")
	}

	return
}

// MulRaw multipies Uint and uint64
func (i Uint) MulRaw(i2 uint64) Uint {
	return i.Mul(NewUint(i2))
}

// Div divides Uint with Uint
func (i Uint) Div(i2 Uint) (res Uint) {
	// Check division-by-zero
	if i2.Sign() == 0 {
		panic("division-by-zero")
	}
	return Uint{div(i.i, i2.i)}
}

// Div divides Uint with uint64
func (i Uint) DivRaw(i2 uint64) Uint {
	return i.Div(NewUint(i2))
}

// Mod returns remainder after dividing with Uint
func (i Uint) Mod(i2 Uint) Uint {
	if i2.Sign() == 0 {
		panic("division-by-zero")
	}
	return Uint{mod(i.i, i2.i)}
}

// ModRaw returns remainder after dividing with uint64
func (i Uint) ModRaw(i2 uint64) Uint {
	return i.Mod(NewUint(i2))
}

// Return the minimum of the Uints
func MinUint(i1, i2 Uint) Uint {
	return Uint{min(i1.BigInt(), i2.BigInt())}
}

// MaxUint returns the maximum between two unsigned integers.
func MaxUint(i, i2 Uint) Uint {
	return Uint{max(i.BigInt(), i2.BigInt())}
}

// Human readable string
func (i Uint) String() string {
	return i.i.String()
}

// Testing purpose random Uint generator
func randomUint(i Uint) Uint {
	return NewUintFromBigInt(random(i.BigInt()))
}

// MarshalAmino defines custom encoding scheme
func (i Uint) MarshalAmino() (string, error) {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return marshalAmino(i.i)
}

// UnmarshalAmino defines custom decoding scheme
func (i *Uint) UnmarshalAmino(text string) error {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return unmarshalAmino(i.i, text)
}

// MarshalJSON defines custom encoding scheme
func (i Uint) MarshalJSON() ([]byte, error) {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return marshalJSON(i.i)
}

// UnmarshalJSON defines custom decoding scheme
func (i *Uint) UnmarshalJSON(bz []byte) error {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return unmarshalJSON(i.i, bz)
}

//__________________________________________________________________________

// UintOverflow returns true if a given unsigned integer overflows and false
// otherwise.
func UintOverflow(x Uint) bool {
	return x.i.Sign() == -1 || x.i.Sign() == 1 && x.i.BitLen() > 256
}

// intended to be used with require/assert:  require.True(IntEq(...))
func IntEq(t *testing.T, exp, got Int) (*testing.T, bool, string, string, string) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}
