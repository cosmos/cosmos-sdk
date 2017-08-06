package commands

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/client/commands"
	"github.com/tendermint/basecoin/client/commands/txs"
	"github.com/tendermint/basecoin/modules/etc"
)

// SetTxCmd is CLI command to set data
var SetTxCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets a key value pair",
	RunE:  commands.RequireInit(setTxCmd),
}

// RemoveTxCmd is CLI command to remove data
var RemoveTxCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes a key value pair",
	RunE:  commands.RequireInit(removeTxCmd),
}

//nolint
const (
	FlagKey   = "key"
	FlagValue = "value"
)

func init() {
	SetTxCmd.Flags().String(FlagKey, "", "Key to store data under (hex)")
	SetTxCmd.Flags().String(FlagValue, "", "Data to store (hex)")

	RemoveTxCmd.Flags().String(FlagKey, "", "Key under which to remove data (hex)")
}

// setTxCmd creates a SetTx, wraps, signs, and delivers it
func setTxCmd(cmd *cobra.Command, args []string) error {
	key, err := commands.ParseHexFlag(FlagKey)
	if err != nil {
		return err
	}
	value, err := commands.ParseHexFlag(FlagValue)
	if err != nil {
		return err
	}

	tx := etc.NewSetTx(key, value)
	return txs.DoTx(tx)
}

// removeTxCmd creates a RemoveTx, wraps, signs, and delivers it
func removeTxCmd(cmd *cobra.Command, args []string) error {
	key, err := commands.ParseHexFlag(FlagKey)
	if err != nil {
		return err
	}

	tx := etc.NewRemoveTx(key)
	return txs.DoTx(tx)
}
