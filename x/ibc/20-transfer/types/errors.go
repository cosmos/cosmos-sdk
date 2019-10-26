package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// transfer error codes
const (
	DefaultCodespace sdk.CodespaceType = SubModuleName

	CodeIncorrectDenom    sdk.CodeType = 101
	CodeInvalidAmount     sdk.CodeType = 102
	CodeInvalidAddress    sdk.CodeType = 103
	CodeInvalidReceiver   sdk.CodeType = 104
	CodeErrSendPacket     sdk.CodeType = 105
	CodeInvalidPacketData sdk.CodeType = 106
)
