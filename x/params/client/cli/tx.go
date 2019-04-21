package cli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govccli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramscutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
)

// GetCmdSubmitProposal implements a command handler for submitting a parameter
// change proposal transaction.
func GetCmdSubmitProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "paramchange",
		Short: "Submit a parameter change proposal",
		Long: strings.TrimSpace(`
Submit a parameter proposal along with an initial deposit. The proposal details
must be supplied via a JSON file.

Example:
$ gaiacli tx gov submit-proposal paramchange --proposal="path/to/proposal.json" --from <key_or_address>
where proposal.json contains:
{
  "title": "Test Proposal",
  "description": "Increase max validator",
  "changes": [
    {
	  space: "staking",
	  key: "MaxValidators",
	  value: 105
    }
  ],
  "deposit": "10atom"
}
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			proposal, err := paramscutils.ParseParamChangeProposalJSON(cdc, viper.GetString(govccli.FlagProposal))
			if err != nil {
				return err
			}

			from := cliCtx.GetFromAddress()
			content := params.NewParameterChangeProposal(proposal.Title, proposal.Description, proposal.Changes)

			msg := gov.NewMsgSubmitProposal(content, proposal.Deposit, from)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}

	cmd.Flags().String(govccli.FlagProposal, "", "The proposal file path (if set, other proposal flags are ignored)")

	return cmd
}
