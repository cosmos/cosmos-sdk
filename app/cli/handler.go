package cli

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/spf13/cobra"
)

type Handler struct {
	app.BasicGenesisHandler
	TxCommand    *cobra.Command
	QueryCommand *cobra.Command
}
