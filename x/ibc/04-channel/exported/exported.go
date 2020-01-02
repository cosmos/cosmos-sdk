package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PacketDataI defines the standard interface for IBC packets
type PacketI interface {
	GetSequence() uint64
	GetTimeoutHeight() uint64
	GetSourcePort() string
	GetSourceChannel() string
	GetDestPort() string
	GetDestChannel() string
	GetData() []byte
	ValidateBasic() error
}

type PacketDataI interface {
	GetCommitment() []byte    // GetCommitment returns (possibly non-recoverable) commitment bytes from its Data and Timeout
	GetTimeoutHeight() uint64 // GetTimeoutHeight returns the timeout height defined specifically for each packet data type instance

	ValidateBasic() sdk.Error // ValidateBasic validates basic properties of the packet data, implements sdk.Msg
	Type() string             // Type returns human readable identifier, implements sdk.Msg
}
