package provider

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/container"

	"github.com/spf13/cobra"
	"go.uber.org/dig"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

type inputs struct {
	dig.In

	DefaultHome             client.DefaultHome
	Handlers                []types.Handler `group:"app"`
	GenesisBalancesIterator types.GenesisBalancesIterator
	TxConfig                client.TxConfig
}

type outputs struct {
	dig.Out

	RootCommands []*cobra.Command `group:"root,flatten"`
}

var Module = container.Provide(
	func(inputs inputs) outputs {
		validateGenesis := func(cdc codec.JSONCodec, txEncCfg client.TxEncodingConfig, genesis map[string]json.RawMessage) error {
			for _, b := range inputs.Handlers {
				if err := b.ValidateGenesis(cdc, txEncCfg, genesis[b.ID.Name()]); err != nil {
					return err
				}
			}
			return nil
		}

		defaultHome := string(inputs.DefaultHome)

		return outputs{
			RootCommands: []*cobra.Command{
				genutilcli.InitCmd(func(cdc codec.JSONCodec) map[string]json.RawMessage {
					genesis := make(map[string]json.RawMessage)
					for _, b := range inputs.Handlers {
						genesis[b.ID.Name()] = b.DefaultGenesis(cdc)
					}
					return genesis
				}, defaultHome),
				genutilcli.CollectGenTxsCmd(inputs.GenesisBalancesIterator, defaultHome),
				genutilcli.MigrateGenesisCmd(),
				genutilcli.GenTxCmd(validateGenesis, inputs.TxConfig, inputs.GenesisBalancesIterator, defaultHome),
				genutilcli.ValidateGenesisCmd(validateGenesis),
			},
		}
	})
