package types

import (
	"math/big"
)

func newIntegerFromString(s string) (*big.Int, bool) {
	return new(big.Int).SetString(s, 0)
}

func newIntegerWithDecimal(n int64, dec int) (res *big.Int) {
	if dec < 0 {
		return
	}
	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(dec)), nil)
	i := new(big.Int)
	return i.Mul(big.NewInt(n), exp)
}

func equal(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) == 0 }

func gt(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) == 1 }

func lt(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) == -1 }

func add(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Add(i, i2) }

func sub(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Sub(i, i2) }

func mul(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Mul(i, i2) }

func div(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Div(i, i2) }

func mod(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Mod(i, i2) }

func neg(i *big.Int) *big.Int { return new(big.Int).Neg(i) }

// MarshalAmino for custom encoding scheme
func marshalAmino(i *big.Int) (string, error) {
	bz, err := i.MarshalText()
	return string(bz), err
}

// UnmarshalAmino for custom decoding scheme
func unmarshalAmino(i *big.Int, text string) (err error) {
	return i.UnmarshalText([]byte(text))
}

// MarshalJSON for custom encodig scheme
func marshalJSON(i *big.Int) ([]byte, error) {
	return i.MarshalText()
}

// UnmarshalJSON for custom decoding scheme
func unmarshalJSON(i *big.Int, bz []byte) error {
	return i.UnmarshalText(bz)
}

// Int wraps integer with 256 bit range bound
// Checks overflow, underflow and division by zero
// Exists in range from -(2^255-1) to 2^255-1
type Int struct {
	i big.Int
}

// BigInt converts Int to big.Int
func (i Int) BigInt() *big.Int {
	return new(big.Int).Set(&(i.i))
}

// NewInt constructs Int from int64
func NewInt(n int64) Int {
	return Int{*big.NewInt(n)}
}

// NewIntFromBigInt constructs Int from big.Int
func NewIntFromBigInt(i *big.Int) Int {
	if i.BitLen() > 255 {
		panic("NewIntFromBigInt() out of bound")
	}
	return Int{*i}
}

// NewIntFromString constructs Int from string
func NewIntFromString(s string) (res Int, ok bool) {
	i, ok := newIntegerFromString(s)
	if !ok {
		return
	}
	// Check overflow
	if i.BitLen() > 255 {
		ok = false
		return
	}
	return Int{*i}, true
}

// NewIntWithDecimal constructs Int with decimal
// Result value is n*10^dec
func NewIntWithDecimal(n int64, dec int) Int {
	i := newIntegerWithDecimal(n, dec)
	// Check overflow
	if i.BitLen() > 255 {
		panic("NewIntWithDecimal() out of bound")
	}
	return Int{*i}
}

// ZeroInt returns Int value with zero
func ZeroInt() Int { return Int{*big.NewInt(0)} }

// OneInt returns Int value with one
func OneInt() Int { return Int{*big.NewInt(1)} }

// Int64 converts Int to int64
// Panics if the value is out of range
func (i Int) Int64() int64 {
	if !i.i.IsInt64() {
		panic("Int64() out of bound")
	}
	return i.i.Int64()
}

// IsZero returns true if Int is zero
func (i Int) IsZero() bool {
	return i.i.Sign() == 0
}

// Sign returns sign of Int
func (i Int) Sign() int {
	return i.i.Sign()
}

// Equal compares two Ints
func (i Int) Equal(i2 Int) bool {
	return equal(&(i.i), &(i2.i))
}

// GT returns true if first Int is greater than second
func (i Int) GT(i2 Int) bool {
	return gt((&i.i), &(i2.i))
}

// LT returns true if first Int is lesser than second
func (i Int) LT(i2 Int) bool {
	return lt((&i.i), &(i2.i))
}

// Add adds Int from another
func (i Int) Add(i2 Int) (res Int) {
	res = Int{*add(&(i.i), &(i2.i))}
	// Check overflow
	if res.i.BitLen() > 255 {
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
	res = Int{*sub(&(i.i), &(i2.i))}
	// Check overflow
	if res.i.BitLen() > 255 {
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
	if i.i.BitLen()+i2.i.BitLen()-1 > 255 {
		panic("Int overflow")
	}
	res = Int{*mul(&(i.i), &(i2.i))}
	// Check overflow if sign of both are same
	if res.i.BitLen() > 255 {
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
	return Int{*div(&(i.i), &(i2.i))}
}

// DivRaw divides Int with int64
func (i Int) DivRaw(i2 int64) Int {
	return i.Div(NewInt(i2))
}

// Neg negates Int
func (i Int) Neg() (res Int) {
	return Int{*neg(&(i.i))}
}

// MarshalAmino defines custom encoding scheme
func (i Int) MarshalAmino() (string, error) {
	return marshalAmino(&(i.i))
}

// UnmarshalAmino defines custom decoding scheme
func (i *Int) UnmarshalAmino(text string) error {
	return unmarshalAmino(&(i.i), text)
}

// MarshalJSON defines custom encoding scheme
func (i Int) MarshalJSON() ([]byte, error) {
	return marshalJSON(&(i.i))
}

// UnmarshalJSON defines custom decoding scheme
func (i *Int) UnmarshalJSON(bz []byte) error {
	return unmarshalJSON(&(i.i), bz)
}

// Int wraps integer with 256 bit range bound
// Checks overflow, underflow and division by zero
// Exists in range from 0 to 2^256-1
type Uint struct {
	i big.Int
}

// BigInt converts Uint to big.Unt
func (i Uint) BigInt() *big.Int {
	return new(big.Int).Set(&(i.i))
}

// NewUint constructs Uint from int64
func NewUint(n uint64) Uint {
	i := new(big.Int)
	i.SetUint64(n)
	return Uint{*i}
}

// NewUintFromBigUint constructs Uint from big.Uint
func NewUintFromBigInt(i *big.Int) Uint {
	// Check overflow
	if i.Sign() == -1 || i.Sign() == 1 && i.BitLen() > 256 {
		panic("Uint overflow")
	}
	return Uint{*i}
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
	return Uint{*i}, true
}

// NewUintWithDecimal constructs Uint with decimal
// Result value is n*10^dec
func NewUintWithDecimal(n int64, dec int) Uint {
	i := newIntegerWithDecimal(n, dec)
	// Check overflow
	if i.Sign() == -1 || i.Sign() == 1 && i.BitLen() > 256 {
		panic("NewUintWithDecimal() out of bound")
	}
	return Uint{*i}
}

// ZeroUint returns Uint value with zero
func ZeroUint() Uint { return Uint{*big.NewInt(0)} }

// OneUint returns Uint value with one
func OneUint() Uint { return Uint{*big.NewInt(1)} }

// Uint64 converts Uint to uint64
// Panics if the value is out of range
func (i Uint) Uint64() uint64 {
	if !i.i.IsUint64() {
		panic("Uint64() out of bound")
	}
	return i.i.Uint64()
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
	return equal(&(i.i), &(i2.i))
}

// GT returns true if first Uint is greater than second
func (i Uint) GT(i2 Uint) bool {
	return gt(&(i.i), &(i2.i))
}

// LT returns true if first Uint is lesser than second
func (i Uint) LT(i2 Uint) bool {
	return lt(&(i.i), &(i2.i))
}

// Add adds Uint from another
func (i Uint) Add(i2 Uint) (res Uint) {
	res = Uint{*add(&(i.i), &(i2.i))}
	// Check overflow
	if res.Sign() == -1 || res.Sign() == 1 && res.i.BitLen() > 256 {
		panic("Uint overflow")
	}
	return
}

// AddRaw adds int64 to Uint
func (i Uint) AddRaw(i2 uint64) Uint {
	return i.Add(NewUint(i2))
}

// Sub subtracts Uint from another
func (i Uint) Sub(i2 Uint) (res Uint) {
	res = Uint{*sub(&(i.i), &(i2.i))}
	// Check overflow
	if res.Sign() == -1 || res.Sign() == 1 && res.i.BitLen() > 256 {
		panic("Uint overflow")
	}
	return
}

// SubRaw subtracts int64 from Uint
func (i Uint) SubRaw(i2 uint64) Uint {
	return i.Sub(NewUint(i2))
}

// Mul multiples two Uints
func (i Uint) Mul(i2 Uint) (res Uint) {
	// Check overflow
	if i.i.BitLen()+i2.i.BitLen()-1 > 256 {
		panic("Uint overflow")
	}
	res = Uint{*mul(&(i.i), &(i2.i))}
	// Check overflow
	if res.Sign() == -1 || res.Sign() == 1 && res.i.BitLen() > 256 {
		panic("Uint overflow")
	}
	return
}

// MulRaw multipies Uint and int64
func (i Uint) MulRaw(i2 uint64) Uint {
	return i.Mul(NewUint(i2))
}

// Div divides Uint with Uint
func (i Uint) Div(i2 Uint) (res Uint) {
	// Check division-by-zero
	if i2.Sign() == 0 {
		panic("division-by-zero")
	}
	return Uint{*div(&(i.i), &(i2.i))}
}

// Div divides Uint with int64
func (i Uint) DivRaw(i2 uint64) Uint {
	return i.Div(NewUint(i2))
}

// MarshalAmino defines custom encoding scheme
func (i Uint) MarshalAmino() (string, error) {
	return marshalAmino(&(i.i))
}

// UnmarshalAmino defines custom decoding scheme
func (i *Uint) UnmarshalAmino(text string) error {
	return unmarshalAmino(&(i.i), text)
}

// MarshalJSON defines custom encoding scheme
func (i Uint) MarshalJSON() ([]byte, error) {
	return marshalJSON(&(i.i))
}

// UnmarshalJSON defines custom decoding scheme
func (i *Uint) UnmarshalJSON(bz []byte) error {
	return unmarshalJSON(&(i.i), bz)
}
