package math

import "fmt"

func ExampleDec() {
	d := NewDecFromInt64(1) // 1
	fmt.Println(d.String())

	d = NewDecWithExp(-1234, -3) // -1.234
	fmt.Println(d.String())
	d = NewDecWithExp(1234, 0) // 1234
	fmt.Println(d.String())
	d = NewDecWithExp(1234, 1) // 12340
	fmt.Println(d.String())

	// scientific notation
	d, err := NewDecFromString("1.23E+4") // 12300
	if err != nil {
		panic(err)
	}
	fmt.Println(d.String())

	// decimal notation
	d, err = NewDecFromString("1.234")
	if err != nil {
		panic(err)
	}
	fmt.Println(d.String())

	// Output: 1
	// -1.234
	// 1234
	// 12340
	// 12300
	// 1.234
}

func ExampleDec_Add() {
	sum, err := NewDecFromInt64(1).Add(NewDecFromInt64(1)) // 1 + 1 = 2
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	const maxExp = 100_000
	_, err = NewDecWithExp(1, maxExp).Add(NewDecFromInt64(1)) // 1E+1000000 + 1
	if err != nil {
		fmt.Println(err.Error())
	}

	sum, err = NewDecWithExp(1, maxExp).Add(NewDecWithExp(1, maxExp)) // 1E+1000000 + 1E+1000000
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.Text('E'))

	// the max exponent must not be exceeded
	_, err = NewDecWithExp(1, maxExp+1).Add(NewDecFromInt64(1)) // 1E+1000001 + 1
	if err != nil {
		fmt.Println(err.Error())
	}

	const minExp = -100_000
	// same for min exponent
	_, err = NewDecWithExp(1, minExp-1).Add(NewDecFromInt64(1)) // 1E-1000001 + 1
	if err != nil {
		fmt.Println(err.Error())
	}
	// not even by adding 0
	_, err = NewDecWithExp(1, minExp-1).Add(NewDecFromInt64(0)) // 1E-1000001 + 0
	if err != nil {
		fmt.Println(err.Error())
	}

	// Output: 2
	// 2E+100000
	// add: exponent out of range: invalid decimal
	// add: exponent out of range: invalid decimal
	// add: exponent out of range: invalid decimal
}

func ExampleDec_Sub() {
	sum, err := NewDecFromInt64(2).Sub(NewDecFromInt64(1)) // 2 - 1
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	const maxExp = 100_000
	_, err = NewDecWithExp(1, maxExp).Sub(NewDecFromInt64(1)) // 1E+1000000 - 1
	if err != nil {
		fmt.Println(err.Error())
	}

	sum, err = NewDecWithExp(1, maxExp).Sub(NewDecWithExp(1, maxExp)) // 1E+1000000 - 1E+1000000
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.Text('E'))

	// the max exponent must not be exceeded
	_, err = NewDecWithExp(1, maxExp+1).Sub(NewDecFromInt64(1)) // 1E+1000001 - 1
	if err != nil {
		fmt.Println(err.Error())
	}

	const minExp = -100_000
	// same for min exponent
	_, err = NewDecWithExp(1, minExp-1).Sub(NewDecFromInt64(1)) // 1E-1000001 - 1
	if err != nil {
		fmt.Println(err.Error())
	}
	// not even by adding 0
	_, err = NewDecWithExp(1, minExp-1).Sub(NewDecFromInt64(0)) // 1E-1000001 - 0
	if err != nil {
		fmt.Println(err.Error())
	}

	// Output: 1
	// 0E+100000
	// sub: exponent out of range: invalid decimal
	// sub: exponent out of range: invalid decimal
	// sub: exponent out of range: invalid decimal
}

func ExampleDec_Quo() {
	sum, err := NewDecFromInt64(6).Quo(NewDecFromInt64(2)) // 6 / 2
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	sum, err = NewDecFromInt64(7).Quo(NewDecFromInt64(2)) // 7 / 2
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	sum, err = NewDecFromInt64(4).Quo(NewDecFromInt64(9)) // 4 / 9
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	const minExp = -100_000
	sum, err = NewDecWithExp(1, minExp).Quo(NewDecFromInt64(10)) // 1e-100000 / 10
	if err != nil {
		fmt.Println(err.Error())
	}

	sum, err = NewDecFromInt64(1).Quo(NewDecFromInt64(0)) // 1 / 0 -> error
	if err != nil {
		fmt.Println(err.Error())
	}

	// Output: 3.000000000000000000000000000000000
	// 3.500000000000000000000000000000000
	// 0.4444444444444444444444444444444444
	// exponent out of range: invalid decimal
	// division by zero: invalid decimal
}

func ExampleDec_QuoExact() {
	sum, err := NewDecFromInt64(6).QuoExact(NewDecFromInt64(2)) // 6 / 2
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	sum, err = NewDecFromInt64(7).QuoExact(NewDecFromInt64(2)) // 7 / 2
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	sum, err = NewDecFromInt64(4).QuoExact(NewDecFromInt64(9)) // 4 / 9 -> error
	if err != nil {
		fmt.Println(err.Error())
	}

	const minExp = -100_000
	sum, err = NewDecWithExp(1, minExp).QuoExact(NewDecFromInt64(10)) // 1e-100000 / 10 -> error
	if err != nil {
		fmt.Println(err.Error())
	}

	sum, err = NewDecFromInt64(1).QuoExact(NewDecFromInt64(0)) // 1 / 0 -> error
	if err != nil {
		fmt.Println(err.Error())
	}

	// Output: 3.000000000000000000000000000000000
	// 3.500000000000000000000000000000000
	// unexpected rounding
	// exponent out of range: invalid decimal
	// division by zero: invalid decimal
}

func ExampleDec_QuoInteger() {
	sum, err := NewDecFromInt64(6).QuoInteger(NewDecFromInt64(2)) // 6 / 2
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	sum, err = NewDecFromInt64(7).QuoInteger(NewDecFromInt64(2)) // 7 / 2
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	sum, err = NewDecFromInt64(4).QuoInteger(NewDecFromInt64(9)) // 4 / 9 -> error
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	const minExp = -100_000
	sum, err = NewDecWithExp(1, minExp).QuoInteger(NewDecFromInt64(10)) // 1e-100000 / 10 -> 0
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	sum, err = NewDecFromInt64(1).QuoInteger(NewDecFromInt64(0)) // 1 / 0 -> error
	if err != nil {
		fmt.Println(err.Error())
	}

	// Output: 3
	// 3
	// 0
	// 0
	// division by zero: invalid decimal
}

func ExampleDec_Mul() {
	sum, err := NewDecFromInt64(2).Mul(NewDecFromInt64(3)) // 2 * 3
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	sum, err = NewDecWithExp(125, -2).Mul(NewDecFromInt64(2)) // 1.25 * 2
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	const maxExp = 100_000
	sum, err = NewDecWithExp(1, maxExp).Mul(NewDecFromInt64(10)) // 1e100000 * 10 -> err
	if err != nil {
		fmt.Println(err.Error())
	}

	sum, err = NewDecFromInt64(1).Mul(NewDecFromInt64(0)) // 1 * 0
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	// Output: 6
	// 2.50
	// exponent out of range: invalid decimal
	// 0
}

func ExampleDec_MulExact() {
	sum, err := NewDecFromInt64(2).MulExact(NewDecFromInt64(3)) // 2 * 3
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	sum, err = NewDecWithExp(125, -2).MulExact(NewDecFromInt64(2)) // 1.25 * 2
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	const maxExp = 100_000
	sum, err = NewDecWithExp(1, maxExp).MulExact(NewDecFromInt64(10)) // 1e100000 * 10 -> err
	if err != nil {
		fmt.Println(err.Error())
	}
	a, err := NewDecFromString("0.12345678901234567890123456789012345") // 35 digits after the comma
	if err != nil {
		panic(err)
	}
	sum, err = a.MulExact(NewDecFromInt64(1))
	if err != nil {
		fmt.Println(err.Error())
	}

	sum, err = a.MulExact(NewDecFromInt64(0))
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	sum, err = NewDecFromInt64(1).MulExact(NewDecFromInt64(0)) // 1 * 0
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	// Output: 6
	// 2.50
	// exponent out of range: invalid decimal
	// unexpected rounding
	// 0E-35
	// 0
}

func ExampleDec_Modulo() {
	sum, err := NewDecFromInt64(7).Modulo(NewDecFromInt64(3)) // 7 mod 3 = 1
	if err != nil {
		panic(err)
	}
	fmt.Println(sum.String())

	// Output: 1
}
