package provider

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

type Inputs struct {
	dig.In

	DefaultHome             string          `name:"cli.default-home"`
	Handlers                []types.Handler `group:"cosmos.genutil.v1.Handler"`
	GenesisBalancesIterator types.GenesisBalancesIterator
	TxConfig                client.TxConfig
}

type Outputs struct {
	dig.Out

	RootCommands []*cobra.Command `group:"cli.root,flatten"`
}

func Provider(inputs Inputs) Outputs {
	validateGenesis := func(cdc codec.JSONCodec, txEncCfg client.TxEncodingConfig, genesis map[string]json.RawMessage) error {
		for _, b := range inputs.Handlers {
			if err := b.ValidateGenesis(cdc, txEncCfg, genesis[b.ID.Name()]); err != nil {
				return err
			}
		}
		return nil
	}

	return Outputs{
		RootCommands: []*cobra.Command{
			genutilcli.InitCmd(func(cdc codec.JSONCodec) map[string]json.RawMessage {
				genesis := make(map[string]json.RawMessage)
				for _, b := range inputs.Handlers {
					genesis[b.ID.Name()] = b.DefaultGenesis(cdc)
				}
				return genesis
			}, inputs.DefaultHome),
			genutilcli.CollectGenTxsCmd(inputs.GenesisBalancesIterator, inputs.DefaultHome),
			genutilcli.MigrateGenesisCmd(),
			genutilcli.GenTxCmd(validateGenesis, inputs.TxConfig, inputs.GenesisBalancesIterator, inputs.DefaultHome),
			genutilcli.ValidateGenesisCmd(validateGenesis),
		},
	}
}
