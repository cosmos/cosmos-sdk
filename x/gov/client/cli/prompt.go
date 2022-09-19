package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
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
	draftMetadataFileName = "draft_metadata.json"
)

// ProposalMetadata is the metadata of a proposal
// This metadata is supposed to live off-chain when submitted in a proposal
type ProposalMetadata struct {
	Title             string `json:"title"`
	Authors           string `json:"authors"`
	Summary           string `json:"summary"`
	Details           string `json:"details"`
	ProposalForumUrl  string `json:"proposal_forum_url"` // named 'Url' instead of 'URL' for avoiding the camel case split
	VoteOptionContext string `json:"vote_option_context"`
}

// Prompt prompts the user for all values of the given type.
// data is the struct to be filled
// namePrefix is the name to be display as "Enter <namePrefix> <field>"
func Prompt[T any](data T, namePrefix string) (T, error) {
	v := reflect.ValueOf(&data).Elem()
	if v.Kind() == reflect.Interface {
		v = reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
	}

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Kind() == reflect.Struct || v.Field(i).Kind() == reflect.Slice {
			// if the field is a struct skip
			// in a future we can add a recursive call to Prompt
			continue
		}

		// create prompts
		prompt := promptui.Prompt{
			Label:    fmt.Sprintf("Enter %s %s", namePrefix, strings.ToLower(client.CamelCaseToString(v.Type().Field(i).Name))),
			Validate: client.ValidatePromptNotEmpty,
		}

		fieldName := strings.ToLower(v.Type().Field(i).Name)
		// validation per field name
		if strings.Contains(fieldName, "url") {
			prompt.Validate = client.ValidatePromptURL
		}

		if strings.EqualFold(fieldName, "authority") {
			// pre-fill with gov address
			prompt.Default = authtypes.NewModuleAddress(types.ModuleName).String()
			prompt.Validate = client.ValidatePromptAddress
		}

		if strings.Contains(fieldName, "addr") ||
			strings.Contains(fieldName, "sender") ||
			strings.Contains(fieldName, "voter") ||
			strings.Contains(fieldName, "depositor") ||
			strings.Contains(fieldName, "granter") ||
			strings.Contains(fieldName, "grantee") ||
			strings.Contains(fieldName, "recipient") {
			prompt.Validate = client.ValidatePromptAddress
		}

		result, err := prompt.Run()
		if err != nil {
			return data, fmt.Errorf("failed to prompt for %s: %w", fieldName, err)
		}

		switch v.Field(i).Kind() {
		case reflect.String:
			v.Field(i).SetString(result)
		case reflect.Int:
			resultInt, _ := strconv.Atoi(result)
			v.Field(i).SetInt(int64(resultInt))
		default:
			// skip other types
			// possibly in the future we can add more types (like slices)
			continue
		}
	}

	return data, nil
}

type proposalTypes struct {
	Type    string
	MsgType string
	Msg     sdk.Msg
}

// Prompt the proposal type values and return the proposal and its metadata
func (p *proposalTypes) Prompt(cdc codec.Codec) (*proposal, ProposalMetadata, error) {
	proposal := &proposal{}

	// set metadata
	metadata, err := Prompt(ProposalMetadata{}, "proposal")
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set proposal metadata: %w", err)
	}
	proposal.Metadata = "ipfs://CID"

	// set deposit
	depositPrompt := promptui.Prompt{
		Label:    "Enter proposal deposit",
		Validate: client.ValidatePromptCoins,
	}
	proposal.Deposit, err = depositPrompt.Run()
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set proposal deposit: %w", err)
	}

	if p.Msg == nil {
		return proposal, metadata, nil
	}

	// set messages field
	result, err := Prompt(p.Msg, "msg")
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set proposal message: %w", err)
	}

	message, err := cdc.MarshalInterfaceJSON(result)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to marshal proposal message: %w", err)
	}
	proposal.Messages = append(proposal.Messages, message)
	return proposal, metadata, nil
}

var supportedProposalTypes = []proposalTypes{
	{
		Type:    proposalText,
		MsgType: "", // no message for text proposal
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

			var proposal proposalTypes
			for _, p := range supportedProposalTypes {
				if strings.EqualFold(p.Type, proposalType) {
					proposal = p
					break
				}
			}

			// create any proposal type
			if proposal.Type == proposalOther {
				// prompt proposal type
				msgPrompt := promptui.Select{
					Label: "Select proposal message type:",
					Items: func() []string {
						msgs := clientCtx.InterfaceRegistry.ListImplementations(sdk.MsgInterfaceProtoName)
						sort.Strings(msgs)
						return msgs
					}(),
				}

				_, result, err := msgPrompt.Run()
				if err != nil {
					return fmt.Errorf("failed to prompt proposal types: %w", err)
				}

				proposal.MsgType = result
			}

			if proposal.MsgType != "" {
				proposal.Msg, err = getProposalMsg(clientCtx.Codec, proposal.MsgType)
				if err != nil {
					// should never happen
					panic(err)
				}
			}

			prop, metadata, err := proposal.Prompt(clientCtx.Codec)
			if err != nil {
				return err
			}

			if err := writeFile(draftMetadataFileName, metadata); err != nil {
				return err
			}

			if err := writeFile(draftProposalFileName, prop); err != nil {
				return err
			}

			fmt.Printf("Your draft proposal has successfully been generated.\nProposals should contain off-chain metadata, please upload the metadata JSON to IPFS.\nThen, replace the generated metadata field with the IPFS CID.\n")

			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func writeFile(fileName string, input any) error {
	raw, err := json.MarshalIndent(input, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal proposal: %w", err)
	}

	if err := os.WriteFile(fileName, raw, 0o600); err != nil {
		return err
	}

	return nil
}
