package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ibcmockbank errors reserve 100 ~ 199.
const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeInvalidAddress    sdk.CodeType = 101
	CodeErrReceivePacket  sdk.CodeType = 102
	CodeProofMissing      sdk.CodeType = 103
	CodeInvalidPacketData sdk.CodeType = 104
)
