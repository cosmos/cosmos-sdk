package indexerbase

import (
	"encoding/json"
	"fmt"
	"time"
)

type Type int

const (
	TypeUnknown Type = iota

	// TypeString is a string type and values of this type must be of the go type string
	// or implement fmt.Stringer().
	TypeString

	// TypeBytes is a bytes type and values of this type must be of the go type []byte.
	TypeBytes

	// TypeInt8 is an int8 type and values of this type must be of the go type int8.
	TypeInt8

	// TypeUint8 is a uint8 type and values of this type must be of the go type uint8.
	TypeUint8

	// TypeInt16 is an int16 type and values of this type must be of the go type int16.
	TypeInt16

	// TypeUint16 is a uint16 type and values of this type must be of the go type uint16.
	TypeUint16

	// TypeInt32 is an int32 type and values of this type must be of the go type int32.
	TypeInt32

	// TypeUint32 is a uint32 type and values of this type must be of the go type uint32.
	TypeUint32

	// TypeInt64 is an int64 type and values of this type must be of the go type int64.
	TypeInt64

	// TypeUint64 is a uint64 type and values of this type must be of the go type uint64.
	TypeUint64

	// TypeDecimal represents an arbitrary precision decimal or integer number. Values of this type
	// must be of the go type string or a type that implements fmt.Stringer with the resulting string
	// formatted as decimal numbers with an optional fractional part. Exponential E-notation
	// is supported but NaN and Infinity are not.
	TypeDecimal

	// TypeBool is a boolean type and values of this type must be of the go type bool.
	TypeBool

	// TypeTime is a time type and values of this type must be of the go type time.Time.
	TypeTime

	// TypeDuration is a duration type and values of this type must be of the go type time.Duration.
	TypeDuration

	// TypeFloat32 is a float32 type and values of this type must be of the go type float32.
	TypeFloat32

	// TypeFloat64 is a float64 type and values of this type must be of the go type float64.
	TypeFloat64

	// TypeBech32Address is a bech32 address type and values of this type must be of the go type string or []byte.
	// Columns of this type are expected to set the AddressPrefix field in the column definition to the bech32
	// address prefix.
	TypeBech32Address

	// TypeEnum is an enum type and values of this type must be of the go type string or implement fmt.Stringer.
	// Columns of this type are expected to set the EnumDefinition field in the column definition to the enum definition.
	TypeEnum

	// TypeJSON is a JSON type and values of this type can either be of go type json.RawMessage
	// or any type that can be marshaled to JSON using json.Marshal.
	TypeJSON
)

func (t Type) Validate() error {
	if t <= TypeUnknown {
		return fmt.Errorf("unknown type: %d", t)
	}
	if t > TypeJSON {
		return fmt.Errorf("invalid type: %d", t)
	}
	return nil
}

func (t Type) ValidateValue(value any) error {
	switch t {
	case TypeString:
		_, ok := value.(string)
		_, ok2 := value.(fmt.Stringer)
		if !ok && !ok2 {
			return fmt.Errorf("expected string or type that implements fmt.Stringer, got %T", value)
		}
	case TypeBytes:
		_, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("expected []byte, got %T", value)
		}
	case TypeInt8:
		_, ok := value.(int8)
		if !ok {
			return fmt.Errorf("expected int8, got %T", value)
		}
	case TypeUint8:
		_, ok := value.(uint8)
		if !ok {
			return fmt.Errorf("expected uint8, got %T", value)
		}
	case TypeInt16:
		_, ok := value.(int16)
		if !ok {
			return fmt.Errorf("expected int16, got %T", value)
		}
	case TypeUint16:
		_, ok := value.(uint16)
		if !ok {
			return fmt.Errorf("expected uint16, got %T", value)
		}
	case TypeInt32:
		_, ok := value.(int32)
		if !ok {
			return fmt.Errorf("expected int32, got %T", value)
		}
	case TypeUint32:
		_, ok := value.(uint32)
		if !ok {
			return fmt.Errorf("expected uint32, got %T", value)
		}
	case TypeInt64:
		_, ok := value.(int64)
		if !ok {
			return fmt.Errorf("expected int64, got %T", value)
		}
	case TypeUint64:
		_, ok := value.(uint64)
		if !ok {
			return fmt.Errorf("expected uint64, got %T", value)
		}
	case TypeDecimal:
		_, ok := value.(string)
		_, ok2 := value.(fmt.Stringer)
		if !ok && !ok2 {
			return fmt.Errorf("expected string or type that implements fmt.Stringer, got %T", value)
		}
	case TypeBool:
		_, ok := value.(bool)
		if !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	case TypeTime:
		_, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("expected time.Time, got %T", value)
		}
	case TypeDuration:
		_, ok := value.(time.Duration)
		if !ok {
			return fmt.Errorf("expected time.Duration, got %T", value)
		}
	case TypeFloat32:
		_, ok := value.(float32)
		if !ok {
			return fmt.Errorf("expected float32, got %T", value)
		}
	case TypeFloat64:
		_, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected float64, got %T", value)
		}
	case TypeBech32Address:
		_, ok := value.(string)
		_, ok2 := value.([]byte)
		if !ok && !ok2 {
			return fmt.Errorf("expected string or []byte, got %T", value)
		}
	case TypeEnum:
		_, ok := value.(string)
		_, ok2 := value.(fmt.Stringer)
		if !ok && !ok2 {
			return fmt.Errorf("expected string or type that implements fmt.Stringer, got %T", value)
		}
	case TypeJSON:
		return nil
	default:
		return fmt.Errorf("invalid type: %d", t)
	}
	return nil
}

func (t Type) String() string {
	switch t {
	case TypeString:
		return "string"
	case TypeBytes:
		return "bytes"
	case TypeInt8:
		return "int8"
	case TypeUint8:
		return "uint8"
	case TypeInt16:
		return "int16"
	case TypeUint16:
		return "uint16"
	case TypeInt32:
		return "int32"
	case TypeUint32:
		return "uint32"
	case TypeInt64:
		return "int64"
	case TypeUint64:
		return "uint64"
	case TypeDecimal:
		return "decimal"
	case TypeBool:
		return "bool"
	case TypeTime:
		return "time"
	case TypeDuration:
		return "duration"
	case TypeFloat32:
		return "float32"
	case TypeFloat64:
		return "float64"
	case TypeBech32Address:
		return "bech32address"
	case TypeEnum:
		return "enum"
	case TypeJSON:
		return "json"
	default:
		return ""
	}
}

func TypeForGoValue(value any) Type {
	switch value.(type) {
	case string, fmt.Stringer:
		return TypeString
	case []byte:
		return TypeBytes
	case int8:
		return TypeInt8
	case uint8:
		return TypeUint8
	case int16:
		return TypeInt16
	case uint16:
		return TypeUint16
	case int32:
		return TypeInt32
	case uint32:
		return TypeUint32
	case int64:
		return TypeInt64
	case uint64:
		return TypeUint64
	case float32:
		return TypeFloat32
	case float64:
		return TypeFloat64
	case bool:
		return TypeBool
	case time.Time:
		return TypeTime
	case time.Duration:
		return TypeDuration
	case json.RawMessage:
		return TypeJSON
	default:
		return TypeUnknown
	}
}
