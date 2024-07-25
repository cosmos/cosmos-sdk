package schema

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestKind_Validate(t *testing.T) {
	for kind := InvalidKind + 1; kind <= MAX_VALID_KIND; kind++ {
		if err := kind.Validate(); err != nil {
			t.Errorf("expected valid kind %s to pass validation, got: %v", kind, err)
		}
	}

	invalidKinds := []Kind{
		Kind(-1),
		InvalidKind,
		Kind(100),
	}

	for _, kind := range invalidKinds {
		if err := kind.Validate(); err == nil {
			t.Errorf("expected invalid kind %s to fail validation, got: %v", kind, err)
		}
	}
}

func TestKind_ValidateValueType(t *testing.T) {
	tests := []struct {
		kind  Kind
		value interface{}
		valid bool
	}{
		{kind: StringKind, value: "hello", valid: true},
		{kind: StringKind, value: []byte("hello"), valid: false},
		{kind: BytesKind, value: []byte("hello"), valid: true},
		{kind: BytesKind, value: "hello", valid: false},
		{kind: Int8Kind, value: int8(1), valid: true},
		{kind: Int8Kind, value: int16(1), valid: false},
		{kind: Uint8Kind, value: uint8(1), valid: true},
		{kind: Uint8Kind, value: uint16(1), valid: false},
		{kind: Int16Kind, value: int16(1), valid: true},
		{kind: Int16Kind, value: int32(1), valid: false},
		{kind: Uint16Kind, value: uint16(1), valid: true},
		{kind: Uint16Kind, value: uint32(1), valid: false},
		{kind: Int32Kind, value: int32(1), valid: true},
		{kind: Int32Kind, value: int64(1), valid: false},
		{kind: Uint32Kind, value: uint32(1), valid: true},
		{kind: Uint32Kind, value: uint64(1), valid: false},
		{kind: Int64Kind, value: int64(1), valid: true},
		{kind: Int64Kind, value: int32(1), valid: false},
		{kind: Uint64Kind, value: uint64(1), valid: true},
		{kind: Uint64Kind, value: uint32(1), valid: false},
		{kind: IntegerStringKind, value: "1", valid: true},
		{kind: IntegerStringKind, value: int32(1), valid: false},
		{kind: DecimalStringKind, value: "1.0", valid: true},
		{kind: DecimalStringKind, value: "1", valid: true},
		{kind: DecimalStringKind, value: "1.1e4", valid: true},
		{kind: DecimalStringKind, value: int32(1), valid: false},
		{kind: AddressKind, value: []byte("hello"), valid: true},
		{kind: AddressKind, value: 1, valid: false},
		{kind: BoolKind, value: true, valid: true},
		{kind: BoolKind, value: false, valid: true},
		{kind: BoolKind, value: 1, valid: false},
		{kind: EnumKind, value: "hello", valid: true},
		{kind: EnumKind, value: 1, valid: false},
		{kind: TimeKind, value: time.Now(), valid: true},
		{kind: TimeKind, value: "hello", valid: false},
		{kind: DurationKind, value: time.Second, valid: true},
		{kind: DurationKind, value: "hello", valid: false},
		{kind: Float32Kind, value: float32(1.0), valid: true},
		{kind: Float32Kind, value: float64(1.0), valid: false},
		{kind: Float64Kind, value: float64(1.0), valid: true},
		{kind: Float64Kind, value: float32(1.0), valid: false},
		{kind: JSONKind, value: json.RawMessage("{}"), valid: true},
		{kind: JSONKind, value: "hello", valid: false},
		{kind: InvalidKind, value: "hello", valid: false},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			err := tt.kind.ValidateValueType(tt.value)
			if tt.valid && err != nil {
				t.Errorf("test %d: expected valid value %v for kind %s to pass validation, got: %v", i, tt.value, tt.kind, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("test %d: expected invalid value %v for kind %s to fail validation, got: %v", i, tt.value, tt.kind, err)
			}
		})
	}

	// nils get rejected
	for kind := InvalidKind + 1; kind <= MAX_VALID_KIND; kind++ {
		if err := kind.ValidateValueType(nil); err == nil {
			t.Errorf("expected nil value to fail validation for kind %s", kind)
		}
	}
}

func TestKind_ValidateValue(t *testing.T) {
	tests := []struct {
		kind  Kind
		value interface{}
		valid bool
	}{
		// check a few basic cases that should get caught be ValidateValueType
		{StringKind, "hello", true},
		{Int64Kind, int64(1), true},
		{Int32Kind, "abc", false},
		{BytesKind, nil, false},
		// string must be valid UTF-8
		{StringKind, string([]byte{0xff, 0xfe, 0xfd}), false},
		// strings with null characters are invalid
		{StringKind, string([]byte{1, 2, 0, 3}), false},
		// check integer, decimal and json more thoroughly
		{IntegerStringKind, "1", true},
		{IntegerStringKind, "0", true},
		{IntegerStringKind, "10", true},
		{IntegerStringKind, "-100", true},
		{IntegerStringKind, "1.0", false},
		{IntegerStringKind, "00", true}, // leading zeros are allowed
		{IntegerStringKind, "001", true},
		{IntegerStringKind, "-01", true},
		// 100 digits
		{IntegerStringKind, "1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", true},
		// more than 100 digits
		{IntegerStringKind, "10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", false},
		{IntegerStringKind, "", false},
		{IntegerStringKind, "abc", false},
		{IntegerStringKind, "abc100", false},
		{DecimalStringKind, "1.0", true},
		{DecimalStringKind, "0.0", true},
		{DecimalStringKind, "-100.075", true},
		{DecimalStringKind, "1002346.000", true},
		{DecimalStringKind, "0", true},
		{DecimalStringKind, "10", true},
		{DecimalStringKind, "-100", true},
		{DecimalStringKind, "1", true},
		{DecimalStringKind, "1.0e4", true},
		{DecimalStringKind, "1.0e-4", true},
		{DecimalStringKind, "1.0e+4", true},
		{DecimalStringKind, "1.0e", false},
		{DecimalStringKind, "1.0e4.0", false},
		{DecimalStringKind, "1.0e-4.0", false},
		{DecimalStringKind, "1.0e+4.0", false},
		{DecimalStringKind, "-1.0e-4", true},
		{DecimalStringKind, "-1.0e+4", true},
		{DecimalStringKind, "-1.0E4", true},
		{DecimalStringKind, "1E-9", true},
		{DecimalStringKind, "1E-99", true},
		{DecimalStringKind, "1E+9", true},
		{DecimalStringKind, "1E+99", true},
		// 50 digits before and after the decimal point
		{DecimalStringKind, "10000000000000000000000000000000000000000000000000.10000000000000000000000000000000000000000000000001", true},
		// too many digits before the decimal point
		{DecimalStringKind, "10000000000000000000000000000000000000000000000000000000000000000000000000", false},
		// too many digits after the decimal point
		{DecimalStringKind, "1.0000000000000000000000000000000000000000000000000000000000000000000000001", false},
		// exponent too big
		{DecimalStringKind, "1E-999", false},
		{DecimalStringKind, "", false},
		{DecimalStringKind, "abc", false},
		{DecimalStringKind, "abc", false},
		{JSONKind, json.RawMessage(`{"a":10}`), true},
		{JSONKind, json.RawMessage("10"), true},
		{JSONKind, json.RawMessage("10.0"), true},
		{JSONKind, json.RawMessage("true"), true},
		{JSONKind, json.RawMessage("null"), true},
		{JSONKind, json.RawMessage(`"abc"`), true},
		{JSONKind, json.RawMessage(`[1,true,0.1,"abc",{"b":3}]`), true},
		{JSONKind, json.RawMessage(`"abc`), false},
		{JSONKind, json.RawMessage(`tru`), false},
		{JSONKind, json.RawMessage(`[`), false},
		{JSONKind, json.RawMessage(`{`), false},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %v %s", tt.kind, tt.value), func(t *testing.T) {
			err := tt.kind.ValidateValue(tt.value)
			if tt.valid && err != nil {
				t.Errorf("test %d: expected valid value %v for kind %s to pass validation, got: %v", i, tt.value, tt.kind, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("test %d: expected invalid value %v for kind %s to fail validation, got: %v", i, tt.value, tt.kind, err)
			}
		})
	}
}

func TestKind_String(t *testing.T) {
	tests := []struct {
		kind Kind
		want string
	}{
		{StringKind, "string"},
		{BytesKind, "bytes"},
		{Int8Kind, "int8"},
		{Uint8Kind, "uint8"},
		{Int16Kind, "int16"},
		{Uint16Kind, "uint16"},
		{Int32Kind, "int32"},
		{Uint32Kind, "uint32"},
		{Int64Kind, "int64"},
		{Uint64Kind, "uint64"},
		{IntegerStringKind, "integer"},
		{DecimalStringKind, "decimal"},
		{BoolKind, "bool"},
		{TimeKind, "time"},
		{DurationKind, "duration"},
		{Float32Kind, "float32"},
		{Float64Kind, "float64"},
		{JSONKind, "json"},
		{EnumKind, "enum"},
		{AddressKind, "address"},
		{InvalidKind, "invalid(0)"},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %s", tt.kind), func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Errorf("test %d: Kind.String() = %v, want %v", i, got, tt.want)
			}
		})
	}
}

func TestKindForGoValue(t *testing.T) {
	tests := []struct {
		value interface{}
		want  Kind
	}{
		{"hello", StringKind},
		{[]byte("hello"), BytesKind},
		{int8(1), Int8Kind},
		{uint8(1), Uint8Kind},
		{int16(1), Int16Kind},
		{uint16(1), Uint16Kind},
		{int32(1), Int32Kind},
		{uint32(1), Uint32Kind},
		{int64(1), Int64Kind},
		{uint64(1), Uint64Kind},
		{float32(1.0), Float32Kind},
		{float64(1.0), Float64Kind},
		{true, BoolKind},
		{time.Now(), TimeKind},
		{time.Second, DurationKind},
		{json.RawMessage("{}"), JSONKind},
		{map[string]interface{}{"a": 1}, InvalidKind},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			if got := KindForGoValue(tt.value); got != tt.want {
				t.Errorf("test %d: KindForGoValue(%v) = %v, want %v", i, tt.value, got, tt.want)
			}

			// for valid kinds check valid value
			if tt.want.Validate() == nil {
				if err := tt.want.ValidateValue(tt.value); err != nil {
					t.Errorf("test %d: expected valid value %v for kind %s to pass validation, got: %v", i, tt.value, tt.want, err)
				}
			}
		})
	}
}
