package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// BankKeeper defines the needed bank keeper interface
// Interface provides support to use non-sdk bank keeper for AnteHandler's decorators.
type BankKeeper interface {
	GetParams(ctx sdk.Context) (params types.Params)
}

// TokenKeeper defines the expected token keeper interface
type TokenKeeper interface {
	GetOwner(ctx sdk.Context, denom string) (sdk.AccAddress, error)
}
