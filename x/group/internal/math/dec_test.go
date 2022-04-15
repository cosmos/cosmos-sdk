package math

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestDec(t *testing.T) {
	// Property tests
	t.Run("TestNewDecFromInt64", rapid.MakeCheck(testDecInt64))

	// Properties about addition
	t.Run("TestAddLeftIdentity", rapid.MakeCheck(testAddLeftIdentity))
	t.Run("TestAddRightIdentity", rapid.MakeCheck(testAddRightIdentity))
	t.Run("TestAddCommutative", rapid.MakeCheck(testAddCommutative))
	t.Run("TestAddAssociative", rapid.MakeCheck(testAddAssociative))

	// Properties about subtraction
	t.Run("TestSubRightIdentity", rapid.MakeCheck(testSubRightIdentity))
	t.Run("TestSubZero", rapid.MakeCheck(testSubZero))

	// Properties combining operations
	t.Run("TestSubAdd", rapid.MakeCheck(testSubAdd))
	t.Run("TestAddSub", rapid.MakeCheck(testAddSub))

	// Properties about comparision and equality
	t.Run("TestCmpInverse", rapid.MakeCheck(testCmpInverse))
	t.Run("TestEqualCommutative", rapid.MakeCheck(testEqualCommutative))

	// Properties about tests on a single Dec
	t.Run("TestIsNegative", rapid.MakeCheck(testIsNegative))

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

	res, err := two.Add(zero)
	require.NoError(t, err)
	require.True(t, res.IsEqual(two))

	res, err = five.Sub(two)
	require.NoError(t, err)
	require.True(t, res.IsEqual(three))

	res, err = onePointOneFive.Add(twoPointThreeFour)
	require.NoError(t, err)
	require.True(t, res.IsEqual(threePointFourNine))

	res, err = threePointFourNine.Sub(two)
	require.NoError(t, err)
	require.True(t, res.IsEqual(onePointFourNine))

	res, err = minusOne.Sub(four)
	require.NoError(t, err)
	require.True(t, res.IsEqual(minusFivePointZero))

	_, err = four.Quo(zero)
	require.Error(t, err)

	res, err = four.Quo(two)
	require.NoError(t, err)
	require.True(t, res.IsEqual(two))

	require.False(t, zero.IsNegative())
	require.False(t, one.IsNegative())
	require.True(t, minusOne.IsNegative())
}

var genDec *rapid.Generator = rapid.Custom(func(t *rapid.T) Dec {
	f := rapid.Float64().Draw(t, "f").(float64)
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
var genFloatAndDec *rapid.Generator = rapid.Custom(func(t *rapid.T) floatAndDec {
	f := rapid.Float64().Draw(t, "f").(float64)
	dec, err := NewDecFromString(fmt.Sprintf("%g", f))
	require.NoError(t, err)
	return floatAndDec{f, dec}
})

// Property: n == NewDecFromInt64(n).Int64()
func testDecInt64(t *rapid.T) {
	nIn := rapid.Int64().Draw(t, "n").(int64)
	nOut, err := NewDecFromInt64(nIn).Int64()

	require.NoError(t, err)
	require.Equal(t, nIn, nOut)
}

// Property: 0 + a == a
func testAddLeftIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a").(Dec)
	zero := NewDecFromInt64(0)

	b, err := zero.Add(a)
	require.NoError(t, err)

	require.True(t, a.IsEqual(b))
}

// Property: a + 0 == a
func testAddRightIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a").(Dec)
	zero := NewDecFromInt64(0)

	b, err := a.Add(zero)
	require.NoError(t, err)

	require.True(t, a.IsEqual(b))
}

// Property: a + b == b + a
func testAddCommutative(t *rapid.T) {
	a := genDec.Draw(t, "a").(Dec)
	b := genDec.Draw(t, "b").(Dec)

	c, err := a.Add(b)
	require.NoError(t, err)

	d, err := b.Add(a)
	require.NoError(t, err)

	require.True(t, c.IsEqual(d))
}

// Property: (a + b) + c == a + (b + c)
func testAddAssociative(t *rapid.T) {
	a := genDec.Draw(t, "a").(Dec)
	b := genDec.Draw(t, "b").(Dec)
	c := genDec.Draw(t, "c").(Dec)

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

	require.True(t, e.IsEqual(g))
}

// Property: a - 0 == a
func testSubRightIdentity(t *rapid.T) {
	a := genDec.Draw(t, "a").(Dec)
	zero := NewDecFromInt64(0)

	b, err := a.Sub(zero)
	require.NoError(t, err)

	require.True(t, a.IsEqual(b))
}

// Property: a - a == 0
func testSubZero(t *rapid.T) {
	a := genDec.Draw(t, "a").(Dec)
	zero := NewDecFromInt64(0)

	b, err := a.Sub(a)
	require.NoError(t, err)

	require.True(t, b.IsEqual(zero))
}

// Property: (a - b) + b == a
func testSubAdd(t *rapid.T) {
	a := genDec.Draw(t, "a").(Dec)
	b := genDec.Draw(t, "b").(Dec)

	c, err := a.Sub(b)
	require.NoError(t, err)

	d, err := c.Add(b)
	require.NoError(t, err)

	require.True(t, a.IsEqual(d))
}

// Property: (a + b) - b == a
func testAddSub(t *rapid.T) {
	a := genDec.Draw(t, "a").(Dec)
	b := genDec.Draw(t, "b").(Dec)

	c, err := a.Add(b)
	require.NoError(t, err)

	d, err := c.Sub(b)
	require.NoError(t, err)

	require.True(t, a.IsEqual(d))
}

// Property: Cmp(a, b) == -Cmp(b, a)
func testCmpInverse(t *rapid.T) {
	a := genDec.Draw(t, "a").(Dec)
	b := genDec.Draw(t, "b").(Dec)

	require.Equal(t, a.Cmp(b), -b.Cmp(a))
}

// Property: IsEqual(a, b) == IsEqual(b, a)
func testEqualCommutative(t *rapid.T) {
	a := genDec.Draw(t, "a").(Dec)
	b := genDec.Draw(t, "b").(Dec)

	require.Equal(t, a.IsEqual(b), b.IsEqual(a))
}

// Property: isNegative(f) == isNegative(NewDecFromString(f.String()))
func testIsNegative(t *rapid.T) {
	floatAndDec := genFloatAndDec.Draw(t, "floatAndDec").(floatAndDec)
	f, dec := floatAndDec.float, floatAndDec.dec

	require.Equal(t, f < 0, dec.IsNegative())
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
	if res := basePlaces - exp; res <= 0 {
		return 0
	} else {
		return uint32(res)
	}
}
