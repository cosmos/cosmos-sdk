package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/prompt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	proposalText          = "text"
	proposalOther         = "other"
	draftProposalFileName = "draft_proposal.json"
	draftMetadataFileName = "draft_metadata.json"
)

var suggestedProposalTypes = []proposalType{
	{
		Name:    proposalText,
		MsgType: "", // no message for text proposal
	},
	{
		Name:    "community-pool-spend",
		MsgType: "/cosmos.distribution.v1beta1.MsgCommunityPoolSpend",
	},
	{
		Name:    "software-upgrade",
		MsgType: "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
	},
	{
		Name:    "cancel-software-upgrade",
		MsgType: "/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
	},
	{
		Name:    proposalOther,
		MsgType: "", // user will input the message type
	},
}

// Prompt prompts the user for all values of the given type.
// Deprecated: This is not used anymore anywhere thanks to client proto prompt.
// It will be removed in a future release.
func Prompt[T any](data T, namePrefix string) (T, error) {
	return data, errors.New("use client/prompt instead")
}

type proposalType struct {
	Name    string
	MsgType string
	Msg     sdk.Msg
}

// Prompt the proposal type values and return the proposal and its metadata
func (p *proposalType) Prompt(cdc codec.Codec, skipMetadata bool) (*proposal, types.ProposalMetadata, error) {
	// before codecs are not more in context, we re-create them from global SDK config
	cfg := sdk.GetConfig()
	addressCodec := address.NewBech32Codec(cfg.GetBech32AccountAddrPrefix())
	validatorAddressCodec := address.NewBech32Codec(cfg.GetBech32ValidatorAddrPrefix())
	consensusAddressCodec := address.NewBech32Codec(cfg.GetBech32ConsensusAddrPrefix())

	metadata, err := PromptMetadata(skipMetadata)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set proposal metadata: %w", err)
	}

	proposal := &proposal{
		Metadata: "ipfs://CID", // the metadata must be saved on IPFS, set placeholder
		Title:    metadata.Title,
		Summary:  metadata.Summary,
	}

	// set deposit
	proposal.Deposit, err = prompt.PromptString("Enter proposal deposit", prompt.ValidatePromptCoins)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set proposal deposit: %w", err)
	}

	if p.Msg == nil {
		return proposal, metadata, nil
	}

	// set messages field
	msg, err := protoregistry.GlobalTypes.FindMessageByURL(p.MsgType)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to find proposal msg: %w", err)
	}
	newMsg := msg.New()
	govAddr := sdkaddress.Module(types.ModuleName)
	govAddrStr, err := addressCodec.BytesToString(govAddr)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to convert gov address to string: %w", err)
	}

	prompt.SetDefaults(newMsg, map[string]interface{}{"authority": govAddrStr})
	result, err := prompt.PromptMessage(
		addressCodec,
		validatorAddressCodec,
		consensusAddressCodec,
		"msg",
		newMsg,
	)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set proposal message: %w", err)
	}

	// message must be converted to gogoproto so @type is not lost
	resultBytes, err := proto.Marshal(result.Interface())
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to marshal proposal message: %w", err)
	}

	err = gogoproto.Unmarshal(resultBytes, p.Msg)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to unmarshal proposal message: %w", err)
	}

	message, err := cdc.MarshalInterfaceJSON(p.Msg)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to marshal proposal message: %w", err)
	}
	proposal.Messages = append(proposal.Messages, message)

	return proposal, metadata, nil
}

// getProposalSuggestions suggests a list of proposal types
func getProposalSuggestions() []string {
	types := make([]string, len(suggestedProposalTypes))
	for i, p := range suggestedProposalTypes {
		types[i] = p.Name
	}
	return types
}

// PromptMetadata prompts for proposal metadata or only title and summary if skip is true
func PromptMetadata(skip bool) (types.ProposalMetadata, error) {
	if !skip {
		metadata, err := prompt.PromptStruct("proposal", types.ProposalMetadata{})
		if err != nil {
			return types.ProposalMetadata{}, err
		}

		return metadata, nil
	}

	title, err := prompt.PromptString("Enter proposal title", prompt.ValidatePromptNotEmpty)
	if err != nil {
		return types.ProposalMetadata{}, fmt.Errorf("failed to set proposal title: %w", err)
	}

	summary, err := prompt.PromptString("Enter proposal summary", prompt.ValidatePromptNotEmpty)
	if err != nil {
		return types.ProposalMetadata{}, fmt.Errorf("failed to set proposal summary: %w", err)
	}

	return types.ProposalMetadata{Title: title, Summary: summary}, nil
}

// NewCmdDraftProposal let a user generate a draft proposal.
func NewCmdDraftProposal() *cobra.Command {
	flagSkipMetadata := "skip-metadata"

	cmd := &cobra.Command{
		Use:          "draft-proposal",
		Short:        "Generate a draft proposal json file. The generated proposal json contains only one message (skeleton).",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			selectedProposalType, err := prompt.Select("Select proposal type", getProposalSuggestions())
			if err != nil {
				return fmt.Errorf("failed to prompt proposal types: %w", err)
			}
			var proposal proposalType
			for _, p := range suggestedProposalTypes {
				if strings.EqualFold(p.Name, selectedProposalType) {
					proposal = p
					break
				}
			}

			// create any proposal type
			if proposal.Name == proposalOther {
				msgs := clientCtx.InterfaceRegistry.ListImplementations(sdk.MsgInterfaceProtoName)
				sort.Strings(msgs)

				result, err := prompt.Select("Select proposal message type:", msgs)
				if err != nil {
					return fmt.Errorf("failed to prompt proposal types: %w", err)
				}

				proposal.MsgType = result
			}

			if proposal.MsgType != "" {
				proposal.Msg, err = sdk.GetMsgFromTypeURL(clientCtx.Codec, proposal.MsgType)
				if err != nil {
					// should never happen
					panic(err)
				}
			}

			skipMetadataPrompt, _ := cmd.Flags().GetBool(flagSkipMetadata)

			result, metadata, err := proposal.Prompt(clientCtx.Codec, skipMetadataPrompt)
			if err != nil {
				return err
			}

			if err := writeFile(draftProposalFileName, result); err != nil {
				return err
			}

			if !skipMetadataPrompt {
				if err := writeFile(draftMetadataFileName, metadata); err != nil {
					return err
				}
			}

			cmd.Println("The draft proposal has successfully been generated.\nProposals should contain off-chain metadata, please upload the metadata JSON to IPFS.\nThen, replace the generated metadata field with the IPFS CID.")

			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Bool(flagSkipMetadata, false, "skip metadata prompt")

	return cmd
}

// writeFile writes the input to the file
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
