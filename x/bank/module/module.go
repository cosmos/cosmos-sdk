package module

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

var _ app.Provisioner = Module{}

func (m Module) Provision(app.ModuleKey) container.Option {
	return container.Provide(
		func() genutiltypes.GenesisBalancesIterator {
			return types.GenesisBalancesIterator{}
		},
	)
}
