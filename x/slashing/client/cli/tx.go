package cli

import (
	"bufio"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// NewTxCmd returns a root CLI command handler for all x/slashing transaction commands.
func NewTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	slashingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Slashing transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	slashingTxCmd.AddCommand(NewUnjailTxCmd(m, txg, ar))
	return slashingTxCmd
}

func NewUnjailTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unjail",
		Args:  cobra.NoArgs,
		Short: "unjail validator previously jailed for downtime",
		Long: `unjail a jailed validator:

$ <appcli> tx slashing unjail --from mykey
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txf := tx.NewFactoryFromCLI(inBuf).
				WithTxGenerator(txg).
				WithAccountRetriever(ar)

			cliCtx := context.NewCLIContextWithInput(inBuf).WithMarshaler(m)

			valAddr := cliCtx.GetFromAddress()
			msg := types.NewMsgUnjail(sdk.ValAddress(valAddr))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(cliCtx, txf, msg)
		},
	}
	return flags.PostCommands(cmd)[0]
}

// ---------------------------------------------------------------------------
// Deprecated
//
// TODO: Remove once client-side Protobuf migration has been completed.
// ---------------------------------------------------------------------------

// GetTxCmd returns the transaction commands for this module
//
// TODO: Remove once client-side Protobuf migration has been completed.
// ref: https://github.com/cosmos/cosmos-sdk/issues/5864
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	slashingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Slashing transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	slashingTxCmd.AddCommand(flags.PostCommands(
		GetCmdUnjail(cdc),
	)...)

	return slashingTxCmd
}

// GetCmdUnjail implements the create unjail validator command.
//
// TODO: Remove once client-side Protobuf migration has been completed.
// ref: https://github.com/cosmos/cosmos-sdk/issues/5864
func GetCmdUnjail(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unjail",
		Args:  cobra.NoArgs,
		Short: "unjail validator previously jailed for downtime",
		Long: `unjail a jailed validator:

$ <appcli> tx slashing unjail --from mykey
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			valAddr := cliCtx.GetFromAddress()

			msg := types.NewMsgUnjail(sdk.ValAddress(valAddr))
			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
