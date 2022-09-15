package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

const draftProposalFileName = "draft_proposal.json"

type ProposalMetadata struct {
	Title             string `json:"title"`
	Authors           string `json:"authors"`
	Summary           string `json:"summary"`
	Details           string `json:"details"`
	ProposalForumUrl  string `json:"proposal_forum_url"` // named 'Url' instead of 'URL' for avoiding the camel case split
	VoteOptionContext string `json:"vote_option_context"`
}

// ProposalMetadataPrompt prompts the user for filling proposal metadata
func ProposalMetadataPrompt() (ProposalMetadata, error) {
	proposalMetadata := ProposalMetadata{}
	v := reflect.ValueOf(&proposalMetadata).Elem()
	for i := 0; i < v.NumField(); i++ {
		fieldName := strings.ToLower(client.CamelCaseToString(v.Type().Field(i).Name))
		prompt := promptui.Prompt{
			Label:    fmt.Sprintf("Enter proposal %s", fieldName),
			Validate: client.ValidatePromptNotEmpty,
		}

		if strings.Contains(fieldName, "url") {
			prompt.Validate = client.ValidatePromptURL
		}

		result, err := prompt.Run()
		if err != nil {
			return ProposalMetadata{}, fmt.Errorf("failed to prompt for %s: %w", fieldName, err)
		}

		v.Field(i).SetString(result)
	}

	return proposalMetadata, nil
}

type proposalTypes struct {
	Type string
	Msg  interface{}
}

func (p *proposalTypes) Prompt(cdc codec.Codec) (*proposal, error) {
	var proposal = &proposal{}

	// set metadata
	metadata, err := ProposalMetadataPrompt()
	if err != nil {
		return nil, fmt.Errorf("failed to set proposal metadata: %w", err)
	}

	rawMetadata, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proposal metadata: %w", err)
	}

	proposal.Metadata = string(rawMetadata)

	// set deposit
	depositPrompt := client.CoinsAmountPrompt
	depositPrompt.Label = "Enter proposal deposit"

	proposal.Deposit, err = depositPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to set proposal deposit: %w", err)
	}

	if p.Msg == nil {
		return proposal, nil
	}

	// set messages field
	result, err := p.Msg.(proto.Message), nil // TODO - create prompt for this
	if err != nil {
		return nil, fmt.Errorf("failed to set proposal message: %w", err)
	}

	proposal.Messages = append(proposal.Messages, cdc.MustMarshalJSON(result))
	return proposal, nil
}

var supportedProposalTypes = []proposalTypes{
	{
		Type: "text",
		Msg:  nil, // no message for text proposal
	},
	{
		Type: "community-pool-spend",
		Msg:  distrtypes.MsgCommunityPoolSpend{},
	},
	{
		Type: "software-upgrade",
	},
	{
		Type: "cancel-software-upgrade",
	},
	{
		Type: "parameter-change",
	},
}

func getProposalTypes() []string {
	types := make([]string, len(supportedProposalTypes))
	for i, p := range supportedProposalTypes {
		types[i] = p.Type
	}
	return types
}

// NewCmdDraftProposal let a user generate a draft proposal.
func NewCmdDraftProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "draft-proposal [proposal-type]",
		Short: "Generate a draft proposal json file. The generated proposal json contains only one message.",
		Long: `Generate a draft proposal json file. The generated proposal json contains only one message.
The proposal-type can be one of the following:` + strings.Join(getProposalTypes(), ", "),
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var proposal *proposalTypes
			for _, p := range supportedProposalTypes {
				if strings.EqualFold(p.Type, args[0]) {
					proposal = &p
					break
				}
			}
			if proposal == nil {
				return fmt.Errorf("unsupported proposal type \"%s\", supported types are:\n- %s", args[0], strings.Join(getProposalTypes(), "\n- "))
			}

			fmt.Printf("You are generating a draft %s proposal. Please fill in the following information:\n", proposal.Type)

			result, err := proposal.Prompt(clientCtx.Codec)
			if err != nil {
				return err
			}

			rawProposal, err := json.MarshalIndent(result, "", " ")
			if err != nil {
				return fmt.Errorf("failed to marshal proposal: %w", err)
			}

			if err := os.WriteFile(draftProposalFileName, rawProposal, 0o600); err != nil {
				return err
			}

			fmt.Printf("A draft proposal has been generated.\nNote that proposal should contains off-chain metadata.\nPlease upload the metadata.json to IPFS.\nIf you have modified metadata.json, do not forget to update the metadata field in the proposal with the correct CID.\n")

			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
