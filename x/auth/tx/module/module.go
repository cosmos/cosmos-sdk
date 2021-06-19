package module

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/spf13/cobra"
	"go.uber.org/dig"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

type Inputs struct {
	dig.In

	Commands       []*cobra.Command `group:"tx"`
	Marshaler      codec.ProtoCodecMarshaler
	AuthKeeper     auth.Keeper
	BankKeeper     bankkeeper.Keeper
	FeegrantKeeper feegrantkeeper.Keeper `optional:"true"`
}

type Outputs struct {
	dig.Out

	TxDecoder     types.TxDecoder
	TxConfig      client.TxConfig
	Command       *cobra.Command   `group:"root"`
	QueryCommands []*cobra.Command `group:"query,flatten"`
	AnteHandler   types.AnteHandler
}

func (m Module) Provide(inputs Inputs) (Outputs, error) {
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

	txConfig := tx.NewTxConfig(inputs.Marshaler, signModes)

	anteHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   inputs.AuthKeeper,
		BankKeeper:      inputs.BankKeeper,
		FeegrantKeeper:  inputs.FeegrantKeeper,
		SignModeHandler: txConfig.SignModeHandler(),
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
	})

	if err != nil {
		return Outputs{}, err
	}

	return Outputs{
		TxDecoder: txConfig.TxDecoder(),
		TxConfig:  txConfig,
		Command:   cmd,
		QueryCommands: []*cobra.Command{
			authcmd.QueryTxsByEventsCmd(),
			authcmd.QueryTxCmd(),
		},
		AnteHandler: anteHandler,
	}, nil
}
