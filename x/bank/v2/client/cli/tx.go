package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/x/bank/v2/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: Use AutoCLI commands
// https://github.com/cosmos/cosmos-sdk/issues/21682
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bank v2 transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewSendTxCmd(),
		NewCreateDenomTxCmd(),
		NewMintCmd(),
	)

	return txCmd
}

// NewSendTxCmd returns a CLI command handler for creating a MsgSend transaction.
func NewSendTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [from_key_or_address] [to_address] [amount]",
		Short: "Send funds from one account to another.",
		Long: `Send funds from one account to another.
Note, the '--from' flag is ignored as it is implied from [from_key_or_address].
When using '--dry-run' a key name cannot be used, only a bech32 address.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgSend(clientCtx.GetFromAddress().String(), args[1], coins)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewCreateDenomTxCmd returns a CLI command handler for creating a MsgCreateDenom transaction.
func NewCreateDenomTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-denom [subdenom]",
		Short: "Create new tokenfactory denom",
		Long: `Create new tokenfactory denom.
Note, the '--from' flag is ignored as it is implied from [from_key_or_address].
When using '--dry-run' a key name cannot be used, only a bech32 address.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			subDenom := args[0]
			if subDenom == "" {
				return fmt.Errorf("empty denom")
			}

			fromAddr := clientCtx.GetFromAddress()

			msg := &types.MsgCreateDenom{
				Sender:   fromAddr.String(),
				Subdenom: args[0],
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewMintCmd returns a CLI command handler for creating a MsgMint transaction.
func NewMintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint [sender] [mint_to] [amount]",
		Short: "Mint a denom to an address. Must have admin authority to do so.",
		Long: `Mint a denom to an address. Must have admin authority to do so..
Note, the '--from' flag is ignored as it is implied from [from_key_or_address].
When using '--dry-run' a key name cannot be used, only a bech32 address.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			toAddr := args[1]
			if toAddr == "" {
				return fmt.Errorf("Mint to address is empty")
			}

			coin, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgMint{
				Authority: clientCtx.GetFromAddress().String(),
				ToAddress: toAddr,
				Amount: coin,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
