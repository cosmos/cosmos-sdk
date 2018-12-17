package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/gov"

	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	govClientUtils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
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
			proposal, err := parseSubmitProposalFlags()
			if err != nil {
				return err
			}

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// Get from address
			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			// Pull associated account
			account, err := cliCtx.GetAccount(from)
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
				return errors.Errorf("Address %s doesn't have enough coins to pay for this transaction.", from)
			}

			proposalType, err := gov.ProposalTypeFromString(proposal.Type)
			if err != nil {
				return err
			}

			msg := gov.NewMsgSubmitProposal(proposal.Title, proposal.Description, proposalType, from, amount)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			if cliCtx.GenerateOnly {
				return utils.PrintUnsignedStdTx(os.Stdout, txBldr, cliCtx, []sdk.Msg{msg}, false)
			}

			// Build and sign the transaction, then broadcast to Tendermint
			// proposalID must be returned, and it is a part of response.
			cliCtx.PrintResponse = true
			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
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
		proposal.Type = govClientUtils.NormalizeProposalType(viper.GetString(flagProposalType))
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
	cmd := &cobra.Command{
		Use:   "deposit [proposal-id] [deposit]",
		Args:  cobra.ExactArgs(2),
		Short: "Deposit tokens for activing proposal",
		Long: strings.TrimSpace(`
Submit a deposit for an acive proposal. You can find the proposal-id by running gaiacli query gov proposals:

$ gaiacli tx gov deposit 1 10STAKE --from mykey
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = queryProposal(proposalID, cliCtx, cdc, queryRoute)
			if err != nil {
				return fmt.Errorf("Failed to fetch proposal-id %d: %s", proposalID, err)
			}

			// Get from address
			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			// Fetch associated account
			account, err := cliCtx.GetAccount(from)
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
				return errors.Errorf("Address %s doesn't have enough coins to pay for this transaction.", from)
			}

			msg := gov.NewMsgDeposit(from, proposalID, amount)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			if cliCtx.GenerateOnly {
				return utils.PrintUnsignedStdTx(os.Stdout, txBldr, cliCtx, []sdk.Msg{msg}, false)
			}

			// Build and sign the transaction, then broadcast to a Tendermint node.
			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdVote implements creating a new vote command.
func GetCmdVote(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [option]",
		Args:  cobra.ExactArgs(2),
		Short: "Vote for an active proposal, options: yes/no/no_with_veto/abstain",
		Long: strings.TrimSpace(`
Submit a vote for an acive proposal. You can find the proposal-id by running gaiacli query gov proposals:

$ gaiacli tx gov vote 1 yes --from mykey
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// Get voting address
			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			// validate that the proposal id is a uint
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("proposal-id %s not a valid int, please input a valid proposal-id", args[0])
			}

			// check to see if the proposal is in the store
			_, err = queryProposal(proposalID, cliCtx, cdc, queryRoute)
			if err != nil {
				return fmt.Errorf("Failed to fetch proposal-id %d: %s", proposalID, err)
			}

			// Find out which vote option user chose
			byteVoteOption, err := gov.VoteOptionFromString(govClientUtils.NormalizeVoteOption(args[1]))
			if err != nil {
				return err
			}

			// Build vote message and run basic validation
			msg := gov.NewMsgVote(from, proposalID, byteVoteOption)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			// If generate only print the transaction
			if cliCtx.GenerateOnly {
				return utils.PrintUnsignedStdTx(os.Stdout, txBldr, cliCtx, []sdk.Msg{msg}, false)
			}

			// Build and sign the transaction, then broadcast to a Tendermint node.
			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}

	return cmd
}
