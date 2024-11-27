package v1

import (
	"context"

	"cosmossdk.io/math"
)

type MsgInitAssetAccountWrapper struct {
	MsgInitAssetAccount
	TransferFunc func(aa AssetAccountI) sendFunc
}

type sendFunc = func(ctx context.Context, from, to []byte, amount math.Int) error

type AssetAccountI interface {
	GetDenom(ctx context.Context) (string, error)
	GetBalance(ctx context.Context, addr []byte) math.Int
	GetSupply(ctx context.Context) (math.Int, error)
	SetBalance(ctx context.Context, addr []byte, amt math.Int) error
	SubUnlockedCoins(ctx context.Context, addr []byte, amt math.Int) error
	AddCoins(ctx context.Context, addr []byte, amt math.Int) error
}