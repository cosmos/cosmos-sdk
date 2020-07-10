package flags

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

const (
	// DefaultGasAdjustment is applied to gas estimates to avoid tx execution
	// failures due to state changes that might occur between the tx simulation
	// and the actual run.
	DefaultGasAdjustment = 1.0
	DefaultGasLimit      = 200000
	GasFlagAuto          = "auto"

	// DefaultKeyringBackend
	DefaultKeyringBackend = keyring.BackendOS
)

const (
	// BroadcastBlock defines a tx broadcasting mode where the client waits for
	// the tx to be committed in a block.
	BroadcastBlock = "block"
	// BroadcastSync defines a tx broadcasting mode where the client waits for
	// a CheckTx execution response only.
	BroadcastSync = "sync"
	// BroadcastAsync defines a tx broadcasting mode where the client returns
	// immediately.
	BroadcastAsync = "async"
)

// List of CLI flags
const (
	FlagHome             = tmcli.HomeFlag
	FlagUseLedger        = "ledger"
	FlagChainID          = "chain-id"
	FlagNode             = "node"
	FlagHeight           = "height"
	FlagGasAdjustment    = "gas-adjustment"
	FlagTrustNode        = "trust-node"
	FlagFrom             = "from"
	FlagName             = "name"
	FlagAccountNumber    = "account-number"
	FlagSequence         = "sequence"
	FlagMemo             = "memo"
	FlagFees             = "fees"
	FlagGasPrices        = "gas-prices"
	FlagBroadcastMode    = "broadcast-mode"
	FlagDryRun           = "dry-run"
	FlagGenerateOnly     = "generate-only"
	FlagOffline          = "offline"
	FlagOutputDocument   = "output-document" // inspired by wget -O
	FlagSkipConfirmation = "yes"
	FlagProve            = "prove"
	FlagKeyringBackend   = "keyring-backend"
	FlagPage             = "page"
	FlagLimit            = "limit"
	FlagSignMode         = "sign-mode"
)

// LineBreak can be included in a command list to provide a blank line
// to help with readability
var (
	LineBreak  = &cobra.Command{Run: func(*cobra.Command, []string) {}}
	GasFlagVar = GasSetting{Gas: DefaultGasLimit}
)

// AddQueryFlagsToCmd adds common flags to a module query command.
func AddQueryFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
	cmd.Flags().Bool(FlagUseLedger, false, "Use a connected Ledger device")
	cmd.Flags().String(FlagNode, "tcp://localhost:26657", "<host>:<port> to Tendermint RPC interface for this chain")
	cmd.Flags().Int64(FlagHeight, 0, "Use a specific height to query state at (this can error if the node is pruning state)")
	cmd.Flags().String(FlagKeyringBackend, DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().StringP(tmcli.OutputFlag, "o", "text", "Output format (text|json)")

	cmd.MarkFlagRequired(FlagChainID)

	cmd.SetErr(cmd.ErrOrStderr())
	cmd.SetOut(cmd.OutOrStdout())

	// TODO: REMOVE VIPER CALLS!
	viper.BindPFlag(FlagTrustNode, cmd.Flags().Lookup(FlagTrustNode))
	viper.BindPFlag(FlagUseLedger, cmd.Flags().Lookup(FlagUseLedger))
	viper.BindPFlag(FlagNode, cmd.Flags().Lookup(FlagNode))
	viper.BindPFlag(FlagKeyringBackend, cmd.Flags().Lookup(FlagKeyringBackend))
}

// AddTxFlagsToCmd adds common flags to a module tx command.
func AddTxFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().String(FlagFrom, "", "Name or address of private key with which to sign")
	cmd.Flags().Uint64P(FlagAccountNumber, "a", 0, "The account number of the signing account (offline mode only)")
	cmd.Flags().Uint64P(FlagSequence, "s", 0, "The sequence number of the signing account (offline mode only)")
	cmd.Flags().String(FlagMemo, "", "Memo to send along with transaction")
	cmd.Flags().String(FlagFees, "", "Fees to pay along with transaction; eg: 10uatom")
	cmd.Flags().String(FlagGasPrices, "", "Gas prices to determine the transaction fee (e.g. 10uatom)")
	cmd.Flags().String(FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")
	cmd.Flags().Bool(FlagUseLedger, false, "Use a connected Ledger device")
	cmd.Flags().Float64(FlagGasAdjustment, DefaultGasAdjustment, "adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored ")
	cmd.Flags().StringP(FlagBroadcastMode, "b", BroadcastSync, "Transaction broadcasting mode (sync|async|block)")
	cmd.Flags().Bool(FlagTrustNode, true, "Trust connected full node (don't verify proofs for responses)")
	cmd.Flags().Bool(FlagDryRun, false, "ignore the --gas flag and perform a simulation of a transaction, but don't broadcast it")
	cmd.Flags().Bool(FlagGenerateOnly, false, "Build an unsigned transaction and write it to STDOUT (when enabled, the local Keybase is not accessible)")
	cmd.Flags().Bool(FlagOffline, false, "Offline mode (does not allow any online functionality")
	cmd.Flags().BoolP(FlagSkipConfirmation, "y", false, "Skip tx broadcasting prompt confirmation")
	cmd.Flags().String(FlagKeyringBackend, DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(FlagSignMode, "", "Choose sign mode (direct|amino-json), this is an advanced feature")

	// --gas can accept integers and "simulate"
	//
	// TODO: Remove usage of var in favor of string as this is technical creating
	// a singleton usage pattern and can cause issues in parallel tests.
	//
	// REF: https://github.com/cosmos/cosmos-sdk/issues/6545
	cmd.Flags().Var(&GasFlagVar, "gas", fmt.Sprintf(
		"gas limit to set per-transaction; set to %q to calculate required gas automatically (default %d)",
		GasFlagAuto, DefaultGasLimit,
	))

	cmd.MarkFlagRequired(FlagChainID)

	cmd.SetErr(cmd.ErrOrStderr())
	cmd.SetOut(cmd.OutOrStdout())

	// TODO: REMOVE VIPER CALLS!
	viper.BindPFlag(FlagTrustNode, cmd.Flags().Lookup(FlagTrustNode))
	viper.BindPFlag(FlagUseLedger, cmd.Flags().Lookup(FlagUseLedger))
	viper.BindPFlag(FlagNode, cmd.Flags().Lookup(FlagNode))
	viper.BindPFlag(FlagKeyringBackend, cmd.Flags().Lookup(FlagKeyringBackend))
}

// GetCommands adds common flags to query commands.
//
// TODO: REMOVE.
func GetCommands(cmds ...*cobra.Command) []*cobra.Command {
	for _, c := range cmds {
		AddQueryFlagsToCmd(c)
	}

	return cmds
}

// PostCommands adds common flags for commands to post tx
//
// TODO: REMOVE.
func PostCommands(cmds ...*cobra.Command) []*cobra.Command {
	for _, c := range cmds {
		AddTxFlagsToCmd(c)
	}
	return cmds
}

// GasSetting encapsulates the possible values passed through the --gas flag.
type GasSetting struct {
	Simulate bool
	Gas      uint64
}

// Type returns the flag's value type.
func (v *GasSetting) Type() string { return "string" }

// Set parses and sets the value of the --gas flag.
func (v *GasSetting) Set(s string) (err error) {
	v.Simulate, v.Gas, err = ParseGas(s)
	return
}

func (v *GasSetting) String() string {
	if v.Simulate {
		return GasFlagAuto
	}
	return strconv.FormatUint(v.Gas, 10)
}

// ParseGas parses the value of the gas option.
func ParseGas(gasStr string) (simulateAndExecute bool, gas uint64, err error) {
	switch gasStr {
	case "":
		gas = DefaultGasLimit
	case GasFlagAuto:
		simulateAndExecute = true
	default:
		gas, err = strconv.ParseUint(gasStr, 10, 64)
		if err != nil {
			err = fmt.Errorf("gas must be either integer or %q", GasFlagAuto)
			return
		}
	}
	return
}
