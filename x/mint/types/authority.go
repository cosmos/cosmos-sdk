package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MintAuthority sdk.AccAddress

func (a MintAuthority) String() string {
	return sdk.AccAddress(a).String()
}
