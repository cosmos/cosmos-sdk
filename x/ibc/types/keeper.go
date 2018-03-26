package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Sender interface {
	Push(sdk.Context, Payload, string)
}
