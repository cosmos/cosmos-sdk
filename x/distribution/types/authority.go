package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type DistrAuthority sdk.AccAddress

func (a DistrAuthority) String() string {
	return sdk.AccAddress(a).String()
}
