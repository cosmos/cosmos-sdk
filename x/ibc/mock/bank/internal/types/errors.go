package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ibcmockbank errors reserve 100 ~ 199.
const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeInvalidAmount     sdk.CodeType = 101
	CodeInvalidAddress    sdk.CodeType = 102
	CodeErrReceivePacket  sdk.CodeType = 103
	CodeProofMissing      sdk.CodeType = 104
	CodeInvalidPacketData sdk.CodeType = 105
)
