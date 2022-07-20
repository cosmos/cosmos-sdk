package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CrisisAuthority sdk.AccAddress

func (a CrisisAuthority) String() string {
	return sdk.AccAddress(a).String()
}
