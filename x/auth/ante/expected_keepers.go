package ante

import (
	"context"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
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

// FeegrantKeeper defines the expected feegrant keeper.
type FeegrantKeeper interface {
	UseGrantedFees(ctx context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}

type AuthzKeeper interface {
	GetAuthzOptions() map[string]map[string]string
	GetAuthzRulesKeys(ctx context.Context) (map[string]interface{}, error)
	GetAuthzWithRules(ctx context.Context, grantee, granter sdk.AccAddress, msgType string) (authz.Authorization, map[string]interface{})
	AddressCodec() address.Codec
}
