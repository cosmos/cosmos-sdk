package flags

import "github.com/spf13/cobra"

const (
	FlagQuery   = "query"
	FlagType    = "type"
	FlagOrderBy = "order_by"
)

const (
	FlagChainID          = "chain-id"
	FlagNode             = "node"
	FlagGRPC             = "grpc-addr"
	FlagGRPCInsecure     = "grpc-insecure"
	FlagHeight           = "height"
	FlagGasAdjustment    = "gas-adjustment"
	FlagFrom             = "from"
	FlagName             = "name"
	FlagAccountNumber    = "account-number"
	FlagSequence         = "sequence"
	FlagNote             = "note"
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
	FlagUnordered        = "unordered"
	FlagKeyAlgorithm     = "algo"
	FlagKeyType          = "key-type"
	FlagFeePayer         = "fee-payer"
	FlagFeeGranter       = "fee-granter"
	FlagReverse          = "reverse"
	FlagTip              = "tip"
	FlagAux              = "aux"
	FlagInitHeight       = "initial-height"
	// FlagOutput is the flag to set the output format.
	// This differs from FlagOutputDocument that is used to set the output file.
	FlagOutput = "output"
	// Logging flags
	FlagLogLevel   = "log_level"
	FlagLogFormat  = "log_format"
	FlagLogNoColor = "log_no_color"
)

// List of supported output formats
const (
	OutputFormatJSON = "json"
	OutputFormatText = "text"
)

// AddQueryFlagsToCmd adds common flags to a module query command.
func AddQueryFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().String(FlagNode, "tcp://localhost:26657", "<host>:<port> to CometBFT RPC interface for this chain")
	cmd.Flags().String(FlagGRPC, "", "the gRPC endpoint to use for this chain")
	cmd.Flags().Bool(FlagGRPCInsecure, false, "allow gRPC over insecure channels, if not the server must use TLS")
	cmd.Flags().Int64(FlagHeight, 0, "Use a specific height to query state at (this can error if the node is pruning state)")
	cmd.Flags().StringP(FlagOutput, "o", "text", "Output format (text|json)")

	// some base commands does not require chainID e.g `simd testnet` while subcommands do
	// hence the flag should not be required for those commands
	_ = cmd.MarkFlagRequired(FlagChainID)
}
