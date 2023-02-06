package cli

import (
	"cosmossdk.io/x/upgrade/types"
	"github.com/spf13/pflag"
)

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
