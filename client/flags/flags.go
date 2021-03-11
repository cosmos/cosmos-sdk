package flags

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
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

	// BroadcastBlock defines a tx broadcasting mode where the client waits for
	// the tx to be committed in a block.
	BroadcastBlock = "block"
	// BroadcastSync defines a tx broadcasting mode where the client waits for
	// a CheckTx execution response only.
	BroadcastSync = "sync"
	// BroadcastAsync defines a tx broadcasting mode where the client returns
	// immediately.
	BroadcastAsync = "async"

	// SignModeDirect is the value of the --sign-mode flag for SIGN_MODE_DIRECT
	SignModeDirect = "direct"
	// SignModeLegacyAminoJSON is the value of the --sign-mode flag for SIGN_MODE_LEGACY_AMINO_JSON
	SignModeLegacyAminoJSON = "amino-json"
)

// List of CLI flags
const (
	FlagHome             = tmcli.HomeFlag
	FlagKeyringDir       = "keyring-dir"
	FlagUseLedger        = "ledger"
	FlagChainID          = "chain-id"
	FlagNode             = "node"
	FlagHeight           = "height"
	FlagGasAdjustment    = "gas-adjustment"
	FlagFrom             = "from"
	FlagName             = "name"
	FlagAccountNumber    = "account-number"
	FlagSequence         = "sequence"
	FlagMemo             = "memo"
	FlagFees             = "fees"
	FlagGas              = "gas"
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
	FlagPageKey          = "page-key"
	FlagOffset           = "offset"
	FlagCountTotal       = "count-total"
	FlagTimeoutHeight    = "timeout-height"
	FlagKeyAlgorithm     = "algo"

	// Tendermint logging flags
	FlagLogLevel  = "log_level"
	FlagLogFormat = "log_format"
)

// LineBreak can be included in a command list to provide a blank line
// to help with readability
var LineBreak = &cobra.Command{Run: func(*cobra.Command, []string) {}}

// AddQueryFlagsToCmd adds common flags to a module query command.
func AddQueryFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().String(FlagNode, "tcp://localhost:26657", "<host>:<port> to Tendermint RPC interface for this chain")
	cmd.Flags().Int64(FlagHeight, 0, "Use a specific height to query state at (this can error if the node is pruning state)")
	cmd.Flags().StringP(tmcli.OutputFlag, "o", "text", "Output format (text|json)")

	cmd.MarkFlagRequired(FlagChainID)

	cmd.SetErr(cmd.ErrOrStderr())
	cmd.SetOut(cmd.OutOrStdout())
}

// AddTxFlagsToCmd adds common flags to a module tx command.
func AddTxFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().String(FlagKeyringDir, "", "The client Keyring directory; if omitted, the default 'home' directory will be used")
	cmd.Flags().String(FlagFrom, "", "Name or address of private key with which to sign")
	cmd.Flags().Uint64P(FlagAccountNumber, "a", 0, "The account number of the signing account (offline mode only)")
	cmd.Flags().Uint64P(FlagSequence, "s", 0, "The sequence number of the signing account (offline mode only)")
	cmd.Flags().String(FlagMemo, "", "Memo to send along with transaction")
	cmd.Flags().String(FlagFees, "", "Fees to pay along with transaction; eg: 10uatom")
	cmd.Flags().String(FlagGasPrices, "", "Gas prices in decimal format to determine the transaction fee (e.g. 0.1uatom)")
	cmd.Flags().String(FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")
	cmd.Flags().Bool(FlagUseLedger, false, "Use a connected Ledger device")
	cmd.Flags().Float64(FlagGasAdjustment, DefaultGasAdjustment, "adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored ")
	cmd.Flags().StringP(FlagBroadcastMode, "b", BroadcastSync, "Transaction broadcasting mode (sync|async|block)")
	cmd.Flags().Bool(FlagDryRun, false, "ignore the --gas flag and perform a simulation of a transaction, but don't broadcast it")
	cmd.Flags().Bool(FlagGenerateOnly, false, "Build an unsigned transaction and write it to STDOUT (when enabled, the local Keybase is not accessible)")
	cmd.Flags().Bool(FlagOffline, false, "Offline mode (does not allow any online functionality")
	cmd.Flags().BoolP(FlagSkipConfirmation, "y", false, "Skip tx broadcasting prompt confirmation")
	cmd.Flags().String(FlagKeyringBackend, DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test|memory)")
	cmd.Flags().String(FlagSignMode, "", "Choose sign mode (direct|amino-json), this is an advanced feature")
	cmd.Flags().Uint64(FlagTimeoutHeight, 0, "Set a block timeout height to prevent the tx from being committed past a certain height")

	// --gas can accept integers and "auto"
	cmd.Flags().String(FlagGas, "", fmt.Sprintf("gas limit to set per-transaction; set to %q to calculate sufficient gas automatically (default %d)", GasFlagAuto, DefaultGasLimit))

	cmd.MarkFlagRequired(FlagChainID)

	cmd.SetErr(cmd.ErrOrStderr())
	cmd.SetOut(cmd.OutOrStdout())
}

// AddPaginationFlagsToCmd adds common pagination flags to cmd
func AddPaginationFlagsToCmd(cmd *cobra.Command, query string) {
	cmd.Flags().Uint64(FlagPage, 1, fmt.Sprintf("pagination page of %s to query for. This sets offset to a multiple of limit", query))
	cmd.Flags().String(FlagPageKey, "", fmt.Sprintf("pagination page-key of %s to query for", query))
	cmd.Flags().Uint64(FlagOffset, 0, fmt.Sprintf("pagination offset of %s to query for", query))
	cmd.Flags().Uint64(FlagLimit, 100, fmt.Sprintf("pagination limit of %s to query for", query))
	cmd.Flags().Bool(FlagCountTotal, false, fmt.Sprintf("count total number of records in %s to query for", query))
}

// GasSetting encapsulates the possible values passed through the --gas flag.
type GasSetting struct {
	Simulate bool
	Gas      uint64
}

func (v *GasSetting) String() string {
	if v.Simulate {
		return GasFlagAuto
	}

	return strconv.FormatUint(v.Gas, 10)
}

// ParseGasSetting parses a string gas value. The value may either be 'auto',
// which indicates a transaction should be executed in simulate mode to
// automatically find a sufficient gas value, or a string integer. It returns an
// error if a string integer is provided which cannot be parsed.
func ParseGasSetting(gasStr string) (GasSetting, error) {
	switch gasStr {
	case "":
		return GasSetting{false, DefaultGasLimit}, nil

	case GasFlagAuto:
		return GasSetting{true, 0}, nil

	default:
		gas, err := strconv.ParseUint(gasStr, 10, 64)
		if err != nil {
			return GasSetting{}, fmt.Errorf("gas must be either integer or %s", GasFlagAuto)
		}

		return GasSetting{false, gas}, nil
	}
}
