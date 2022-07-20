package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankAuthority sdk.AccAddress

func (a BankAuthority) String() string {
	return sdk.AccAddress(a).String()
}
