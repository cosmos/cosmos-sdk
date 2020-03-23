package exported

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supply "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// SupplyKeeper defines the expected supply Keeper (noalias)
type SupplyKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	GetModuleAccount(ctx sdk.Context, moduleName string) supply.ModuleAccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// SupplyKeeper defines the expected auth Account Keeper (noalias)
type AccountKeeper interface {
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) auth.Account
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) auth.Account
	SetAccount(ctx sdk.Context, acc auth.Account)
}

type MsgGrantFeeAllowance interface {
	sdk.Msg

	GetFeeGrant() *FeeAllowance
	GetGranter() sdk.AccAddress
	GetGrantee() sdk.AccAddress
	PrepareForExport(time.Time, int64) FeeAllowanceGrant
}

type FeeAllowanceGrant interface {
	GetFeeGrant() *FeeAllowance
	GetGranter() sdk.AccAddress
	GetGrantee() sdk.AccAddress

	ValidateBasic() error
	PrepareForExport(time.Time, int64) FeeAllowanceGrant
}
