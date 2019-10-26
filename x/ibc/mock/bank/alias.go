package mockbank

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/mock/bank/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/mock/bank/internal/types"
)

type (
	MsgRecvTransferPacket = types.MsgRecvTransferPacket
	Keeper                = keeper.Keeper
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	QuerierRoute = types.QuerierRoute
	RouterKey    = types.RouterKey
)

// nolint
var (
	RegisterCdc = types.RegisterCodec

	NewKeeper                = keeper.NewKeeper
	NewMsgRecvTransferPacket = types.NewMsgRecvTransferPacket
)
