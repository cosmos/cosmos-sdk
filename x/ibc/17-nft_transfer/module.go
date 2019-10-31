package nft_transfer

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/17-nft_transfer/client/cli"
)

// Name returns the IBC nft_transfer ICS name
func Name() string {
	return SubModuleName
}

// GetTxCmd returns the root tx command for the IBC nft_transfer.
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}
