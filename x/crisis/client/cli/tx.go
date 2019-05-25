// nolint
package cli

import (
	"github.com/spf13/cobra"

	"github.com/YunSuk-Yeo/cosmos-sdk/client/context"
	"github.com/YunSuk-Yeo/cosmos-sdk/client/utils"
	"github.com/YunSuk-Yeo/cosmos-sdk/codec"
	sdk "github.com/YunSuk-Yeo/cosmos-sdk/types"
	authtxb "github.com/YunSuk-Yeo/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/YunSuk-Yeo/cosmos-sdk/x/crisis"
)

// command to replace a delegator's withdrawal address
func GetCmdInvariantBroken(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invariant-broken [module-name] [invariant-route]",
		Short: "submit proof that an invariant broken to halt the chain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			senderAddr := cliCtx.GetFromAddress()
			moduleName, route := args[0], args[1]
			msg := crisis.NewMsgVerifyInvariant(senderAddr, moduleName, route)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
	return cmd
}
