package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Handler func(sdk.Context, Payload) sdk.Result
