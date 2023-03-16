package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/x/auth/rekeying/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns rekeying module's transaction commands.
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Rekeying transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewMsgChangePubKeyCmd(),
	)

	return txCmd
}

// NewMsgChangePubKeyCmd returns a CLI command handler for creating a
// MsgChangePubKey transaction.
func NewMsgChangePubKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-pubkey [pubkey]",
		Short: "Change PubKey of an account.",
		Long: `This msg will update the public key associated with an account
		 to a new public key, while keeping the same address.
		 This can be used for purposes such as passing ownership of an account
		 to a new key for security reasons or upgrading multisig signers.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var pk cryptotypes.PubKey
			if err = clientCtx.Codec.UnmarshalInterfaceJSON([]byte(args[0]), &pk); err != nil {
				return err
			}

			msg := types.NewMsgChangePubKey(clientCtx.GetFromAddress().String(), pk)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
