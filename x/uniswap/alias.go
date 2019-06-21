package uniswap

import (
	"github.com/cosmos/cosmos-sdk/x/uniswap/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/uniswap/internal/types"
)

type (
	Keeper             = keeper.Keeper
	MsgSwapOrder       = types.MsgSwapOrder
	MsgAddLiquidity    = types.MsgAddLiquidity
	MsgRemoveLiquidity = types.MsgRemoveLiquidity
)

var (
	ErrInvalidDeadline = types.ErrInvalidDeadline
	ErrNotPositive     = types.ErrNotPositive
)

const (
	DefaultCodespace = types.DefaultCodespace
	ModuleName       = types.ModuleName
)
