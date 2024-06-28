package cometbft

import "github.com/spf13/cobra"

const (
	FlagQuery   = "query"
	FlagType    = "type"
	FlagOrderBy = "order_by"
)

const (
	FlagWithComet     = "with-comet"
	FlagAddress       = "address"
	FlagTransport     = "transport"
	FlagTraceStore    = "trace-store"
	FlagCPUProfile    = "cpu-profile"
	FlagMinGasPrices  = "minimum-gas-prices"
	FlagQueryGasLimit = "query-gas-limit"
	FlagHaltHeight    = "halt-height"
	FlagHaltTime      = "halt-time"
	FlagTrace         = "trace"
)

const (
	FlagChainID      = "chain-id"
	FlagNode         = "node"
	FlagGRPC         = "grpc-addr"
	FlagGRPCInsecure = "grpc-insecure"
	FlagHeight       = "height"
	FlagPage         = "page"
	FlagLimit        = "limit"
	FlagOutput       = "output"
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
	cmd.Flags().StringP(FlagOutput, "o", OutputFormatText, "Output format (text|json)")

	// some base commands does not require chainID e.g `simd testnet` while subcommands do
	// hence the flag should not be required for those commands
	_ = cmd.MarkFlagRequired(FlagChainID)
}
