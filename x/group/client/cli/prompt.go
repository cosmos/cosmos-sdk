package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"

	"cosmossdk.io/client/v2/autocli/prompt"
	"cosmossdk.io/core/address"
	govcli "cosmossdk.io/x/gov/client/cli"
	govtypes "cosmossdk.io/x/gov/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	proposalText          = "text"
	proposalOther         = "other"
	draftProposalFileName = "draft_group_proposal.json"
	draftMetadataFileName = "draft_group_metadata.json"
)

type proposalType struct {
	Name    string
	MsgType string
	Msg     sdk.Msg
}

// Prompt the proposal type values and return the proposal and its metadata.
func (p *proposalType) Prompt(cdc codec.Codec, skipMetadata bool, addressCodec, validatorAddressCodec, consensusAddressCodec address.Codec) (*Proposal, govtypes.ProposalMetadata, error) {
	// set metadata
	metadata, err := govcli.PromptMetadata(skipMetadata)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set proposal metadata: %w", err)
	}

	proposal := &Proposal{
		Metadata: "ipfs://CID", // the metadata must be saved on IPFS, set placeholder
		Title:    metadata.Title,
		Summary:  metadata.Summary,
	}

	// set group policy address
	groupPolicyAddress, err := prompt.PromptString("Enter group policy address", prompt.ValidateAddress(addressCodec))
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set group policy address: %w", err)
	}
	proposal.GroupPolicyAddress = groupPolicyAddress

	// set proposer address
	proposerAddress, err := prompt.PromptString("Enter proposer address", prompt.ValidateAddress(addressCodec))
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set proposer address: %w", err)
	}
	proposal.Proposers = []string{proposerAddress}

	if p.Msg == nil {
		return proposal, metadata, nil
	}

	// set messages field
	msg, err := protoregistry.GlobalTypes.FindMessageByURL(p.MsgType)
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to find proposal msg: %w", err)
	}
	newMsg := msg.New()

	result, err := prompt.PromptMessage(addressCodec, validatorAddressCodec, consensusAddressCodec, "msg", newMsg)
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

			// prompt proposal type
			selectedProposalType, err := prompt.Select("Select proposal type", []string{proposalText, proposalOther})
			if err != nil {
				return fmt.Errorf("failed to prompt proposal types: %w", err)
			}

			var proposal *proposalType
			switch selectedProposalType {
			case proposalText:
				proposal = &proposalType{Name: proposalText}
			case proposalOther:
				// prompt proposal type
				proposal = &proposalType{Name: proposalOther}

				msgs := clientCtx.InterfaceRegistry.ListImplementations(sdk.MsgInterfaceProtoName)
				sort.Strings(msgs)

				result, err := prompt.Select("Select proposal message type:", msgs)
				if err != nil {
					return fmt.Errorf("failed to prompt proposal types: %w", err)
				}
				proposal.MsgType = result
				proposal.Msg, err = sdk.GetMsgFromTypeURL(clientCtx.Codec, result)
				if err != nil {
					// should never happen
					panic(err)
				}
			default:
				panic("unexpected proposal type")
			}

			skipMetadataPrompt, _ := cmd.Flags().GetBool(flagSkipMetadata)

			result, metadata, err := proposal.Prompt(clientCtx.Codec, skipMetadataPrompt, clientCtx.AddressCodec, clientCtx.ValidatorAddressCodec, clientCtx.ConsensusAddressCodec)
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

// writeFile writes the input to the file.
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
