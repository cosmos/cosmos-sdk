package module

import (
	"github.com/spf13/cobra"
)

type ClientModule interface {
	// cli
	NewTxCmd(TxCmdContext) *cobra.Command
}
