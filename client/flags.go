package client

import "github.com/spf13/cobra"

// nolint
const (
	// DefaultGasAdjustment is applied to gas estimates to avoid tx
	// execution failures due to state changes that might
	// occur between the tx simulation and the actual run.
	DefaultGasAdjustment = 1.0
	DefaultGasLimit      = 200000

	FlagUseLedger     = "ledger"
	FlagChainID       = "chain-id"
	FlagNode          = "node"
	FlagHeight        = "height"
	FlagGas           = "gas"
	FlagGasAdjustment = "gas-adjustment"
	FlagTrustNode     = "trust-node"
	FlagFrom          = "from"
	FlagName          = "name"
	FlagAccountNumber = "account-number"
	FlagSequence      = "sequence"
	FlagMemo          = "memo"
	FlagFee           = "fee"
	FlagAsync         = "async"
	FlagJson          = "json"
	FlagPrintResponse = "print-response"
	FlagDryRun        = "dry-run"
	FlagGenerateOnly  = "generate-only"
)

// LineBreak can be included in a command list to provide a blank line
// to help with readability
var LineBreak = &cobra.Command{Run: func(*cobra.Command, []string) {}}

// GetCommands adds common flags to query commands
func GetCommands(cmds ...*cobra.Command) []*cobra.Command {
	for _, c := range cmds {
		// TODO: make this default false when we support proofs
		c.Flags().Bool(FlagTrustNode, true, "Don't verify proofs for responses")
		c.Flags().Bool(FlagUseLedger, false, "Use a connected Ledger device")
		c.Flags().String(FlagChainID, "", "Chain ID of tendermint node")
		c.Flags().String(FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")
		c.Flags().Int64(FlagHeight, 0, "block height to query, omit to get most recent provable block")
	}
	return cmds
}

// PostCommands adds common flags for commands to post tx
func PostCommands(cmds ...*cobra.Command) []*cobra.Command {
	for _, c := range cmds {
		c.Flags().String(FlagFrom, "", "Name of private key with which to sign")
		c.Flags().Int64(FlagAccountNumber, 0, "AccountNumber number to sign the tx")
		c.Flags().Int64(FlagSequence, 0, "Sequence number to sign the tx")
		c.Flags().String(FlagMemo, "", "Memo to send along with transaction")
		c.Flags().String(FlagFee, "", "Fee to pay along with transaction")
		c.Flags().String(FlagChainID, "", "Chain ID of tendermint node")
		c.Flags().String(FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")
		c.Flags().Bool(FlagUseLedger, false, "Use a connected Ledger device")
		c.Flags().Int64(FlagGas, DefaultGasLimit, "gas limit to set per-transaction; set to 0 to calculate required gas automatically")
		c.Flags().Float64(FlagGasAdjustment, DefaultGasAdjustment, "adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored ")
		c.Flags().Bool(FlagAsync, false, "broadcast transactions asynchronously")
		c.Flags().Bool(FlagJson, false, "return output in json format")
		c.Flags().Bool(FlagPrintResponse, true, "return tx response (only works with async = false)")
		c.Flags().Bool(FlagTrustNode, true, "Don't verify proofs for query responses")
		c.Flags().Bool(FlagDryRun, false, "ignore the --gas flag and perform a simulation of a transaction, but don't broadcast it")
		c.Flags().Bool(FlagGenerateOnly, false, "build an unsigned transaction and write it to STDOUT")
	}
	return cmds
}
