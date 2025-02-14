package math

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// Rapid is a Go library for property-based testing.
func TestDecWithRapid(t *testing.T) {
	// Property tests
	t.Run("TestNewDecFromInt64", rapid.MakeCheck(testDecInt64))

	// Properties about *FromString functions
	t.Run("TestInvalidNewDecFromString", rapid.MakeCheck(testInvalidNewDecFromString))

	// Properties about addition
	t.Run("TestAddLeftIdentity", rapid.MakeCheck(testAddLeftIdentity))
	t.Run("TestAddRightIdentity", rapid.MakeCheck(testAddRightIdentity))
	t.Run("TestAddCommutative", rapid.MakeCheck(testAddCommutative))
	t.Run("TestAddAssociative", rapid.MakeCheck(testAddAssociative))

	// Properties about subtraction
	t.Run("TestSubRightIdentity", rapid.MakeCheck(testSubRightIdentity))
	t.Run("TestSubZero", rapid.MakeCheck(testSubZero))

	// Properties about multiplication
	t.Run("TestMulLeftIdentity", rapid.MakeCheck(testMulLeftIdentity))
	t.Run("TestMulRightIdentity", rapid.MakeCheck(testMulRightIdentity))
	t.Run("TestMulCommutative", rapid.MakeCheck(testMulCommutative))
	t.Run("TestMulAssociative", rapid.MakeCheck(testMulAssociative))
	t.Run("TestZeroIdentity", rapid.MakeCheck(testMulZero))

	// Properties about division
	t.Run("TestDivisionBySelf", rapid.MakeCheck(testSelfQuo))
	t.Run("TestDivisionByOne", rapid.MakeCheck(testQuoByOne))

	// Properties combining operations
	t.Run("TestSubAdd", rapid.MakeCheck(testSubAdd))
	t.Run("TestAddSub", rapid.MakeCheck(testAddSub))
	t.Run("TestMulQuoA", rapid.MakeCheck(testMulQuoA))
	t.Run("TestMulQuoB", rapid.MakeCheck(testMulQuoB))
	t.Run("TestMulQuoExact", rapid.MakeCheck(testMulQuoExact))
	t.Run("TestQuoMulExact", rapid.MakeCheck(testQuoMulExact))

	// Properties about comparison and equality
	t.Run("TestCmpInverse", rapid.MakeCheck(testCmpInverse))
	t.Run("TestEqualCommutative", rapid.MakeCheck(testEqualCommutative))

	// Properties about tests on a single Dec
	t.Run("TestIsZero", rapid.MakeCheck(testIsZero))
	t.Run("TestIsNegative", rapid.MakeCheck(testIsNegative))
	t.Run("TestIsPositive", rapid.MakeCheck(testIsPositive))
	t.Run("TestNumDecimalPlaces", rapid.MakeCheck(testNumDecimalPlaces))

	// Unit tests
	zero := Dec{}
	one := NewDecFromInt64(1)
	two := NewDecFromInt64(2)
	three := NewDecFromInt64(3)
	four := NewDecFromInt64(4)
	five := NewDecFromInt64(5)
	minusOne := NewDecFromInt64(-1)

	onePointOneFive, err := NewDecFromString("1.15")
	require.NoError(t, err)
	twoPointThreeFour, err := NewDecFromString("2.34")
	require.NoError(t, err)
	threePointFourNine, err := NewDecFromString("3.49")
	require.NoError(t, err)
	onePointFourNine, err := NewDecFromString("1.49")
	require.NoError(t, err)
	minusFivePointZero, err := NewDecFromString("-5.0")
	require.NoError(t, err)

	twoThousand := NewDecWithExp(2, 3)
	require.True(t, twoThousand.Equal(NewDecFromInt64(2000)))

	res, err := two.Add(zero)
	require.NoError(t, err)
	require.True(t, res.Equal(two))

	res, err = five.Sub(two)
	require.NoError(t, err)
	require.True(t, res.Equal(three))

	res, err = four.Quo(two)
	require.NoError(t, err)
	require.True(t, res.Equal(two))

	res, err = five.QuoInteger(two)
	require.NoError(t, err)
	require.True(t, res.Equal(two))

	res, err = five.Modulo(two)
	require.NoError(t, err)
	require.True(t, res.Equal(one))

	x, err := four.Int64()
	require.NoError(t, err)
	require.Equal(t, int64(4), x)

	require.Equal(t, "5", five.String())

	res, err = onePointOneFive.Add(twoPointThreeFour)
	require.NoError(t, err)
	require.True(t, res.Equal(threePointFourNine))

	res, err = threePointFourNine.Sub(two)
	require.NoError(t, err)
	require.True(t, res.Equal(onePointFourNine))

	res, err = minusOne.Sub(four)
	require.NoError(t, err)
	require.True(t, res.Equal(minusFivePointZero))

	require.True(t, zero.IsZero())
	require.False(t, zero.IsPositive())
	require.False(t, zero.IsNegative())

	require.False(t, one.IsZero())
	require.True(t, one.IsPositive())
	require.False(t, one.IsNegative())

	require.False(t, minusOne.IsZero())
	require.False(t, minusOne.IsPositive())
	require.True(t, minusOne.IsNegative())

	res, err = one.MulExact(two)
	require.NoError(t, err)
	require.True(t, res.Equal(two))
}

var genDec *rapid.Generator[Dec] = rapid.Custom(func(t *rapid.T) Dec {
	f := rapid.Float64().Draw(t, "f")
	dec, err := NewDecFromString(fmt.Sprintf("%g", f))
	require.NoError(t, err)
	return dec
})

// A Dec value and the float used to create it
type floatAndDec struct {
	float float64
	dec   Dec
}

// Generate a Dec value along with the float used to create it
var genFloatAndDec *rapid.Generator[floatAndDec] = rapid.Custom(func(t *rapid.T) floatAndDec {
	f := rapid.Float64().Draw(t, "f")
	dec, err := NewDecFromString(fmt.Sprintf("%g", f))
	require.NoError(t, err)
	return floatAndDec{f, dec}
})

// Property: n == NewDecFromInt64(n).Int64()
func testDecInt64(t *rapid.T) {
	nIn := rapid.Int64().Draw(t, "n")
	nOut, err := NewDecFromInt64(nIn).Int64()

	require.NoError(t, err)
	require.Equal(t, nIn, nOut)
}

// Property: invalid_number_string(s) => NewDecFromString(s) == err
func testInvalidNewDecFromString(t *rapid.T) {
	s := rapid.StringMatching("[[:alpha:]]+").Draw(t, "s")
	_, err := NewDecFromString(s)
	require.Error(t, err)
}

// Property: 0 + a == a
func testAddLeftIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a")
	zero := NewDecFromInt64(0)

	b, err := zero.Add(a)
	require.NoError(t, err)

	require.True(t, a.Equal(b))
}

// Property: a + 0 == a
func testAddRightIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a")
	zero := NewDecFromInt64(0)

	b, err := a.Add(zero)
	require.NoError(t, err)

	require.True(t, a.Equal(b))
}

// Property: a + b == b + a
func testAddCommutative(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")

	c, err := a.Add(b)
	require.NoError(t, err)

	d, err := b.Add(a)
	require.NoError(t, err)

	require.True(t, c.Equal(d))
}

// Property: (a + b) + c == a + (b + c)
func testAddAssociative(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")
	c := genDec.Draw(t, "c")

	// (a + b) + c
	d, err := a.Add(b)
	require.NoError(t, err)

	e, err := d.Add(c)
	require.NoError(t, err)

	// a + (b + c)
	f, err := b.Add(c)
	require.NoError(t, err)

	g, err := a.Add(f)
	require.NoError(t, err)

	require.True(t, e.Equal(g))
}

// Property: a - 0 == a
func testSubRightIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a")
	zero := NewDecFromInt64(0)

	b, err := a.Sub(zero)
	require.NoError(t, err)

	require.True(t, a.Equal(b))
}

// Property: a - a == 0
func testSubZero(t *rapid.T) {
	a := genDec.Draw(t, "a")
	zero := NewDecFromInt64(0)

	b, err := a.Sub(a)
	require.NoError(t, err)

	require.True(t, b.Equal(zero))
}

// Property: 1 * a == a
func testMulLeftIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a")
	one := NewDecFromInt64(1)

	b, err := one.Mul(a)
	require.NoError(t, err)

	require.True(t, a.Equal(b))
}

// Property: a * 1 == a
func testMulRightIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a")
	one := NewDecFromInt64(1)

	b, err := a.Mul(one)
	require.NoError(t, err)

	require.True(t, a.Equal(b))
}

// Property: a * b == b * a
func testMulCommutative(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")

	c, err := a.Mul(b)
	require.NoError(t, err)

	d, err := b.Mul(a)
	require.NoError(t, err)

	require.True(t, c.Equal(d))
}

// Property: (a * b) * c == a * (b * c)
func testMulAssociative(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")
	c := genDec.Draw(t, "c")

	// (a * b) * c
	d, err := a.Mul(b)
	require.NoError(t, err)

	e, err := d.Mul(c)
	require.NoError(t, err)

	// a * (b * c)
	f, err := b.Mul(c)
	require.NoError(t, err)

	g, err := a.Mul(f)
	require.NoError(t, err)

	require.True(t, e.Equal(g))
}

// Property: (a - b) + b == a
func testSubAdd(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")

	c, err := a.Sub(b)
	require.NoError(t, err)

	d, err := c.Add(b)
	require.NoError(t, err)

	require.True(t, a.Equal(d))
}

// Property: (a + b) - b == a
func testAddSub(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")

	c, err := a.Add(b)
	require.NoError(t, err)

	d, err := c.Sub(b)
	require.NoError(t, err)

	require.True(t, a.Equal(d))
}

// Property: a * 0 = 0
func testMulZero(t *rapid.T) {
	a := genDec.Draw(t, "a")
	zero := Dec{}

	c, err := a.Mul(zero)
	require.NoError(t, err)
	require.True(t, c.IsZero())
}

// Property: a/a = 1
func testSelfQuo(t *rapid.T) {
	decNotZero := func(d Dec) bool { return !d.IsZero() }
	a := genDec.Filter(decNotZero).Draw(t, "a")
	one := NewDecFromInt64(1)

	b, err := a.Quo(a)
	require.NoError(t, err)
	require.True(t, one.Equal(b))
}

// Property: a/1 = a
func testQuoByOne(t *rapid.T) {
	a := genDec.Draw(t, "a")
	one := NewDecFromInt64(1)

	b, err := a.Quo(one)
	require.NoError(t, err)
	require.True(t, a.Equal(b))
}

// Property: (a * b) / a == b
func testMulQuoA(t *rapid.T) {
	decNotZero := func(d Dec) bool { return !d.IsZero() }
	a := genDec.Filter(decNotZero).Draw(t, "a")
	b := genDec.Draw(t, "b")

	c, err := a.Mul(b)
	require.NoError(t, err)

	d, err := c.Quo(a)
	require.NoError(t, err)

	require.True(t, b.Equal(d))
}

// Property: (a * b) / b == a
func testMulQuoB(t *rapid.T) {
	decNotZero := func(d Dec) bool { return !d.IsZero() }
	a := genDec.Draw(t, "a")
	b := genDec.Filter(decNotZero).Draw(t, "b")

	c, err := a.Mul(b)
	require.NoError(t, err)

	d, err := c.Quo(b)
	require.NoError(t, err)

	require.True(t, a.Equal(d))
}

// Property: (a * 10^b) / 10^b == a using MulExact and QuoExact
// and a with no more than b decimal places (b <= 32).
func testMulQuoExact(t *rapid.T) {
	b := rapid.Uint32Range(0, 32).Draw(t, "b")
	decPrec := func(d Dec) bool { return d.NumDecimalPlaces() <= b }
	a := genDec.Filter(decPrec).Draw(t, "a")

	c := NewDecWithExp(1, int32(b))

	d, err := a.MulExact(c)
	require.NoError(t, err)

	e, err := d.QuoExact(c)
	require.NoError(t, err)

	require.True(t, a.Equal(e))
}

// Property: (a / b) * b == a using QuoExact and MulExact and
// a as an integer.
func testQuoMulExact(t *rapid.T) {
	a := rapid.Uint64().Draw(t, "a")
	aDec, err := NewDecFromString(fmt.Sprintf("%d", a))
	require.NoError(t, err)
	b := rapid.Uint32Range(0, 32).Draw(t, "b")
	c := NewDecWithExp(1, int32(b))

	require.NoError(t, err)

	d, err := aDec.QuoExact(c)
	require.NoError(t, err)

	e, err := d.MulExact(c)
	require.NoError(t, err)

	require.True(t, aDec.Equal(e))
}

// Property: Cmp(a, b) == -Cmp(b, a)
func testCmpInverse(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")

	require.Equal(t, a.Cmp(b), -b.Cmp(a))
}

// Property: Equal(a, b) == Equal(b, a)
func testEqualCommutative(t *rapid.T) {
	a := genDec.Draw(t, "a")
	b := genDec.Draw(t, "b")

	require.Equal(t, a.Equal(b), b.Equal(a))
}

// Property: isZero(f) == isZero(NewDecFromString(f.String()))
func testIsZero(t *rapid.T) {
	floatAndDec := genFloatAndDec.Draw(t, "floatAndDec")
	f, dec := floatAndDec.float, floatAndDec.dec

	require.Equal(t, f == 0, dec.IsZero())
}

// Property: isNegative(f) == isNegative(NewDecFromString(f.String()))
func testIsNegative(t *rapid.T) {
	floatAndDec := genFloatAndDec.Draw(t, "floatAndDec")
	f, dec := floatAndDec.float, floatAndDec.dec

	require.Equal(t, f < 0, dec.IsNegative())
}

// Property: isPositive(f) == isPositive(NewDecFromString(f.String()))
func testIsPositive(t *rapid.T) {
	floatAndDec := genFloatAndDec.Draw(t, "floatAndDec")
	f, dec := floatAndDec.float, floatAndDec.dec

	require.Equal(t, f > 0, dec.IsPositive())
}

// Property: floatDecimalPlaces(f) == NumDecimalPlaces(NewDecFromString(f.String()))
func testNumDecimalPlaces(t *rapid.T) {
	floatAndDec := genFloatAndDec.Draw(t, "floatAndDec")
	f, dec := floatAndDec.float, floatAndDec.dec

	require.Equal(t, floatDecimalPlaces(t, f), dec.NumDecimalPlaces())
}

func floatDecimalPlaces(t *rapid.T, f float64) uint32 {
	reScientific := regexp.MustCompile(`^\-?(?:[[:digit:]]+(?:\.([[:digit:]]+))?|\.([[:digit:]]+))(?:e?(?:\+?([[:digit:]]+)|(-[[:digit:]]+)))?$`)
	fStr := fmt.Sprintf("%g", f)
	matches := reScientific.FindAllStringSubmatch(fStr, 1)
	if len(matches) != 1 {
		t.Fatalf("Didn't match float: %g", f)
	}

	// basePlaces is the number of decimal places in the decimal part of the
	// string
	basePlaces := 0
	if matches[0][1] != "" {
		basePlaces = len(matches[0][1])
	} else if matches[0][2] != "" {
		basePlaces = len(matches[0][2])
	}
	t.Logf("Base places: %d", basePlaces)

	// exp is the exponent
	exp := 0
	if matches[0][3] != "" {
		var err error
		exp, err = strconv.Atoi(matches[0][3])
		require.NoError(t, err)
	} else if matches[0][4] != "" {
		var err error
		exp, err = strconv.Atoi(matches[0][4])
		require.NoError(t, err)
	}

	// Subtract exponent from base and check if negative
	res := basePlaces - exp
	if res <= 0 {
		return 0
	}

	return uint32(res)
}
