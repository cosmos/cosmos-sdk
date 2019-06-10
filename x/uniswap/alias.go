package uniswap

import (
	"github.com/cosmos/cosmos-sdk/x/uniswap/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/uniswap/internal/types"
)

type (
	Keeper             = keeper.Keeper
	NativeAsset        = types.NativeAsset
	MsgSwapOrder       = types.MsgSwapOrder
	MsgAddLiquidity    = types.MsgAddLiquidity
	MsgRemoveLiquidity = types.MsgRemoveLiquidity
)

const (
	ModuleName = types.ModuleName
)
