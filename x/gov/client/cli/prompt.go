package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	proposalText          = "text"
	proposalOther         = "other"
	draftProposalFileName = "draft_proposal.json"
)

type ProposalMetadata struct {
	Title string `json:"title"`
	// Authors           string `json:"authors"`
	// Summary           string `json:"summary"`
	// Details           string `json:"details"`
	// ProposalForumUrl  string `json:"proposal_forum_url"` // named 'Url' instead of 'URL' for avoiding the camel case split
	// VoteOptionContext string `json:"vote_option_context"`
}

// ProposalPrompt prompts the user for filling proposal data
func ProposalPrompt[T any](data T) (T, error) {
	v := reflect.ValueOf(&data).Elem()
	if v.Kind() == reflect.Interface {
		v = reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
	}

	for i := 0; i < v.NumField(); i++ {
		fieldName := strings.ToLower(client.CamelCaseToString(v.Type().Field(i).Name))
		prompt := promptui.Prompt{
			Label:    fmt.Sprintf("Enter proposal %s", fieldName),
			Validate: client.ValidatePromptNotEmpty,
		}

		if strings.Contains(fieldName, "url") {
			prompt.Validate = client.ValidatePromptURL
		}

		if strings.EqualFold(v.Type().Field(i).Name, "authority") {
			// pre-fill with gov address
			prompt.Default = authtypes.NewModuleAddress(types.ModuleName).String()
			prompt.Validate = client.ValidatePromptAddress
		}

		result, err := prompt.Run()
		if err != nil {
			return data, fmt.Errorf("failed to prompt for %s: %w", fieldName, err)
		}

		switch v.Field(i).Kind() {
		case reflect.String:
			v.Field(i).SetString(result)
		case reflect.Int64:
			resultInt, _ := strconv.Atoi(result)
			v.Field(i).SetInt(int64(resultInt))
		case reflect.Struct:
			// TODO - manage all different types nicely :thinking_face:
		}
	}

	return data, nil
}

type proposalTypes struct {
	Type    string
	MsgType string
	Msg     sdk.Msg
}

func (p *proposalTypes) Prompt(cdc codec.Codec) (*proposal, error) {
	var proposal = &proposal{}

	// set metadata
	metadata, err := ProposalPrompt(ProposalMetadata{})
	if err != nil {
		return nil, fmt.Errorf("failed to set proposal metadata: %w", err)
	}

	rawMetadata, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proposal metadata: %w", err)
	}

	proposal.Metadata = string(rawMetadata)

	// set deposit
	depositPrompt := promptui.Prompt{
		Label:    "Enter proposal deposit",
		Validate: client.ValidatePromptCoins,
	}
	proposal.Deposit, err = depositPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to set proposal deposit: %w", err)
	}

	if p.Msg == nil {
		return proposal, nil
	}

	// set messages field
	result, err := ProposalPrompt(p.Msg)
	if err != nil {
		return nil, fmt.Errorf("failed to set proposal message: %w", err)
	}

	// TODO enrich message type
	proposal.Messages = append(proposal.Messages, cdc.MustMarshalJSON(result))
	return proposal, nil
}

var supportedProposalTypes = []proposalTypes{
	{
		Type:    proposalText,
		MsgType: "", // no message for text proposal
	},
	{
		Type:    "community-pool-spend",
		MsgType: "/cosmos.distribution.v1beta1.MsgCommunityPoolSpend",
	},
	{
		Type:    "software-upgrade",
		MsgType: "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
	},
	{
		Type:    "cancel-software-upgrade",
		MsgType: "/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
	},
	{
		Type:    proposalOther,
		MsgType: "", // user will input the message type
	},
}

func getProposalTypes() []string {
	types := make([]string, len(supportedProposalTypes))
	for i, p := range supportedProposalTypes {
		types[i] = p.Type
	}
	return types
}

func getProposalMsg(cdc codec.Codec, input string) (sdk.Msg, error) {
	var msg sdk.Msg
	bz, err := json.Marshal(struct {
		Type string `json:"@type"`
	}{
		Type: input,
	})
	if err != nil {
		return nil, err
	}

	if err := cdc.UnmarshalInterfaceJSON(bz, &msg); err != nil {
		return nil, fmt.Errorf("failed to determined sdk.Msg from %s proposal type : %w", input, err)
	}

	return msg, nil
}

// NewCmdDraftProposal let a user generate a draft proposal.
func NewCmdDraftProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "draft-proposal",
		Short:        "Generate a draft proposal json file. The generated proposal json contains only one message.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// prompt proposal type
			proposalTypesPrompt := promptui.Select{
				Label: "Select proposal type",
				Items: getProposalTypes(),
			}

			_, proposalType, err := proposalTypesPrompt.Run()
			if err != nil {
				return fmt.Errorf("failed to prompt proposal types: %w", err)
			}

			var proposal *proposalTypes
			for _, p := range supportedProposalTypes {
				if strings.EqualFold(p.Type, proposalType) {
					proposal = &p
					break
				}
			}

			// create any proposal type
			if proposal.Type == proposalOther {
				// prompt proposal type
				msgPrompt := promptui.Prompt{
					Label: "Which message type do you want to use for the proposal",
					Validate: func(input string) error {
						_, err := getProposalMsg(clientCtx.Codec, input)
						return err
					},
				}

				result, err := msgPrompt.Run()
				if err != nil {
					return fmt.Errorf("failed to prompt proposal types: %w", err)
				}

				proposal.Msg, err = getProposalMsg(clientCtx.Codec, result)
				if err != nil {
					return err
				}
			} else if proposal.MsgType != "" {
				proposal.Msg, err = getProposalMsg(clientCtx.Codec, proposal.MsgType)
				if err != nil {
					return err
				}
			}

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

			fmt.Printf("A draft proposal has been generated.\nNote that proposal should contains off-chain metadata.\nPlease upload the metadata object to IPFS.\nThen, replace the generated metadata field with the IPFS CID.\n")

			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
