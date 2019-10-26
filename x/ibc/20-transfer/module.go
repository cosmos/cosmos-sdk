package transfer

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/client/cli"
)

// Name returns the IBC transfer ICS name
func Name() string {
	return SubModuleName
}

// GetTxCmd returns the root tx command for the IBC transfer.
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}
