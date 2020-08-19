package localhost

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/client/cli"
	"github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
)

// Name returns the IBC client name
func Name() string {
	return types.SubModuleName
}

// GetTxCmd returns the root tx command for the IBC localhost client
func GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}
