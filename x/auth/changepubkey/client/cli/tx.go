package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
)

// Transaction command flags
const (
	FlagDelayed = "delayed"
)

// GetTxCmd returns changepubkey module's transaction commands.
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "changepubkey transaction subcommands",
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
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			pubKey, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgChangePubKey(clientCtx.GetFromAddress(), pubKey)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
