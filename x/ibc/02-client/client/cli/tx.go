package cli

import (
	"io/ioutil"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	cli "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
)

const (
	FlagStatePath            = "state"
	FlagClientID             = "client-id"
	FlagConnectionID         = "connection-id"
	FlagChannelID            = "channel-id"
	FlagCounterpartyID       = "counterparty-id"
	FlagCounterpartyClientID = "counterparty-client-id"
	FlagSourceNode           = "source-node"
)

func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcTxCmd := &cobra.Command{
		Use:                        "client",
		Short:                      "Client transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ibcTxCmd.AddCommand(cli.PostCommands(
		GetCmdCreateClient(cdc),
		GetCmdUpdateClient(cdc),
	)...)

	return ibcTxCmd
}

func GetCmdCreateClient(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [client-id] [consensus-state]",
		Short: "create new client with a consensus state",
		Long: `create a new IBC client on the target chain for a chain w/ given consensus-state. Consensus state can be passed in as a filepath or a string:

$ gaiacli tx ibc client create clientFoo $(gaiacli --home /path/to/chain2 q ibc client consensus-state)
$ gaiacli tx ibc client create clientFoo ./state.json
`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc)

			var state client.ConsensusState
			if err := cdc.UnmarshalJSON([]byte(args[1]), &state); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return fmt.Errorf("error opening state file: %v\n", err)
				}
				if err := cdc.UnmarshalJSON(contents, &state); err != nil {
					return fmt.Errorf("error unmarshalling state file: %v\n", err)
				}
			}

			msg := client.MsgCreateClient{
				ClientID:       args[0],
				ConsensusState: state,
				Signer:         cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

func GetCmdUpdateClient(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [client-id] [consensus-state]",
		Short: "update existing client with a header",
		Long: `update an existing IBC client on the target chain for a chain w/ given consensus-state. Consensus state can be passed in as a filepath or a string:

$ gaiacli tx ibc client update clientFoo $(gaiacli --home /path/to/chain2 q ibc client consensus-state)
$ gaiacli tx ibc client update clientFoo ./state.json
`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc)

			var header client.Header
			if err := cdc.UnmarshalJSON([]byte(args[1]), &header); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
				contents, err := ioutil.ReadFile(args[1])
				if err != nil {
					return fmt.Errorf("error opening header file: %v\n", err)
				}
				if err := cdc.UnmarshalJSON(contents, &header); err != nil {
					return fmt.Errorf("error unmarshalling header file: %v\n", err)
				}
			}

			msg := client.MsgUpdateClient{
				ClientID: args[0],
				Header:   header,
				Signer:   cliCtx.GetFromAddress(),
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}
