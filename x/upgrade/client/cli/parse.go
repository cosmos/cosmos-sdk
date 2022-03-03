package cli

import (
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/spf13/cobra"
)

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
