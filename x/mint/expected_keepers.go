package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// StakingKeeper defines the expected staking keeper
type StakingKeeper interface {
	StakingTokenSupply(ctx sdk.Context) sdk.Int // TODO: delete this and use the account
	BondedRatio(ctx sdk.Context) sdk.Dec
}

// SupplyKeeper defines the expected supply keeper
type SupplyKeeper interface {
	Inflate(ctx sdk.Context, amount sdk.Coins)

	GetPoolAccountByName(ctx sdk.Context, name string) (supply.PoolAccount, sdk.Error)
	SetPoolAccount(ctx sdk.Context, macc supply.PoolAccount)

	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsPoolToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	SendCoinsPoolToPool(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) sdk.Error
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error
}
