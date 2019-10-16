package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ibcmockbank errors reserve 100 ~ 199.
const (
	CodeInvalidAmount    sdk.CodeType = 101
	CodeInvalidAddress   sdk.CodeType = 102
	CodeErrGetSequence   sdk.CodeType = 103
	CodeErrSendPacket    sdk.CodeType = 104
	CodeErrReceivePacket sdk.CodeType = 105
	CodeInvalidChannel   sdk.CodeType = 106
)
