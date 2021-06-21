package module

import (
	"github.com/spf13/cobra"
	"go.uber.org/dig"

	"github.com/cosmos/cosmos-sdk/app/compat"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	types2 "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	_ app.TypeProvider = Module{}
)

type inputs struct {
	dig.In

	Codec            codec.Codec
	KeyProvider      app.KVStoreKeyProvider
	SubspaceProvider types2.SubspaceProvider
}

type outputs struct {
	dig.Out

	Handler    app.Handler `group:"tx"`
	ViewKeeper types.ViewKeeper
	Keeper     types.Keeper `security-role:"admin"`
}

type cliCommands struct {
	dig.Out

	TxCmd    *cobra.Command   `group:"tx"`
	QueryCmd []*cobra.Command `group:"query,flatten"`
}

func (m Module) RegisterTypes(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (Module) ProvideAccountRetriever() client.AccountRetriever {
	return types.AccountRetriever{}
}

func (Module) ProvideCLICommands() cliCommands {
	am := auth.AppModuleBasic{}
	return cliCommands{
		TxCmd: am.GetTxCmd(),
		QueryCmd: []*cobra.Command{
			am.GetQueryCmd(),
			authcmd.GetAccountCmd(),
		},
	}
}

func (m Module) ProvideAppHandler(key app.ModuleKey, inputs inputs) (outputs, error) {
	var accCtr types.AccountConstructor
	if m.AccountConstructor != nil {
		err := inputs.Codec.UnpackAny(m.AccountConstructor, &accCtr)
		if err != nil {
			return outputs{}, err
		}
	} else {
		accCtr = DefaultAccountConstructor{}
	}

	perms := map[string][]string{}
	for _, perm := range m.Permissions {
		perms[perm.Address] = perm.Permissions
	}

	var randomGenesisAccountsProvider types.RandomGenesisAccountsProvider
	if m.RandomGenesisAccountsProvider != nil {
		err := inputs.Codec.UnpackAny(m.RandomGenesisAccountsProvider, &randomGenesisAccountsProvider)
		if err != nil {
			return outputs{}, err
		}
	} else {
		randomGenesisAccountsProvider = DefaultRandomGenesisAccountsProvider{}
	}

	keeper := authkeeper.NewAccountKeeper(
		inputs.Codec,
		inputs.KeyProvider(key),
		inputs.SubspaceProvider(key),
		func() types.AccountI {
			return accCtr.NewAccount()
		},
		perms,
	)
	appMod := auth.NewAppModule(inputs.Codec, keeper, func(simState *module.SimulationState) types.GenesisAccounts {
		return randomGenesisAccountsProvider.RandomGenesisAccounts(simState)
	})

	return outputs{
		ViewKeeper: viewOnlyKeeper{keeper},
		Keeper:     keeper,
		Handler:    compat.AppModuleHandler(key.ID(), appMod),
	}, nil
}

func (m DefaultAccountConstructor) NewAccount() types.AccountI {
	return &types.BaseAccount{}
}

func (m DefaultRandomGenesisAccountsProvider) RandomGenesisAccounts(simState *module.SimulationState) types.GenesisAccounts {
	return simulation.RandomGenesisAccounts(simState)
}
