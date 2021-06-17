package module

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"go.uber.org/dig"
)

var _ app.Provisioner = Module{}

type Outputs struct {
	dig.Out

	GenesisBalancesIterator genutiltypes.GenesisBalancesIterator
}

func (m Module) Provision(key app.ModuleKey, registrar container.Registrar) error {
	return registrar.Provide(func() Outputs {
		return Outputs{
			GenesisBalancesIterator: types.GenesisBalancesIterator{},
		}
	})
}
