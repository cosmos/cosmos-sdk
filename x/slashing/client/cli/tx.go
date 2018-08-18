package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
	"github.com/cosmos/cosmos-sdk/x/slashing"

	"github.com/spf13/cobra"
)

// GetCmdUnrevoke implements the create unrevoke validator command.
func GetCmdUnrevoke(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unrevoke",
		Args:  cobra.ExactArgs(0),
		Short: "unrevoke validator previously revoked for downtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			validatorAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			msg := slashing.NewMsgUnrevoke(validatorAddr)

			return utils.SendTx(txCtx, cliCtx, []sdk.Msg{msg})
		},
	}

	return cmd
}
