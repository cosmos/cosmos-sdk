package exported

import (
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// ChannelI defines the standard interface for a channel end.
type ChannelI interface {
	GetState() ibctypes.State
	GetOrdering() ibctypes.Order
	GetCounterparty() CounterpartyI
	GetConnectionHops() []string
	GetVersion() string
	ValidateBasic() error
}

// CounterpartyI defines the standard interface for a channel end's
// counterparty.
type CounterpartyI interface {
	GetPortID() string
	GetChannelID() string
	ValidateBasic() error
}

// PacketI defines the standard interface for IBC packets
type PacketI interface {
	GetSequence() uint64
	GetTimeoutHeight() uint64
	GetSourcePort() string
	GetSourceChannel() string
	GetDestinationPort() string
	GetDestinationChannel() string
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

// PacketAcknowledgementI defines the interface for IBC packet acknowledgements.
type PacketAcknowledgementI interface {
	GetBytes() []byte
}
