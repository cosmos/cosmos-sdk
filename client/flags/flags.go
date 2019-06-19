package flags

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tmcli "github.com/tendermint/tendermint/libs/cli"
)

// nolint
const (
	// DefaultGasAdjustment is applied to gas estimates to avoid tx execution
	// failures due to state changes that might occur between the tx simulation
	// and the actual run.
	DefaultGasAdjustment = 1.0
	DefaultGasLimit      = 200000
	GasFlagAuto          = "auto"

	// BroadcastBlock defines a tx broadcasting mode where the client waits for
	// the tx to be committed in a block.
	BroadcastBlock = "block"
	// BroadcastSync defines a tx broadcasting mode where the client waits for
	// a CheckTx execution response only.
	BroadcastSync = "sync"
	// BroadcastAsync defines a tx broadcasting mode where the client returns
	// immediately.
	BroadcastAsync = "async"

	FlagHome               = tmcli.HomeFlag
	FlagUseLedger          = "ledger"
	FlagChainID            = "chain-id"
	FlagNode               = "node"
	FlagHeight             = "height"
	FlagGasAdjustment      = "gas-adjustment"
	FlagTrustNode          = "trust-node"
	FlagFrom               = "from"
	FlagName               = "name"
	FlagAccountNumber      = "account-number"
	FlagSequence           = "sequence"
	FlagMemo               = "memo"
	FlagFees               = "fees"
	FlagGasPrices          = "gas-prices"
	FlagBroadcastMode      = "broadcast-mode"
	FlagDryRun             = "dry-run"
	FlagGenerateOnly       = "generate-only"
	FlagIndentResponse     = "indent"
	FlagListenAddr         = "laddr"
	FlagMaxOpenConnections = "max-open"
	FlagRPCReadTimeout     = "read-timeout"
	FlagRPCWriteTimeout    = "write-timeout"
	FlagOutputDocument     = "output-document" // inspired by wget -O
	FlagSkipConfirmation   = "yes"
)

// LineBreak can be included in a command list to provide a blank line
// to help with readability
var (
	LineBreak  = &cobra.Command{Run: func(*cobra.Command, []string) {}}
	GasFlagVar = GasSetting{Gas: DefaultGasLimit}
)

// GetCommands adds common flags to query commands
func GetCommands(cmds ...*cobra.Command) []*cobra.Command {
	for _, c := range cmds {
		c.Flags().Bool(FlagIndentResponse, false, "Add indent to JSON response")
		c.Flags().Bool(FlagTrustNode, false, "Trust connected full node (don't verify proofs for responses)")
		c.Flags().Bool(FlagUseLedger, false, "Use a connected Ledger device")
		c.Flags().String(FlagNode, "tcp://localhost:26657", "<host>:<port> to Tendermint RPC interface for this chain")
		c.Flags().Int64(FlagHeight, 0, "Use a specific height to query state at (this can error if the node is pruning state)")

		viper.BindPFlag(FlagTrustNode, c.Flags().Lookup(FlagTrustNode))
		viper.BindPFlag(FlagUseLedger, c.Flags().Lookup(FlagUseLedger))
		viper.BindPFlag(FlagNode, c.Flags().Lookup(FlagNode))

		c.MarkFlagRequired(FlagChainID)
	}
	return cmds
}

// PostCommands adds common flags for commands to post tx
func PostCommands(cmds ...*cobra.Command) []*cobra.Command {
	for _, c := range cmds {
		c.Flags().Bool(FlagIndentResponse, false, "Add indent to JSON response")
		c.Flags().String(FlagFrom, "", "Name or address of private key with which to sign")
		c.Flags().Uint64P(FlagAccountNumber, "a", 0, "The account number of the signing account (offline mode only)")
		c.Flags().Uint64P(FlagSequence, "s", 0, "The sequence number of the signing account (offline mode only)")
		c.Flags().String(FlagMemo, "", "Memo to send along with transaction")
		c.Flags().String(FlagFees, "", "Fees to pay along with transaction; eg: 10uatom")
		c.Flags().String(FlagGasPrices, "", "Gas prices to determine the transaction fee (e.g. 10uatom)")
		c.Flags().String(FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")
		c.Flags().Bool(FlagUseLedger, false, "Use a connected Ledger device")
		c.Flags().Float64(FlagGasAdjustment, DefaultGasAdjustment, "adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored ")
		c.Flags().StringP(FlagBroadcastMode, "b", BroadcastSync, "Transaction broadcasting mode (sync|async|block)")
		c.Flags().Bool(FlagTrustNode, true, "Trust connected full node (don't verify proofs for responses)")
		c.Flags().Bool(FlagDryRun, false, "ignore the --gas flag and perform a simulation of a transaction, but don't broadcast it")
		c.Flags().Bool(FlagGenerateOnly, false, "Build an unsigned transaction and write it to STDOUT (when enabled, the local Keybase is not accessible and the node operates offline)")
		c.Flags().BoolP(FlagSkipConfirmation, "y", false, "Skip tx broadcasting prompt confirmation")

		// --gas can accept integers and "simulate"
		c.Flags().Var(&GasFlagVar, "gas", fmt.Sprintf(
			"gas limit to set per-transaction; set to %q to calculate required gas automatically (default %d)",
			GasFlagAuto, DefaultGasLimit,
		))

		viper.BindPFlag(FlagTrustNode, c.Flags().Lookup(FlagTrustNode))
		viper.BindPFlag(FlagUseLedger, c.Flags().Lookup(FlagUseLedger))
		viper.BindPFlag(FlagNode, c.Flags().Lookup(FlagNode))

		c.MarkFlagRequired(FlagChainID)
	}
	return cmds
}

// RegisterRestServerFlags registers the flags required for rest server
func RegisterRestServerFlags(cmd *cobra.Command) *cobra.Command {
	cmd = GetCommands(cmd)[0]
	cmd.Flags().String(FlagListenAddr, "tcp://localhost:1317", "The address for the server to listen on")
	cmd.Flags().Uint(FlagMaxOpenConnections, 1000, "The number of maximum open connections")
	cmd.Flags().Uint(FlagRPCReadTimeout, 10, "The RPC read timeout (in seconds)")
	cmd.Flags().Uint(FlagRPCWriteTimeout, 10, "The RPC write timeout (in seconds)")

	return cmd
}

// Gas flag parsing functions

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

// NewCompletionCmd builds a cobra.Command that generate bash completion
// scripts for the given root command. If hidden is true, the command
// will not show up in the root command's list of available commands.
func NewCompletionCmd(rootCmd *cobra.Command, hidden bool) *cobra.Command {
	flagZsh := "zsh"
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate Bash/Zsh completion script to STDOUT",
		Long: `To load completion script run

. <(completion_script)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(completion_script)
`,
		RunE: func(_ *cobra.Command, _ []string) error {
			if viper.GetBool(flagZsh) {
				return rootCmd.GenZshCompletion(os.Stdout)
			}
			return rootCmd.GenBashCompletion(os.Stdout)
		},
		Hidden: hidden,
		Args:   cobra.NoArgs,
	}

	cmd.Flags().Bool(flagZsh, false, "Generate Zsh completion script")

	return cmd
}
