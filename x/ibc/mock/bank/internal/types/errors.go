package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ibcmockbank errors reserve 100 ~ 199.
const (
	CodeIncorrectDenom    sdk.CodeType = 101
	CodeInvalidAmount     sdk.CodeType = 102
	CodeInvalidAddress    sdk.CodeType = 103
	CodeInvalidReceiver   sdk.CodeType = 104
	CodeErrSendPacket     sdk.CodeType = 105
	CodeErrReceivePacket  sdk.CodeType = 106
	CodeProofMissing      sdk.CodeType = 107
	CodeInvalidPacketData sdk.CodeType = 108
)
