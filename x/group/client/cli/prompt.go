package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const (
	proposalText          = "text"
	proposalOther         = "other"
	draftProposalFileName = "draft_group_proposal.json"
	draftMetadataFileName = "draft_group_metadata.json"
)

type proposalType struct {
	Name string
	Msg  sdk.Msg
}

// Prompt the proposal type values and return the proposal and its metadata
func (p *proposalType) Prompt(cdc codec.Codec) (*Proposal, govtypes.ProposalMetadata, error) {
	proposal := &Proposal{}

	// set metadata
	metadata, err := govcli.Prompt(govtypes.ProposalMetadata{}, "proposal")
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set proposal metadata: %w", err)
	}
	// the metadata must be saved on IPFS, set placeholder
	proposal.Metadata = "ipfs://CID"

	// set group policy address
	policyAddressPrompt := promptui.Prompt{
		Label:    "Enter group policy address",
		Validate: client.ValidatePromptAddress,
	}
	groupPolicyAddress, err := policyAddressPrompt.Run()
	if err != nil {
		return nil, metadata, fmt.Errorf("failed to set group policy address: %w", err)
	}
	proposal.GroupPolicyAddress = groupPolicyAddress

	if p.Msg == nil {
		return proposal, metadata, nil
	}

	// set messages field
	result, err := govcli.Prompt(p.Msg, "msg")
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

// NewCmdDraftProposal let a user generate a draft proposal.
func NewCmdDraftProposal() *cobra.Command {
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
			proposalTypesPrompt := promptui.Select{
				Label: "Select proposal type",
				Items: []string{proposalText, proposalOther},
			}

			_, selectedProposalType, err := proposalTypesPrompt.Run()
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

				proposal.Msg, err = sdk.GetMsgFromTypeURL(clientCtx.Codec, result)
				if err != nil {
					// should never happen
					panic(err)
				}
			default:
				panic("unexpected proposal type")
			}

			result, metadata, err := proposal.Prompt(clientCtx.Codec)
			if err != nil {
				return err
			}

			if err := writeFile(draftProposalFileName, result); err != nil {
				return err
			}

			if err := writeFile(draftMetadataFileName, metadata); err != nil {
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
