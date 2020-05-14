package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/KiraCore/cosmos-sdk/client/context"
	"github.com/KiraCore/cosmos-sdk/client/flags"
	"github.com/KiraCore/cosmos-sdk/codec"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/version"
	authclient "github.com/KiraCore/cosmos-sdk/x/auth/client"
	authtypes "github.com/KiraCore/cosmos-sdk/x/auth/types"
	"github.com/KiraCore/cosmos-sdk/x/ibc/09-localhost/types"
)

// GetCmdCreateClient defines the command to create a new IBC Client as defined
// in https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#create
func GetCmdCreateClient(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create new localhost client",
		Long: strings.TrimSpace(fmt.Sprintf(`create new localhost (loopback) client:	
Example:	
$ %s tx ibc client localhost create --from node0 --home ../node0/<app>cli --chain-id $CID	
`, version.ClientName),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := authtypes.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc).WithBroadcastMode(flags.BroadcastBlock)

			msg := types.NewMsgCreateClient(cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}
