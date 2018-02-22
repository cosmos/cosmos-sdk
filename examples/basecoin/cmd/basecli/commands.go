package main

import (
	"github.com/spf13/cobra"
)

const (
	flagTo     = "to"
	flagAmount = "amount"
	flagFee    = "fee"
)

func postSendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Create and sign a send tx",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().String(flagTo, "", "Address to send coins")
	cmd.Flags().String(flagAmount, "", "Amount of coins to send")
	cmd.Flags().String(flagFee, "", "Fee to pay along with transaction")
	return cmd
}
