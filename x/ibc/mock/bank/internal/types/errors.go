//nolint
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ibcmockbank errors reserve 100 ~ 199.
const (
	CodeErrGetSequence sdk.CodeType = 100
	CodeErrSendPacket  sdk.CodeType = 101
	CodeInvalidAmount  sdk.CodeType = 102
	CodeInvalidAddress sdk.CodeType = 103
)
