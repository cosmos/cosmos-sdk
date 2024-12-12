package v1

import (
	"context"

	"cosmossdk.io/math"
)

type AssetAccountI interface {
	GetDenom(ctx context.Context) (string, error)
	GetOwner(ctx context.Context) ([]byte, error)
	GetBalance(ctx context.Context, addr []byte) math.Int
	GetSupply(ctx context.Context) math.Int
	SetBalance(ctx context.Context, addr []byte, amt math.Int) error
	SetSupply(ctx context.Context, supply math.Int) error
	SubUnlockedCoins(ctx context.Context, addr []byte, amt math.Int) error
	AddCoins(ctx context.Context, addr []byte, amt math.Int) error
}

type (
	SendFunc = func(ctx context.Context, from, to []byte, amount math.Int) ([][]byte, error)
	MintFunc = func(ctx context.Context, to []byte, amount math.Int) ([][]byte, error)
	BurnFunc = func(ctx context.Context, from []byte, amount math.Int) ([][]byte, error)
)