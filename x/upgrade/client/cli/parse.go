package cli

import (
	"encoding/json"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/spf13/cobra"
)

// softwareUpgradeProposal defines a software upgrade proposal.
type softwareUpgradeProposal struct {
	Plan     json.RawMessage
	Metadata []byte
	Deposit  string
}

func parseSubmitSoftwareUpgradeProposal(cdc codec.Codec, path string) (*types.Plan, []byte, sdk.Coins, error) {
	var proposal softwareUpgradeProposal
	var plan types.Plan

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, err
	}

	err = json.Unmarshal(contents, &proposal)
	if err != nil {
		return nil, nil, nil, err
	}

	err = cdc.UnmarshalJSON(proposal.Plan, &plan)
	if err != nil {
		return nil, nil, nil, err
	}

	deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
	if err != nil {
		return nil, nil, nil, err
	}

	return &plan, proposal.Metadata, deposit, nil
}

func parseArgsToContent(cmd *cobra.Command, name string) (gov.Content, error) {
	title, err := cmd.Flags().GetString(cli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(cli.FlagDescription)
	if err != nil {
		return nil, err
	}

	height, err := cmd.Flags().GetInt64(FlagUpgradeHeight)
	if err != nil {
		return nil, err
	}

	info, err := cmd.Flags().GetString(FlagUpgradeInfo)
	if err != nil {
		return nil, err
	}

	plan := types.Plan{Name: name, Height: height, Info: info}
	content := types.NewSoftwareUpgradeProposal(title, description, plan)
	return content, nil
}
