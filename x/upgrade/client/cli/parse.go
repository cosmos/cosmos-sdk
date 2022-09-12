package cli

import (
	"github.com/spf13/pflag"
)

func parseArgsToContent(fs *pflag.FlagSet, name string) (gov.Content, error) {
	title, err := fs.GetString(cli.FlagTitle) //nolint:staticcheck
	if err != nil {
		return nil, err
	}

	description, err := fs.GetString(cli.FlagDescription) //nolint:staticcheck
	if err != nil {
		return nil, err
	}

func parsePlan(fs *pflag.FlagSet, name string) (types.Plan, error) {
	height, err := fs.GetInt64(FlagUpgradeHeight)
	if err != nil {
		return types.Plan{}, err
	}

	info, err := fs.GetString(FlagUpgradeInfo)
	if err != nil {
		return types.Plan{}, err
	}

	return types.Plan{Name: name, Height: height, Info: info}, nil
}
