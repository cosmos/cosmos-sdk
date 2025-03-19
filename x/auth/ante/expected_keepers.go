package ante

import (
	"context"
	"time"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the contract needed for AccountKeeper related APIs.
// Interface provides support to use non-sdk AccountKeeper for AnteHandler's decorators.
type AccountKeeper interface {
	GetParams(ctx context.Context) (params types.Params)
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	GetModuleAddress(moduleName string) sdk.AccAddress
	AddressCodec() address.Codec
}

// UnorderedSequenceManager defines the contract needed for UnorderedSequence management.
type UnorderedSequenceManager interface {
	RemoveExpiredUnorderedSequences(ctx sdk.Context) error
	AddUnorderedSequence(ctx sdk.Context, sender []byte, timestamp time.Time) error
	ContainsUnorderedSequence(ctx sdk.Context, sender []byte, timestamp time.Time) (bool, error)
}

// FeegrantKeeper defines the expected feegrant keeper.
type FeegrantKeeper interface {
	UseGrantedFees(ctx context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}
