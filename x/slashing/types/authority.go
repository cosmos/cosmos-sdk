package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Authority sdk.AccAddress

func (a Authority) String() string {
	return sdk.AccAddress(a).String()
}
