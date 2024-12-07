package v1

import (
	"context"

	"cosmossdk.io/math"
)

type MsgInitAssetAccountWrapper struct {
	MsgInitAssetAccount
	TransferFunc func(aa AssetAccountI) sendFunc
	MintFunc     func(aa AssetAccountI) mintFunc
	BurnFunc     func(aa AssetAccountI) burnFunc
}

type (
	sendFunc = func(ctx context.Context, from, to []byte, amount math.Int) ([][]byte, error)
	mintFunc = func(ctx context.Context, to []byte, amount math.Int) ([][]byte, error)
	burnFunc = func(ctx context.Context, from []byte, amount math.Int) ([][]byte, error)
)

type AssetAccountI interface {
	GetDenom(ctx context.Context) (string, error)
	GetOwner(ctx context.Context) ([]byte, error)
	GetBalance(ctx context.Context, addr []byte) math.Int
	GetSupply(ctx context.Context) math.Int
	SetBalance(ctx context.Context, addr []byte, amt math.Int) error
	SubUnlockedCoins(ctx context.Context, addr []byte, amt math.Int) error
	AddCoins(ctx context.Context, addr []byte, amt math.Int) error
}
