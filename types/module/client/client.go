package client

import (
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

type ClientModule interface {
	// cli
	NewTxCmd(ctx tx.TxCmdContext) *cobra.Command
}
