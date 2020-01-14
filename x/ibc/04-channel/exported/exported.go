package exported

// PacketI defines the standard interface for IBC packets
type PacketI interface {
	GetSequence() uint64
	GetTimeoutHeight() uint64
	GetSourcePort() string
	GetSourceChannel() string
	GetDestPort() string
	GetDestChannel() string
	GetData() PacketDataI
	ValidateBasic() error
}

// PacketDataI defines the packet data interface for IBC packets
// IBC application modules should define which data they want to
// send and receive over IBC channels.
type PacketDataI interface {
	GetBytes() []byte         // GetBytes returns the serialised packet data (without timeout)
	GetTimeoutHeight() uint64 // GetTimeoutHeight returns the timeout height defined specifically for each packet data type instance

	ValidateBasic() error // ValidateBasic validates basic properties of the packet data, implements sdk.Msg
	Type() string         // Type returns human readable identifier, implements sdk.Msg
}
