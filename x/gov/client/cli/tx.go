package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"

	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	gcutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
)

const (
	flagTitle        = "title"
	flagDescription  = "description"
	flagProposalType = "type"
	flagDeposit      = "deposit"
	flagVoter        = "voter"
	flagOption       = "option"
	flagDepositor    = "depositor"
	flagStatus       = "status"
	flagNumLimit     = "limit"
	flagProposal     = "proposal"
)

type proposal struct {
	Title       string
	Description string
	Type        string
	Deposit     string
}

var proposalFlags = []string{
	flagTitle,
	flagDescription,
	flagProposalType,
	flagDeposit,
}

// GetCmdSubmitProposal implements submitting a proposal transaction command.
func GetCmdSubmitProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal",
		Short: "Submit a proposal along with an initial deposit",
		Long: strings.TrimSpace(`
Submit a proposal along with an initial deposit. Proposal title, description, type and deposit can be given directly or through a proposal JSON file. For example:

$ gaiacli gov submit-proposal --proposal="path/to/proposal.json" --from mykey

where proposal.json contains:

{
  "title": "Test Proposal",
  "description": "My awesome proposal",
  "type": "Text",
  "deposit": "10test"
}

is equivalent to

$ gaiacli gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type="Text" --deposit="10test" --from mykey
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			proposal, err := parseSubmitProposalFlags()
			if err != nil {
				return err
			}

			// Get from address
			from := cliCtx.FromAddr()

			// Pull associated account
			account, err := cliCtx.FetchAccount(from)
			if err != nil {
				return err
			}

			// Find deposit amount
			amount, err := sdk.ParseCoins(proposal.Deposit)
			if err != nil {
				return err
			}

			// ensure account has enough coins
			if !account.GetCoins().IsAllGTE(amount) {
				return context.ErrInsufficientFunds(account, amount)
			}

			proposalType, err := gov.ProposalTypeFromString(proposal.Type)
			if err != nil {
				return err
			}

			msg := gov.NewMsgSubmitProposal(proposal.Title, proposal.Description, proposalType, from, amount)
			return cliCtx.MessageOutput(msg)
		},
	}

	cmd.Flags().String(flagTitle, "", "title of proposal")
	cmd.Flags().String(flagDescription, "", "description of proposal")
	cmd.Flags().String(flagProposalType, "", "proposalType of proposal, types: text/parameter_change/software_upgrade")
	cmd.Flags().String(flagDeposit, "", "deposit of proposal")
	cmd.Flags().String(flagProposal, "", "proposal file path (if this path is given, other proposal flags are ignored)")

	return cmd
}

func parseSubmitProposalFlags() (*proposal, error) {
	proposal := &proposal{}
	proposalFile := viper.GetString(flagProposal)

	if proposalFile == "" {
		proposal.Title = viper.GetString(flagTitle)
		proposal.Description = viper.GetString(flagDescription)
		proposal.Type = gcutils.NormalizeProposalType(viper.GetString(flagProposalType))
		proposal.Deposit = viper.GetString(flagDeposit)
		return proposal, nil
	}

	for _, flag := range proposalFlags {
		if viper.GetString(flag) != "" {
			return nil, fmt.Errorf("--%s flag provided alongside --proposal, which is a noop", flag)
		}
	}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, proposal)
	if err != nil {
		return nil, err
	}

	return proposal, nil
}

// GetCmdDeposit implements depositing tokens for an active proposal.
func GetCmdDeposit(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deposit [proposal-id] [deposit]",
		Args:  cobra.ExactArgs(2),
		Short: "Deposit tokens for activing proposal",
		Long: strings.TrimSpace(`
Submit a deposit for an acive proposal. You can find the proposal-id by running gaiacli query gov proposals:

$ gaiacli tx gov deposit 1 10stake --from mykey
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return gcutils.InvalidProposalID(args[0])
			}

			// check to see if the proposal is in the store
			_, err = gcutils.QueryProposalByID(proposalID, cliCtx, cdc, queryRoute)
			if err != nil {
				return gcutils.FailedToFectchProposal(proposalID, err)
			}

			// Fetch associated account
			account, err := cliCtx.FetchAccount(cliCtx.FromAddr())
			if err != nil {
				return err
			}

			// Get amount of coins
			amount, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			// ensure account has enough coins
			if !account.GetCoins().IsAllGTE(amount) {
				return context.ErrInsufficientFunds(account, amount)
			}

			return cliCtx.MessageOutput(gov.NewMsgDeposit(cliCtx.FromAddr(), proposalID, amount))
		},
	}
}

// GetCmdVote implements creating a new vote command.
func GetCmdVote(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "vote [proposal-id] [option]",
		Args:  cobra.ExactArgs(2),
		Short: "Vote for an active proposal, options: yes/no/no_with_veto/abstain",
		Long: strings.TrimSpace(`
Submit a vote for an acive proposal. You can find the proposal-id by running gaiacli query gov proposals:

$ gaiacli tx gov vote 1 yes --from mykey
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return gcutils.InvalidProposalID(args[0])
			}

			// check to see if the proposal is in the store
			_, err = gcutils.QueryProposalByID(proposalID, cliCtx, cdc, queryRoute)
			if err != nil {
				return gcutils.FailedToFectchProposal(proposalID, err)
			}

			// Find out which vote option user chose
			byteVoteOption, err := gov.VoteOptionFromString(gcutils.NormalizeVoteOption(args[1]))
			if err != nil {
				return err
			}

			// Build vote message and run basic validation
			return cliCtx.MessageOutput(gov.NewMsgVote(cliCtx.FromAddr(), proposalID, byteVoteOption))
		},
	}
}
