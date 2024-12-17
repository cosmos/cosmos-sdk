package governance

import (
	"fmt"

	gogoproto "github.com/cosmos/gogoproto/proto"
	gogoprotoany "github.com/cosmos/gogoproto/types/any"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/types/known/anypb"

	govv1 "cosmossdk.io/api/cosmos/gov/v1"
	"cosmossdk.io/client/v2/internal/coins"
)

const (
	// ModuleName is the name of the governance module name.
	// It should match the module name of the cosmossdk.io/x/gov module.
	ModuleName = "gov"

	FlagDeposit   = "deposit"
	FlagMetadata  = "metadata"
	FlagTitle     = "title"
	FlagSummary   = "summary"
	FlagExpedited = "expedited"
)

// AddGovPropFlagsToCmd adds governance proposal flags to the provided command.
func AddGovPropFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().String(FlagDeposit, "", "The deposit to include with the governance proposal")
	cmd.Flags().String(FlagMetadata, "", "The metadata to include with the governance proposal")
	cmd.Flags().String(FlagTitle, "", "The title to put on the governance proposal")
	cmd.Flags().String(FlagSummary, "", "The summary to include with the governance proposal")
	cmd.Flags().Bool(FlagExpedited, false, "Whether to expedite the governance proposal")
}

// ReadGovPropCmdFlags parses a MsgSubmitProposal from the provided context and flags.
func ReadGovPropCmdFlags(proposer string, flagSet *pflag.FlagSet) (*govv1.MsgSubmitProposal, error) {
	rv := &govv1.MsgSubmitProposal{}

	deposit, err := flagSet.GetString(FlagDeposit)
	if err != nil {
		return nil, fmt.Errorf("could not read deposit: %w", err)
	}
	if len(deposit) > 0 {
		rv.InitialDeposit, err = coins.ParseCoinsNormalized(deposit)
		if err != nil {
			return nil, fmt.Errorf("invalid deposit: %w", err)
		}
	}

	rv.Metadata, err = flagSet.GetString(FlagMetadata)
	if err != nil {
		return nil, fmt.Errorf("could not read metadata: %w", err)
	}

	rv.Title, err = flagSet.GetString(FlagTitle)
	if err != nil {
		return nil, fmt.Errorf("could not read title: %w", err)
	}

	rv.Summary, err = flagSet.GetString(FlagSummary)
	if err != nil {
		return nil, fmt.Errorf("could not read summary: %w", err)
	}

	expedited, err := flagSet.GetBool(FlagExpedited)
	if err != nil {
		return nil, fmt.Errorf("could not read expedited: %w", err)
	}
	if expedited {
		rv.Expedited = true //nolint:staticcheck // We set it in case the message is made for an earlier version of the SDK
		rv.ProposalType = govv1.ProposalType_PROPOSAL_TYPE_EXPEDITED
	}

	rv.Proposer = proposer

	return rv, nil
}

func SetGovMsgs(proposal *govv1.MsgSubmitProposal, msgs ...gogoproto.Message) error {
	if len(msgs) == 0 {
		return fmt.Errorf("zero messages is not supported")
	}

	for _, msg := range msgs {
		anyMsg, err := gogoprotoany.NewAnyWithCacheWithValue(msg)
		if err != nil {
			return err
		}

		proposal.Messages = append(proposal.Messages, &anypb.Any{
			TypeUrl: anyMsg.TypeUrl,
			Value:   anyMsg.Value,
		})
	}

	return nil
}
