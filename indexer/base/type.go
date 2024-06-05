package indexerbase

type Type int

const (
	TypeString Type = iota
	TypeBytes
	TypeInt16
	TypeInt32
	TypeInt64
	TypeDecimal
	TypeBool
	TypeTime
	TypeDuration
	TypeFloat32
	TypeFloat64
	TypeAddress
	TypeEnum
	TypeJSON
)
