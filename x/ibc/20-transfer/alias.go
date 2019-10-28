package transfer

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

type (
	MsgTransfer        = types.MsgTransfer
	TransferPacketData = types.TransferPacketData
	Keeper             = keeper.Keeper
)

const (
	SubModuleName = types.SubModuleName
	StoreKey      = types.StoreKey
	QuerierRoute  = types.QuerierRoute
	RouterKey     = types.RouterKey
)

var (
	RegisterCdc = types.RegisterCodec

	NewKeeper      = keeper.NewKeeper
	NewMsgTransfer = types.NewMsgTransfer
)
