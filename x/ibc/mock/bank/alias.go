package mockbank

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/mock/bank/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/mock/bank/internal/types"
)

// nolint
type (
	MsgTransfer           = types.MsgTransfer
	MsgRecvTransferPacket = types.MsgRecvTransferPacket
	Keeper                = keeper.Keeper
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	TStoreKey    = types.TStoreKey
	QuerierRoute = types.QuerierRoute
	RouterKey    = types.RouterKey
)

// nolint
var (
	RegisterCdc = types.RegisterCodec

	NewKeeper                = keeper.NewKeeper
	NewMsgTransfer           = types.NewMsgTransfer
	NewMsgRecvTransferPacket = types.NewMsgRecvTransferPacket
)
