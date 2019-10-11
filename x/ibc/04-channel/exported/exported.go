package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PacketI interface {
	Sequence() uint64
	TimeoutHeight() uint64
	SourcePort() string
	SourceChannel() string
	DestPort() string
	DestChannel() string
	Data() []byte

	// Non ICS04 interface functions
	Type() string
	ValidateBasic() sdk.Error
	Marshal() []byte // Should exclude PortID/ChannelID info
}
