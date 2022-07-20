package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type UpgradeAuthority sdk.AccAddress

func (a UpgradeAuthority) String() string {
	return sdk.AccAddress(a).String()
}
