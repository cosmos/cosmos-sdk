package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SlashingAuthority sdk.AccAddress

func (a SlashingAuthority) String() string {
	return sdk.AccAddress(a).String()
}
