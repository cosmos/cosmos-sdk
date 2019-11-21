package exported

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// PacketDataI defines the standard interface for IBC packet data
type PacketDataI interface {
	GetTimeoutHeight() uint64 // indicates a consensus height on the destination chain after which the Packet will no longer be processed, and will instead count as having timed-out.
	GetData() []byte          // opaque value which can be defined by the application logic of the associated modules.
	ValidateBasic() error

	GetSourcePort() string      // identifies the port on the sending chain.
	GetDestinationPort() string // identifies the port on the receiving chain.

	Route() string
	Type() string
}

type Packet = types.Packet
