package exported

import "github.com/cosmos/cosmos-sdk/x/ibc/common"

// ChannelI defines the standard interface for a channel end.
type ChannelI interface {
	GetState() common.State
	GetOrdering() common.Order
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
	GetTimeoutTimestamp() uint64
	GetSourcePort() string
	GetSourceChannel() string
	GetDestPort() string
	GetDestChannel() string
	GetData() []byte
	ValidateBasic() error
}
