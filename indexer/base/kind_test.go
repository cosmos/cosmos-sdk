package indexerbase

import (
	"strings"
	"testing"
	"time"
)

func TestKind_Validate(t *testing.T) {
	validKinds := []Kind{
		StringKind,
		BytesKind,
		Int8Kind,
		Uint8Kind,
		Int16Kind,
		Uint16Kind,
		Int32Kind,
		Uint32Kind,
		Int64Kind,
		Uint64Kind,
		IntegerKind,
		DecimalKind,
		BoolKind,
		EnumKind,
		Bech32AddressKind,
	}

	for _, kind := range validKinds {
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

func TestKind_ValidateValue(t *testing.T) {
	tests := []struct {
		kind  Kind
		value any
		valid bool
	}{
		{
			kind:  StringKind,
			value: "hello",
			valid: true,
		},
		{
			kind:  StringKind,
			value: &strings.Builder{},
			valid: true,
		},
		{
			kind:  StringKind,
			value: []byte("hello"),
			valid: false,
		},
		{
			kind:  BytesKind,
			value: []byte("hello"),
			valid: true,
		},
		{
			kind:  BytesKind,
			value: "hello",
			valid: false,
		},
		{
			kind:  Int8Kind,
			value: int8(1),
			valid: true,
		},
		{
			kind:  Int8Kind,
			value: int16(1),
			valid: false,
		},
		{
			kind:  Uint8Kind,
			value: uint8(1),
			valid: true,
		},
		{
			kind:  Uint8Kind,
			value: uint16(1),
			valid: false,
		},
		{
			kind:  Int16Kind,
			value: int16(1),
			valid: true,
		},
		{
			kind:  Int16Kind,
			value: int32(1),
			valid: false,
		},
		{
			kind:  Uint16Kind,
			value: uint16(1),
			valid: true,
		},
		{
			kind:  Uint16Kind,
			value: uint32(1),
			valid: false,
		},
		{
			kind:  Int32Kind,
			value: int32(1),
			valid: true,
		},
		{
			kind:  Int32Kind,
			value: int64(1),
			valid: false,
		},
		{
			kind:  Uint32Kind,
			value: uint32(1),
			valid: true,
		},
		{
			kind:  Uint32Kind,
			value: uint64(1),
			valid: false,
		},
		{
			kind:  Int64Kind,
			value: int64(1),
			valid: true,
		},
		{
			kind:  Int64Kind,
			value: int32(1),
			valid: false,
		},
		{
			kind:  Uint64Kind,
			value: uint64(1),
			valid: true,
		},
		{
			kind:  Uint64Kind,
			value: uint32(1),
			valid: false,
		},
		{
			kind:  IntegerKind,
			value: "1",
			valid: true,
		},
		//{
		//	kind:  IntegerKind,
		//	value: (&strings.Builder{}).WriteString("1"),
		//	valid: true,
		//},
		{
			kind:  IntegerKind,
			value: int32(1),
			valid: false,
		},
		{
			kind:  IntegerKind,
			value: int64(1),
			valid: true,
		},
		{
			kind:  DecimalKind,
			value: "1.0",
			valid: true,
		},
		{
			kind:  DecimalKind,
			value: "1",
			valid: true,
		},
		{
			kind:  DecimalKind,
			value: "1.1e4",
			valid: true,
		},
		//{
		//	kind:  DecimalKind,
		//	value: (&strings.Builder{}).WriteString("1.0"),
		//	valid: true,
		//},
		{
			kind:  DecimalKind,
			value: int32(1),
			valid: false,
		},
		{
			kind:  Bech32AddressKind,
			value: "cosmos1hsk6jryyqjfhp5g7c0nh4n6dd45ygctnxglp5h",
			valid: true,
		},
		//{
		//	kind:  Bech32AddressKind,
		//	value: (&strings.Builder{}).WriteString("cosmos1hsk6jryyqjfhp5g7c0nh4n6dd45ygctnxglp5h"),
		//	valid: true,
		//},
		{
			kind:  Bech32AddressKind,
			value: 1,
			valid: false,
		},
		{
			kind:  BoolKind,
			value: true,
			valid: true,
		},
		{
			kind:  BoolKind,
			value: false,
			valid: true,
		},
		{
			kind:  BoolKind,
			value: 1,
			valid: false,
		},
		{
			kind:  EnumKind,
			value: "hello",
			valid: true,
		},
		//{
		//	kind:  EnumKind,
		//	value: (&strings.Builder{}).WriteString("hello"),
		//	valid: true,
		//},
		{
			kind:  EnumKind,
			value: 1,
			valid: false,
		},
		{
			kind:  TimeKind,
			value: time.Now(),
			valid: true,
		},
		{
			kind:  TimeKind,
			value: "hello",
			valid: false,
		},
		{
			kind:  DurationKind,
			value: time.Second,
			valid: true,
		},
		{
			kind:  DurationKind,
			value: "hello",
			valid: false,
		},
		{
			kind:  Float32Kind,
			value: float32(1.0),
			valid: true,
		},
		{
			kind:  Float32Kind,
			value: float64(1.0),
			valid: false,
		},
		// TODO float64, json
	}

	for i, tt := range tests {
		err := tt.kind.ValidateValueType(tt.value)
		if tt.valid && err != nil {
			t.Errorf("test %d: expected valid value %v for kind %s to pass validation, got: %v", i, tt.value, tt.kind, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("test %d: expected invalid value %v for kind %s to fail validation, got: %v", i, tt.value, tt.kind, err)
		}
	}
}
