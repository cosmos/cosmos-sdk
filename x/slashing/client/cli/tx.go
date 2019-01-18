package cli

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"

	"github.com/spf13/cobra"
)

// GetCmdUnjail implements the create unjail validator command.
func GetCmdUnjail(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unjail",
		Args:  cobra.NoArgs,
		Short: "unjail validator previously jailed for downtime",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			valAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			return cliCtx.MessageOutput(slashing.NewMsgUnjail(sdk.ValAddress(valAddr)))
		},
	}
}
