package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramscutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

// NewSubmitParamChangeProposalTxCmd returns a CLI command handler for creating
// a parameter change proposal governance transaction.
func NewSubmitParamChangeProposalTxCmd(ctx context.CLIContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "param-change [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a parameter change proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a parameter proposal along with an initial deposit.
The proposal details must be supplied via a JSON file. For values that contains
objects, only non-empty fields will be updated.

IMPORTANT: Currently parameter changes are evaluated but not validated, so it is
very important that any "value" change is valid (ie. correct type and within bounds)
for its respective parameter, eg. "MaxValidators" should be an integer and not a decimal.

Proper vetting of a parameter change proposal should prevent this from happening
(no deposits should occur during the governance process), but it should be noted
regardless.

Example:
$ %s tx gov submit-proposal param-change <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "Staking Param Change",
  "description": "Update max validators",
  "changes": [
    {
      "subspace": "staking",
      "key": "MaxValidators",
      "value": 105
    }
  ],
  "deposit": "1000stake"
}
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := ctx.InitWithInput(cmd.InOrStdin())

			proposal, err := paramscutils.ParseParamChangeProposalJSON(ctx.JSONMarshaler, args[0])
			if err != nil {
				return err
			}

			from := cliCtx.GetFromAddress()
			content := paramproposal.NewParameterChangeProposal(
				proposal.Title, proposal.Description, proposal.Changes.ToParamChanges(),
			)

			deposit, err := sdk.ParseCoins(proposal.Deposit)
			if err != nil {
				return err
			}

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(cliCtx, msg)
		},
	}

	return cmd
}
