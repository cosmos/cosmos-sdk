package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type PacketI interface {
	GetSequence() uint64
	GetTimeoutHeight() uint64
	GetSourcePort() string
	GetSourceChannel() string
	GetDestPort() string
	GetDestChannel() string
	GetData() []byte
	ValidateBasic() sdk.Error
}
