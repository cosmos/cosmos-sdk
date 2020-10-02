package gov

import (
	v1 "github.com/cosmos/cosmos-sdk/types/module/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type Module struct{}

var _ v1.Module = Module{}

func (m Module) Init(configurator v1.Configurator) {
	types.RegisterMsgServer(configurator.MsgServer(), keeper.NewMsgServer())
}
