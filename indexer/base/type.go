package indexerbase

type Type int

const (
	// TypeString is a string type and values of this type must be of the go type string.
	TypeString Type = iota

	// TypeBytes is a bytes type and values of this type must be of the go type []byte.
	TypeBytes

	// TypeInt8 is an int8 type and values of this type must be of the go type int8.
	TypeInt8

	// TypeInt16 is an int16 type and values of this type must be of the go type int16.
	TypeInt16

	// TypeInt32 is an int32 type and values of this type must be of the go type int32.
	TypeInt32

	// TypeInt64 is an int64 type and values of this type must be of the go type int64.
	TypeInt64

	// TypeDecimal represents an arbitrary precision decimal or integer number. Values of this type
	// must be of the go type string formatted as decimal numbers. Exponential E-notation is supported
	// but NaN and Infinity are not.
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

	// TypeEnum is an enum type and values of this type must be of the go type string. Columns of this type are
	// expected to set the EnumDefinition field in the column definition to the enum definition.
	TypeEnum

	// TypeJSON is a JSON type and values of this type must be of the go type json.RawMessage.
	TypeJSON
)
