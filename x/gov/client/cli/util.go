package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type legacyProposal struct {
	Title       string
	Description string
	Type        string
	Deposit     string
}

func (p legacyProposal) validate() error {
	if p.Type == "" {
		return fmt.Errorf("proposal type is required")
	}

	if p.Title == "" {
		return fmt.Errorf("proposal title is required")
	}

	if p.Description == "" {
		return fmt.Errorf("proposal description is required")
	}
	return nil
}

func parseSubmitLegacyProposalFlags(fs *pflag.FlagSet) (*legacyProposal, error) {
	proposal := &legacyProposal{}
	proposalFile, _ := fs.GetString(FlagProposal)

	if proposalFile == "" {
		proposalType, _ := fs.GetString(FlagProposalType)
		proposal.Title, _ = fs.GetString(FlagTitle)
		proposal.Description, _ = fs.GetString(FlagDescription)
		proposal.Type = govutils.NormalizeProposalType(proposalType)
		proposal.Deposit, _ = fs.GetString(FlagDeposit)
		if err := proposal.validate(); err != nil {
			return nil, err
		}

		return proposal, nil
	}

	for _, flag := range ProposalFlags {
		if v, _ := fs.GetString(flag); v != "" {
			return nil, fmt.Errorf("--%s flag provided alongside --proposal, which is a noop", flag)
		}
	}

	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, proposal)
	if err != nil {
		return nil, err
	}

	if err := proposal.validate(); err != nil {
		return nil, err
	}

	return proposal, nil
}

// proposal defines the new Msg-based proposal.
type proposal struct {
	// Msgs defines an array of sdk.Msgs proto-JSON-encoded as Anys.
	Messages []json.RawMessage `json:"messages,omitempty"`
	Metadata string            `json:"metadata"`
	Deposit  string            `json:"deposit"`
}

func parseSubmitProposal(cdc codec.Codec, path string) ([]sdk.Msg, string, sdk.Coins, error) {
	var proposal proposal

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, "", nil, err
	}

	err = json.Unmarshal(contents, &proposal)
	if err != nil {
		return nil, "", nil, err
	}

	msgs := make([]sdk.Msg, len(proposal.Messages))
	for i, anyJSON := range proposal.Messages {
		var msg sdk.Msg
		err := cdc.UnmarshalInterfaceJSON(anyJSON, &msg)
		if err != nil {
			return nil, "", nil, err
		}

		msgs[i] = msg
	}

	deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
	if err != nil {
		return nil, "", nil, err
	}

	return msgs, proposal.Metadata, deposit, nil
}

// AddGovPropFlagsToCmd adds flags for defining MsgSubmitProposal fields.
func AddGovPropFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().String(FlagDeposit, "", "The deposit to include with the governance proposal")
	cmd.Flags().String(FlagMetadata, "", "The metadata to include with the governance proposal")
}

// ReadGovPropFlags parses a MsgSubmitProposal from the provided context and flags.
// Setting the messages is up to the caller.
func ReadGovPropFlags(clientCtx client.Context, flagSet *pflag.FlagSet) (*govv1.MsgSubmitProposal, error) {
	rv := &govv1.MsgSubmitProposal{}

	deposit, err := flagSet.GetString(FlagDeposit)
	if err != nil {
		return nil, fmt.Errorf("could not read deposit: %w", err)
	}
	if len(deposit) > 0 {
		rv.InitialDeposit, err = sdk.ParseCoinsNormalized(deposit)
		if err != nil {
			return nil, fmt.Errorf("invalid deposit: %w", err)
		}
	}

	rv.Metadata, err = flagSet.GetString(FlagMetadata)
	if err != nil {
		return nil, fmt.Errorf("could not read metadata: %w", err)
	}

	rv.Proposer = clientCtx.GetFromAddress().String()

	return rv, nil
}

// GenerateOrBroadcastTxCLIAsGovProp wraps the provided msgs in a governance proposal
// and calls GenerateOrBroadcastTxCLI for that proposal.
// At least one msg is required.
// This uses flags added by AddGovPropFlagsToCmd to fill in the rest of the proposal.
func GenerateOrBroadcastTxCLIAsGovProp(clientCtx client.Context, flagSet *pflag.FlagSet, msgs ...sdk.Msg) error {
	if len(msgs) == 0 {
		return fmt.Errorf("no messages to submit")
	}

	prop, err := ReadGovPropFlags(clientCtx, flagSet)
	if err != nil {
		return err
	}

	prop.Messages = make([]*codectypes.Any, len(msgs))
	for i, msg := range msgs {
		prop.Messages[i], err = codectypes.NewAnyWithValue(msg)
		if err != nil {
			if len(msgs) == 1 {
				return fmt.Errorf("could not wrap %T message as Any: %w", msg, err)
			}
			return fmt.Errorf("could not wrap message %d (%T) as Any: %w", i, msg, err)
		}
	}

	return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, prop)
}
