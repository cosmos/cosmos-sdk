package ante

import (
	"context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	feemarkettypes "cosmossdk.io/x/feemarket/types"
)

// AccountKeeper defines the contract needed for AccountKeeper related APIs.
// Interface provides support to use non-sdk AccountKeeper for AnteHandler's decorators.
type AccountKeeper interface {
	GetParams(ctx context.Context) (params authtypes.Params)
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	AddressCodec() address.Codec
}

// FeeGrantKeeper defines the expected feegrant keeper.
type FeeGrantKeeper interface {
	UseGrantedFees(ctx context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}

// BankKeeper defines the contract needed for supply related APIs.
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// FeeMarketKeeper defines the expected feemarket keeper.
type FeeMarketKeeper interface {
	GetState(ctx sdk.Context) (feemarkettypes.State, error)
	GetMinGasPrice(ctx sdk.Context, denom string) (sdk.DecCoin, error)
	GetParams(ctx sdk.Context) (feemarkettypes.Params, error)
	SetState(ctx sdk.Context, state feemarkettypes.State) error
	SetParams(ctx sdk.Context, params feemarkettypes.Params) error
	ResolveToDenom(ctx sdk.Context, coin sdk.DecCoin, denom string) (sdk.DecCoin, error)
}
