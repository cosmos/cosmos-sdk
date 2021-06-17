package module

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/container"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

var (
	_ app.Provisioner = Module{}
)

type Inputs struct {
	dig.In

	Commands  []*cobra.Command `group:"cosmos.tx.v1.Command"`
	Marshaler codec.ProtoCodecMarshaler
}

type Outputs struct {
	dig.Out

	TxConfig      client.TxConfig
	Command       *cobra.Command   `group:"cli.root"`
	QueryCommands []*cobra.Command `group:"cosmos.query.v1.Command,flatten"`
}

func (m Module) Provision(_ app.ModuleKey, registrar container.Registrar) error {
	return registrar.Provide(func(inputs Inputs) Outputs {
		signModes := m.EnabledSignModes
		if signModes == nil {
			signModes = tx.DefaultSignModes
		}

		cmd := &cobra.Command{
			Use:                        "tx",
			Short:                      "Transactions subcommands",
			DisableFlagParsing:         true,
			SuggestionsMinimumDistance: 2,
			RunE:                       client.ValidateCmd,
		}

		cmd.AddCommand(
			authcmd.GetSignCommand(),
			authcmd.GetSignBatchCommand(),
			authcmd.GetMultiSignCommand(),
			authcmd.GetMultiSignBatchCmd(),
			authcmd.GetValidateSignaturesCommand(),
			authcmd.GetBroadcastCommand(),
			authcmd.GetEncodeCommand(),
			authcmd.GetDecodeCommand(),
		)

		for _, c := range inputs.Commands {
			if c != nil {
				cmd.AddCommand(c)
			}
		}

		cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

		return Outputs{
			TxConfig: tx.NewTxConfig(inputs.Marshaler, signModes),
			Command:  cmd,
			QueryCommands: []*cobra.Command{
				authcmd.QueryTxsByEventsCmd(),
				authcmd.QueryTxCmd(),
			},
		}
	})
}
